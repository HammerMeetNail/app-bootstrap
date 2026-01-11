package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/example/notes-template/internal/models"
	"github.com/example/notes-template/internal/services"
)

type NoteHandler struct {
	noteService services.NoteServiceInterface
}

func NewNoteHandler(noteService services.NoteServiceInterface) *NoteHandler {
	return &NoteHandler{noteService: noteService}
}

type NoteRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	notes, err := h.noteService.ListByUser(r.Context(), user.ID)
	if err != nil {
		log.Printf("Error listing notes: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"notes": notes})
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req NoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	if req.Title == "" || len(req.Title) > 200 {
		writeError(w, http.StatusBadRequest, "Title must be between 1 and 200 characters")
		return
	}
	if req.Body == "" || len(req.Body) > 5000 {
		writeError(w, http.StatusBadRequest, "Body must be between 1 and 5000 characters")
		return
	}

	note, err := h.noteService.Create(r.Context(), models.CreateNoteParams{
		UserID: user.ID,
		Title:  req.Title,
		Body:   req.Body,
	})
	if err != nil {
		log.Printf("Error creating note: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"note": note})
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid note id")
		return
	}

	var req NoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	if req.Title == "" || len(req.Title) > 200 {
		writeError(w, http.StatusBadRequest, "Title must be between 1 and 200 characters")
		return
	}
	if req.Body == "" || len(req.Body) > 5000 {
		writeError(w, http.StatusBadRequest, "Body must be between 1 and 5000 characters")
		return
	}

	note, err := h.noteService.Update(r.Context(), user.ID, noteID, models.UpdateNoteParams{
		Title: req.Title,
		Body:  req.Body,
	})
	if err != nil {
		if err == services.ErrNoteNotFound {
			writeError(w, http.StatusNotFound, "Note not found")
			return
		}
		log.Printf("Error updating note: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"note": note})
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid note id")
		return
	}

	if err := h.noteService.Delete(r.Context(), user.ID, noteID); err != nil {
		if err == services.ErrNoteNotFound {
			writeError(w, http.StatusNotFound, "Note not found")
			return
		}
		log.Printf("Error deleting note: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Note deleted"})
}

func (h *NoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	noteID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid note id")
		return
	}

	note, err := h.noteService.GetByID(r.Context(), user.ID, noteID)
	if err != nil {
		if err == services.ErrNoteNotFound {
			writeError(w, http.StatusNotFound, "Note not found")
			return
		}
		log.Printf("Error getting note: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"note": note})
}
