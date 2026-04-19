package domain

import "time"

type Room struct {
	ID            string
	Name          string // Nullable for 1:1 chats
	IsGroup       bool   // False for 1:1, True for Group
	CreatedAt     time.Time
	LastMessageAt time.Time // Used to sort the chat list by newest message
}

type RoomMember struct {
	RoomID     string
	UserID     string
	JoinedAt   time.Time
	LastReadAt time.Time // Used to calculate unread message counts
}

type RoomRepository interface {
	Create(room *Room) error
	GetByID(id string) (*Room, error)
	FindOneToOneRoom(userA, userB string) (*Room, error)
	AddMember(roomID, userID string) error
	RemoveMember(roomID, userID string) error
	GetRoomsByUser(userID string) ([]*Room, error)
}
