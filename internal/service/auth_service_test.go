package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewAuthService(mock, "test-secret")

	expectedID := uuid.New()
	mock.ExpectQuery("INSERT INTO users").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(expectedID))

	token, err := svc.Register(context.Background(), "test@test.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAuthService_Register_DuplicateUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewAuthService(mock, "test-secret")

	pgErr := &pgconn.PgError{Code: "23505"}
	mock.ExpectQuery("INSERT INTO users").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(pgErr)

	_, err = svc.Register(context.Background(), "test@test.com", "password123")
	if err != ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewAuthService(mock, "test-secret")

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	mock.ExpectQuery("SELECT id, password_hash FROM users").
		WithArgs("test@test.com").
		WillReturnRows(pgxmock.NewRows([]string{"id", "password_hash"}).AddRow(userID, string(hashedPassword)))

	token, err := svc.Login(context.Background(), "test@test.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewAuthService(mock, "test-secret")

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)

	mock.ExpectQuery("SELECT id, password_hash FROM users").
		WithArgs("test@test.com").
		WillReturnRows(pgxmock.NewRows([]string{"id", "password_hash"}).AddRow(userID, string(hashedPassword)))

	_, err = svc.Login(context.Background(), "test@test.com", "wrong-password")
	if err != ErrInvalidCreds {
		t.Errorf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	svc := NewAuthService(mock, "test-secret")

	mock.ExpectQuery("SELECT id, password_hash FROM users").
		WithArgs("nonexistent@test.com").
		WillReturnRows(pgxmock.NewRows([]string{"id", "password_hash"}))

	_, err = svc.Login(context.Background(), "nonexistent@test.com", "password123")
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	svc := &AuthService{jwtSecret: []byte("test-secret")}

	token, err := svc.generateToken("test-user-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}
