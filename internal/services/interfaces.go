package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/example/notes-template/internal/models"
)

// UserServiceInterface defines the contract for user operations.
type UserServiceInterface interface {
	Create(ctx context.Context, params models.CreateUserParams) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
}

// AuthServiceInterface defines the contract for authentication operations.
type AuthServiceInterface interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hash, password string) bool
	GenerateSessionToken() (token string, hash string, err error)
	CreateSession(ctx context.Context, userID uuid.UUID) (token string, err error)
	ValidateSession(ctx context.Context, token string) (*models.User, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error
}

// EmailServiceInterface defines the contract for email operations.
type EmailServiceInterface interface {
	SendVerificationEmail(ctx context.Context, userID uuid.UUID, email string) error
	VerifyEmail(ctx context.Context, token string) error
	SendMagicLinkEmail(ctx context.Context, email string) error
	VerifyMagicLink(ctx context.Context, token string) (string, error)
	SendPasswordResetEmail(ctx context.Context, userID uuid.UUID, email string) error
	VerifyPasswordResetToken(ctx context.Context, token string) (uuid.UUID, error)
	MarkPasswordResetUsed(ctx context.Context, token string) error
}

// NoteServiceInterface defines the contract for notes operations.
type NoteServiceInterface interface {
	Create(ctx context.Context, params models.CreateNoteParams) (*models.Note, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Note, error)
	GetByID(ctx context.Context, userID, noteID uuid.UUID) (*models.Note, error)
	Update(ctx context.Context, userID, noteID uuid.UUID, params models.UpdateNoteParams) (*models.Note, error)
	Delete(ctx context.Context, userID, noteID uuid.UUID) error
}
