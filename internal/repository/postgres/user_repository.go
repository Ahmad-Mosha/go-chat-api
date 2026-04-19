package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ahmad-Mosha/go-chat-api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// blank identifier , force the compiler to check implementation immediately
var _ domain.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new instance of the postgres user repository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(context.Background(), query,
		user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt)

	if err != nil {
		return fmt.Errorf("repository: failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their UUID.
func (r *UserRepository) GetByID(id string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id = $1`

	user := &domain.User{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository: user with id %s not found", id)
		}
		return nil, fmt.Errorf("repository: failed to get user by id: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address.
func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1`

	user := &domain.User{}
	err := r.db.QueryRow(context.Background(), query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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
		SET username = $1, email = $2, password_hash = $3
		WHERE id = $4`

	result, err := r.db.Exec(context.Background(), query,
		user.Username, user.Email, user.PasswordHash, user.ID)

	if err != nil {
		return fmt.Errorf("repository: failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repository: user with id %s not found for update", user.ID)
	}

	return nil
}
