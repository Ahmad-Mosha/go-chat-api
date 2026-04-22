package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ahmad-Mosha/go-chat-api/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrNotRoomMember   = errors.New("you are not a member of this room")
	ErrRoomExists      = errors.New("a conversation with this user already exists")
	ErrCannotChatSelf  = errors.New("cannot create a conversation with yourself")
	ErrMemberNotFound  = errors.New("member not found in room")
	ErrAlreadyMember   = errors.New("user is already a member of this room")
)

// ChatService handles the business logic for rooms and messages.
type ChatService struct {
	roomRepo domain.RoomRepository
	msgRepo  domain.MessageRepository
}

// NewChatService creates a new instance of the chat service.
func NewChatService(roomRepo domain.RoomRepository, msgRepo domain.MessageRepository) *ChatService {
	return &ChatService{
		roomRepo: roomRepo,
		msgRepo:  msgRepo,
	}
}

// CreateRoom creates a new chat room and adds the creator as the first member.
// For 1:1 chats (isGroup=false), it checks if a conversation already exists.
func (s *ChatService) CreateRoom(creatorID, name string, isGroup bool, memberIDs []string) (*domain.Room, error) {
	// For 1:1 chats, check if a room already exists between these two users
	if !isGroup {
		if len(memberIDs) != 1 {
			return nil, fmt.Errorf("1:1 chat requires exactly one other member")
		}
		if memberIDs[0] == creatorID {
			return nil, ErrCannotChatSelf
		}

		existingRoom, err := s.roomRepo.FindOneToOneRoom(creatorID, memberIDs[0])
		if err != nil {
			return nil, fmt.Errorf("failed to check existing room: %w", err)
		}
		if existingRoom != nil {
			return nil, ErrRoomExists
		}
	}

	now := time.Now()
	room := &domain.Room{
		ID:            uuid.NewString(),
		Name:          name,
		IsGroup:       isGroup,
		CreatedAt:     now,
		LastMessageAt: now,
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// Add the creator as a member
	if err := s.roomRepo.AddMember(room.ID, creatorID); err != nil {
		return nil, fmt.Errorf("failed to add creator to room: %w", err)
	}

	// Add other members
	for _, memberID := range memberIDs {
		if memberID == creatorID {
			continue // Already added
		}
		if err := s.roomRepo.AddMember(room.ID, memberID); err != nil {
			return nil, fmt.Errorf("failed to add member %s: %w", memberID, err)
		}
	}

	return room, nil
}

// GetUserRooms returns all rooms for a user, sorted by most recent message.
func (s *ChatService) GetUserRooms(userID string) ([]*domain.Room, error) {
	rooms, err := s.roomRepo.GetRoomsByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user rooms: %w", err)
	}
	return rooms, nil
}

// GetRoom returns a single room by ID, only if the user is a member.
func (s *ChatService) GetRoom(roomID, userID string) (*domain.Room, error) {
	isMember, err := s.roomRepo.IsMember(roomID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, ErrNotRoomMember
	}

	room, err := s.roomRepo.GetByID(roomID)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

// SendMessage persists a message and updates the room's last_message_at.
func (s *ChatService) SendMessage(roomID, senderID, content string) (*domain.Message, error) {
	// Verify the sender is a member of the room
	isMember, err := s.roomRepo.IsMember(roomID, senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, ErrNotRoomMember
	}

	now := time.Now()
	msg := &domain.Message{
		ID:        uuid.NewString(),
		RoomID:    roomID,
		SenderID:  senderID,
		Content:   content,
		CreatedAt: now,
	}

	if err := s.msgRepo.Create(msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update the room's last_message_at so it sorts correctly in the chat list
	if err := s.roomRepo.UpdateLastMessage(roomID, now); err != nil {
		return nil, fmt.Errorf("failed to update room timestamp: %w", err)
	}

	return msg, nil
}

// GetRoomMessages returns paginated messages for a room, only if the user is a member.
func (s *ChatService) GetRoomMessages(roomID, userID string, limit, offset int) ([]*domain.Message, error) {
	// Verify the user is a member of the room
	isMember, err := s.roomRepo.IsMember(roomID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, ErrNotRoomMember
	}

	// Set sensible defaults
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	messages, err := s.msgRepo.GetMessagesByRoom(roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	return messages, nil
}
