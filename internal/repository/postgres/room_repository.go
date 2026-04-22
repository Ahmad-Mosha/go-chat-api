package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Ahmad-Mosha/go-chat-api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compile-time check: RoomRepository must implement domain.RoomRepository
var _ domain.RoomRepository = (*RoomRepository)(nil)

type RoomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(room *domain.Room) error {
	query := `INSERT INTO rooms (id, name, is_group, created_at, last_message_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(context.Background(), query, room.ID, room.Name, room.IsGroup, room.CreatedAt, room.LastMessageAt)
	if err != nil {
		return fmt.Errorf("repository: failed to create room: %w", err)
	}
	return nil
}

// CreateWithMembers creates a room and adds all members in a single transaction.
// If any step fails, the entire operation is rolled back.
func (r *RoomRepository) CreateWithMembers(room *domain.Room, memberIDs []string) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("repository: failed to begin transaction: %w", err)
	}
	// Rollback is safe to call even after commit — it becomes a no-op
	defer tx.Rollback(context.Background())

	// Insert the room
	roomQuery := `INSERT INTO rooms (id, name, is_group, created_at, last_message_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(context.Background(), roomQuery, room.ID, room.Name, room.IsGroup, room.CreatedAt, room.LastMessageAt)
	if err != nil {
		return fmt.Errorf("repository: failed to create room: %w", err)
	}

	// Insert all members
	memberQuery := `INSERT INTO room_members (room_id, user_id, joined_at, last_read_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	for _, memberID := range memberIDs {
		_, err = tx.Exec(context.Background(), memberQuery, room.ID, memberID)
		if err != nil {
			return fmt.Errorf("repository: failed to add member %s: %w", memberID, err)
		}
	}

	return tx.Commit(context.Background())
}

func (r *RoomRepository) GetByID(id string) (*domain.Room, error) {
	query := `SELECT id, name, is_group, created_at, last_message_at FROM rooms WHERE id = $1`
	room := &domain.Room{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(&room.ID, &room.Name, &room.IsGroup, &room.CreatedAt, &room.LastMessageAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository: room not found")
		}
		return nil, fmt.Errorf("repository: failed to get room: %w", err)
	}
	return room, nil
}

func (r *RoomRepository) FindOneToOneRoom(userA, userB string) (*domain.Room, error) {
	// This query finds a room that is NOT a group AND contains both userA and userB
	query := `
		SELECT r.id, r.name, r.is_group, r.created_at, r.last_message_at 
		FROM rooms r
		JOIN room_members rm1 ON r.id = rm1.room_id
		JOIN room_members rm2 ON r.id = rm2.room_id
		WHERE r.is_group = false 
		AND rm1.user_id = $1 
		AND rm2.user_id = $2`

	room := &domain.Room{}
	err := r.db.QueryRow(context.Background(), query, userA, userB).Scan(&room.ID, &room.Name, &room.IsGroup, &room.CreatedAt, &room.LastMessageAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil, nil to indicate "No room exists yet" without it being a "system error"
		}
		return nil, fmt.Errorf("repository: failed to find 1:1 room: %w", err)
	}
	return room, nil
}

func (r *RoomRepository) AddMember(roomID, userID string) error {
	query := `INSERT INTO room_members (room_id, user_id, joined_at, last_read_at) VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := r.db.Exec(context.Background(), query, roomID, userID)
	if err != nil {
		return fmt.Errorf("repository: failed to add member: %w", err)
	}
	return nil
}

func (r *RoomRepository) RemoveMember(roomID, userID string) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	_, err := r.db.Exec(context.Background(), query, roomID, userID)
	if err != nil {
		return fmt.Errorf("repository: failed to remove member: %w", err)
	}
	return nil
}

func (r *RoomRepository) GetRoomsByUser(userID string) ([]*domain.Room, error) {
	query := `
		SELECT r.id, r.name, r.is_group, r.created_at, r.last_message_at 
		FROM rooms r
		JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1
		ORDER BY r.last_message_at DESC`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get rooms for user: %w", err)
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		room := &domain.Room{}
		if err := rows.Scan(&room.ID, &room.Name, &room.IsGroup, &room.CreatedAt, &room.LastMessageAt); err != nil {
			return nil, fmt.Errorf("repository: failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: error iterating rooms: %w", err)
	}
	return rooms, nil
}

// IsMember checks if a user is a member of a specific room.
func (r *RoomRepository) IsMember(roomID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`

	var exists bool
	err := r.db.QueryRow(context.Background(), query, roomID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("repository: failed to check membership: %w", err)
	}
	return exists, nil
}

// GetMembers retrieves all members of a specific room.
func (r *RoomRepository) GetMembers(roomID string) ([]*domain.RoomMember, error) {
	query := `SELECT room_id, user_id, joined_at, last_read_at FROM room_members WHERE room_id = $1`

	rows, err := r.db.Query(context.Background(), query, roomID)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get room members: %w", err)
	}
	defer rows.Close()

	var members []*domain.RoomMember
	for rows.Next() {
		m := &domain.RoomMember{}
		if err := rows.Scan(&m.RoomID, &m.UserID, &m.JoinedAt, &m.LastReadAt); err != nil {
			return nil, fmt.Errorf("repository: failed to scan room member: %w", err)
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: error iterating members: %w", err)
	}
	return members, nil
}

// UpdateLastMessage updates the last_message_at timestamp for a room.
func (r *RoomRepository) UpdateLastMessage(roomID string, t time.Time) error {
	query := `UPDATE rooms SET last_message_at = $1 WHERE id = $2`
	_, err := r.db.Exec(context.Background(), query, t, roomID)
	if err != nil {
		return fmt.Errorf("repository: failed to update last message time: %w", err)
	}
	return nil
}
