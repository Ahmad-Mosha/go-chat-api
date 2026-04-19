package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ahmad-Mosha/go-chat-api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create saves a new message to the database.
func (r *MessageRepository) Create(msg *domain.Message) error {
	query := `
		INSERT INTO messages (id, room_id, sender_id, content, created_at) 
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(context.Background(), query, 
		msg.ID, msg.RoomID, msg.SenderID, msg.Content, msg.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("repository: failed to create message: %w", err)
	}

	return nil
}

// GetMessagesByRoom fetches a paginated list of messages for a specific room.
func (r *MessageRepository) GetMessagesByRoom(roomID string, limit, offset int) ([]*domain.Message, error) {
	// We order by created_at ASC to get the conversation flow (oldest to newest)
	query := `
		SELECT id, room_id, sender_id, content, created_at 
		FROM messages 
		WHERE room_id = $1 
		ORDER BY created_at ASC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(context.Background(), query, roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to fetch messages: %w", err)
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		msg := &domain.Message{}
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.SenderID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository: failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: error iterating messages: %w", err)
	}

	return messages, nil
}

// GetByID retrieves a single message by its ID.
func (r *MessageRepository) GetByID(id string) (*domain.Message, error) {
	query := `SELECT id, room_id, sender_id, content, created_at FROM messages WHERE id = $1`

	msg := &domain.Message{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&msg.ID, &msg.RoomID, &msg.SenderID, &msg.Content, &msg.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository: message not found")
		}
		return nil, fmt.Errorf("repository: failed to get message by id: %w", err)
	}

	return msg, nil
}

// Delete removes a message from the database.
func (r *MessageRepository) Delete(id string) error {
	query := `DELETE FROM messages WHERE id = $1`
	
	result, err := r.db.Exec(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("repository: failed to delete message: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repository: message with id %s not found", id)
	}

	return nil
}
