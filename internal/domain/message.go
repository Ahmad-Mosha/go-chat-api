package domain

import "time"

type Message struct {
	ID        string
	RoomID    string
	SenderID  string
	Content   string
	CreatedAt time.Time
}

type MessageRepository interface {
	Create(msg *Message) error
	GetMessagesByRoom(roomID string, limit, offset int) ([]*Message, error)
	GetByID(id string) (*Message, error)
	Delete(id string) error
}
