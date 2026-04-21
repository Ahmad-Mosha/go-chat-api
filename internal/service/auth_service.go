package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Ahmad-Mosha/go-chat-api/internal/config"
	"github.com/Ahmad-Mosha/go-chat-api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthService struct {
	userRepo domain.UserRepository
	cfg      *config.Config
}

// NewAuthService creates a new instance of the authentication service.
func NewAuthService(repo domain.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: repo,
		cfg:      cfg,
	}
}

// Login verifies credentials and returns a JWT token.
func (s *AuthService) Login(email, password string) (string, error) {
	// 1. Get user by email
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// 2. Compare passwords
	// bcrypt.CompareHashAndPassword returns nil if they match
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// 3. Generate JWT
	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// ValidateToken parses a JWT and returns the UserID it contains.
func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Validate the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}

// Internal helper to create the JWT
func (s *AuthService) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		"iat":     time.Now().Unix(),                     // Issued at
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
