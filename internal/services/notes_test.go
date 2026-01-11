package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/example/notes-template/internal/models"
)

type mockRow struct {
	scan func(dest ...any) error
}

func (m mockRow) Scan(dest ...any) error {
	return m.scan(dest...)
}

type mockRows struct {
	rows [][]any
	idx  int
	err  error
}

func (m *mockRows) Close() {}

func (m *mockRows) Err() error { return m.err }

func (m *mockRows) Next() bool {
	if m.idx >= len(m.rows) {
		return false
	}
	m.idx++
	return true
}

func (m *mockRows) Scan(dest ...any) error {
	row := m.rows[m.idx-1]
	for i := range dest {
		switch d := dest[i].(type) {
		case *uuid.UUID:
			*d = row[i].(uuid.UUID)
		case *string:
			*d = row[i].(string)
		case *time.Time:
			*d = row[i].(time.Time)
		default:
			return errors.New("unsupported scan type")
		}
	}
	return nil
}

type mockCommandTag struct {
	affected int64
}

func (m mockCommandTag) RowsAffected() int64 { return m.affected }

type mockDB struct {
	queryRow func(ctx context.Context, sql string, args ...any) Row
	query    func(ctx context.Context, sql string, args ...any) (Rows, error)
	exec     func(ctx context.Context, sql string, args ...any) (CommandTag, error)
}

func (m *mockDB) Exec(ctx context.Context, sql string, args ...any) (CommandTag, error) {
	return m.exec(ctx, sql, args...)
}

func (m *mockDB) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return m.query(ctx, sql, args...)
}

func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return m.queryRow(ctx, sql, args...)
}

func TestNoteService_Create(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()
	now := time.Now()
	db := &mockDB{
		queryRow: func(ctx context.Context, sql string, args ...any) Row {
			return mockRow{scan: func(dest ...any) error {
				*dest[0].(*uuid.UUID) = noteID
				*dest[1].(*uuid.UUID) = userID
				*dest[2].(*string) = "Title"
				*dest[3].(*string) = "Body"
				*dest[4].(*time.Time) = now
				*dest[5].(*time.Time) = now
				return nil
			}}
		},
	}

	svc := NewNoteService(db)
	note, err := svc.Create(context.Background(), models.CreateNoteParams{
		UserID: userID,
		Title:  "Title",
		Body:   "Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.ID != noteID {
		t.Fatalf("expected note id %s, got %s", noteID, note.ID)
	}
}

func TestNoteService_ListByUser(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()
	now := time.Now()
	db := &mockDB{
		query: func(ctx context.Context, sql string, args ...any) (Rows, error) {
			return &mockRows{rows: [][]any{{noteID, userID, "Title", "Body", now, now}}}, nil
		},
	}

	svc := NewNoteService(db)
	notes, err := svc.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
}

func TestNoteService_Update_NotFound(t *testing.T) {
	db := &mockDB{
		queryRow: func(ctx context.Context, sql string, args ...any) Row {
			return mockRow{scan: func(dest ...any) error {
				return pgx.ErrNoRows
			}}
		},
	}

	svc := NewNoteService(db)
	_, err := svc.Update(context.Background(), uuid.New(), uuid.New(), models.UpdateNoteParams{Title: "T", Body: "B"})
	if !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf("expected ErrNoteNotFound, got %v", err)
	}
}

func TestNoteService_Delete_NotFound(t *testing.T) {
	db := &mockDB{
		exec: func(ctx context.Context, sql string, args ...any) (CommandTag, error) {
			return mockCommandTag{affected: 0}, nil
		},
	}

	svc := NewNoteService(db)
	err := svc.Delete(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrNoteNotFound) {
		t.Fatalf("expected ErrNoteNotFound, got %v", err)
	}
}
