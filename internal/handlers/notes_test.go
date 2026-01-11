package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/example/notes-template/internal/models"
	"github.com/example/notes-template/internal/services"
)

type mockNoteService struct {
	create func(ctx context.Context, params models.CreateNoteParams) (*models.Note, error)
	list   func(ctx context.Context, userID uuid.UUID) ([]*models.Note, error)
	get    func(ctx context.Context, userID, noteID uuid.UUID) (*models.Note, error)
	update func(ctx context.Context, userID, noteID uuid.UUID, params models.UpdateNoteParams) (*models.Note, error)
	delete func(ctx context.Context, userID, noteID uuid.UUID) error
}

func (m *mockNoteService) Create(ctx context.Context, params models.CreateNoteParams) (*models.Note, error) {
	return m.create(ctx, params)
}

func (m *mockNoteService) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Note, error) {
	return m.list(ctx, userID)
}

func (m *mockNoteService) GetByID(ctx context.Context, userID, noteID uuid.UUID) (*models.Note, error) {
	return m.get(ctx, userID, noteID)
}

func (m *mockNoteService) Update(ctx context.Context, userID, noteID uuid.UUID, params models.UpdateNoteParams) (*models.Note, error) {
	return m.update(ctx, userID, noteID, params)
}

func (m *mockNoteService) Delete(ctx context.Context, userID, noteID uuid.UUID) error {
	return m.delete(ctx, userID, noteID)
}

func TestNoteHandler_List(t *testing.T) {
	user := &models.User{ID: uuid.New()}
	note := &models.Note{ID: uuid.New(), UserID: user.ID, Title: "Title", Body: "Body", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	service := &mockNoteService{
		list: func(ctx context.Context, userID uuid.UUID) ([]*models.Note, error) {
			return []*models.Note{note}, nil
		},
	}

	h := NewNoteHandler(service)
	req := httptest.NewRequest(http.MethodGet, "/api/notes", nil)
	ctx := SetUserInContext(req.Context(), user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload map[string][]*models.Note
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(payload["notes"]) != 1 {
		t.Fatalf("expected 1 note, got %d", len(payload["notes"]))
	}
}

func TestNoteHandler_Create_Invalid(t *testing.T) {
	user := &models.User{ID: uuid.New()}
	service := &mockNoteService{
		create: func(ctx context.Context, params models.CreateNoteParams) (*models.Note, error) {
			return nil, nil
		},
	}

	h := NewNoteHandler(service)
	body := strings.NewReader(`{"title":"","body":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/notes", body)
	ctx := SetUserInContext(req.Context(), user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestNoteHandler_Update_NotFound(t *testing.T) {
	user := &models.User{ID: uuid.New()}
	noteID := uuid.New()

	service := &mockNoteService{
		update: func(ctx context.Context, userID, noteID uuid.UUID, params models.UpdateNoteParams) (*models.Note, error) {
			return nil, services.ErrNoteNotFound
		},
	}

	h := NewNoteHandler(service)
	body := strings.NewReader(`{"title":"Title","body":"Body"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/notes/"+noteID.String(), body)
	req.SetPathValue("id", noteID.String())
	ctx := SetUserInContext(req.Context(), user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestNoteHandler_Delete_NotFound(t *testing.T) {
	user := &models.User{ID: uuid.New()}
	noteID := uuid.New()

	service := &mockNoteService{
		delete: func(ctx context.Context, userID, noteID uuid.UUID) error {
			return services.ErrNoteNotFound
		},
	}

	h := NewNoteHandler(service)
	req := httptest.NewRequest(http.MethodDelete, "/api/notes/"+noteID.String(), nil)
	req.SetPathValue("id", noteID.String())
	ctx := SetUserInContext(req.Context(), user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}
