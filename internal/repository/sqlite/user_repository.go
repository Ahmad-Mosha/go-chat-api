package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ahmad-Mosha/go-chat-api/internal/domain"
)

// blank identifier , force the compiler to check implementation immediately
var _ domain.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new instance of the sqlite user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at)
		VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(context.Background(), query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt)

	if err != nil {
		return fmt.Errorf("repository: failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their UUID.
func (r *UserRepository) GetByID(id string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id = ?`

	user := &domain.User{}
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("repository: user with id %s not found", id)
		}
		return nil, fmt.Errorf("repository: failed to get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = ?`

	user := &domain.User{}
	err := r.db.QueryRowContext(context.Background(), query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("repository: user with email %s not found", email)
		}
		return nil, fmt.Errorf("repository: failed to get user by email: %w", err)
	}

	return user, nil
}

// Update updates the user's information in the database.
func (r *UserRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET username = ?, email = ?, password_hash = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(context.Background(), query,
		user.Username, user.Email, user.PasswordHash, user.ID)

	if err != nil {
		return fmt.Errorf("repository: failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("repository: user with id %s not found for update", user.ID)
	}

	return nil
}
