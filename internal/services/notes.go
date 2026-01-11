package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/example/notes-template/internal/models"
)

var ErrNoteNotFound = errors.New("note not found")

type NoteService struct {
	db DBConn
}

func NewNoteService(db DBConn) *NoteService {
	return &NoteService{db: db}
}

func (s *NoteService) Create(ctx context.Context, params models.CreateNoteParams) (*models.Note, error) {
	note := &models.Note{}
	err := s.db.QueryRow(ctx,
		`INSERT INTO notes (user_id, title, body)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, title, body, created_at, updated_at`,
		params.UserID, params.Title, params.Body,
	).Scan(&note.ID, &note.UserID, &note.Title, &note.Body, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating note: %w", err)
	}

	return note, nil
}

func (s *NoteService) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Note, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, user_id, title, body, created_at, updated_at
		 FROM notes WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing notes: %w", err)
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note := &models.Note{}
		if err := rows.Scan(&note.ID, &note.UserID, &note.Title, &note.Body, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning note: %w", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating notes: %w", err)
	}

	return notes, nil
}

func (s *NoteService) GetByID(ctx context.Context, userID, noteID uuid.UUID) (*models.Note, error) {
	note := &models.Note{}
	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, title, body, created_at, updated_at
		 FROM notes WHERE id = $1 AND user_id = $2`,
		noteID, userID,
	).Scan(&note.ID, &note.UserID, &note.Title, &note.Body, &note.CreatedAt, &note.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting note: %w", err)
	}

	return note, nil
}

func (s *NoteService) Update(ctx context.Context, userID, noteID uuid.UUID, params models.UpdateNoteParams) (*models.Note, error) {
	note := &models.Note{}
	err := s.db.QueryRow(ctx,
		`UPDATE notes SET title = $1, body = $2 WHERE id = $3 AND user_id = $4
		 RETURNING id, user_id, title, body, created_at, updated_at`,
		params.Title, params.Body, noteID, userID,
	).Scan(&note.ID, &note.UserID, &note.Title, &note.Body, &note.CreatedAt, &note.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoteNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("updating note: %w", err)
	}

	return note, nil
}

func (s *NoteService) Delete(ctx context.Context, userID, noteID uuid.UUID) error {
	result, err := s.db.Exec(ctx, `DELETE FROM notes WHERE id = $1 AND user_id = $2`, noteID, userID)
	if err != nil {
		return fmt.Errorf("deleting note: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNoteNotFound
	}
	return nil
}
