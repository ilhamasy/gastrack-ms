package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ilhamasy/gastrack-ms/internal/database"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrInvalidCreds = errors.New("invalid email or password")
)

// AuthService handles user authentication and registration.
type AuthService struct {
	db        database.DBPool
	jwtSecret []byte
}

// NewAuthService creates a new AuthService.
func NewAuthService(db database.DBPool, secret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: []byte(secret),
	}
}

// Register creates a new user with a hashed password and returns a JWT token.
func (s *AuthService) Register(ctx context.Context, email, password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	var userID uuid.UUID
	err = s.db.QueryRow(ctx,
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email, string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrUserExists
		}
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	return s.generateToken(userID.String())
}

// Login authenticates a user by email and password, returning a JWT token on success.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	var id uuid.UUID
	var hash string

	err := s.db.QueryRow(ctx, "SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvalidCreds
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", ErrInvalidCreds
	}

	return s.generateToken(id.String())
}

func (s *AuthService) generateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	return token.SignedString(s.jwtSecret)
}
