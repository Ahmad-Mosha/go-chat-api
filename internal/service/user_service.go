package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ahmad-Mosha/go-chat-api/internal/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

// UserProfile is the public user shape returned by the API.
type UserProfile struct {
	ID       string
	Username string
	Email    string
}

func toUserProfile(user *domain.User) *UserProfile {
	return &UserProfile{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

type UserService struct {
	userRepo domain.UserRepository
}

// NewUserService creates a new instance of the user service.
func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{
		userRepo: repo,
	}
}

// Register handles the logic for creating a new user account.
func (s *UserService) Register(username, email, password string) (*domain.User, error) {
	// 1. Check if user already exists by email
	existingUser, err := s.userRepo.GetByEmail(email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// 2. Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. Create the User entity
	user := &domain.User{
		ID:           uuid.NewString(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}

	// 4. Save to repository
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

// GetProfile fetches a user's public profile by ID.
func (s *UserService) GetProfile(id string) (*UserProfile, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return toUserProfile(user), nil
}

// GetByEmail fetches a user's public profile by email (for starting 1:1 chats).
func (s *UserService) GetByEmail(email string) (*UserProfile, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return toUserProfile(user), nil
}
