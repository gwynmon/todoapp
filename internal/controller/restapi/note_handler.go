package restapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"todoapp/internal/controller/restapi/middleware"
	"todoapp/internal/entity"
	"todoapp/internal/usecases/note"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type NoteHandler struct {
	svc *note.Service
}

func NewNoteHandler(svc *note.Service) *NoteHandler {
	return &NoteHandler{svc: svc}
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeJSONError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.Atoi(r.PathValue("taskID"))
	if err != nil {
		writeJSONError(w, nil, "invalid task id", http.StatusBadRequest)
		return
	}

	var input struct {
		Text string         `json:"text"`
		Meta map[string]any `json:"meta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, nil, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.svc.Create(r.Context(), userID, taskID, input.Text, input.Meta); err != nil {
		if errors.Is(err, entity.ErrAccessDenied) {
			writeJSONError(w, nil, err.Error(), http.StatusForbidden)
			return
		}

		writeJSONError(w, nil, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeJSONError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.Atoi(r.PathValue("taskID"))
	if err != nil {
		writeJSONError(w, nil, "invalid task id", http.StatusBadRequest)
		return
	}

	notes, err := h.svc.GetByTaskID(r.Context(), userID, taskID)
	if err != nil {
		if errors.Is(err, entity.ErrAccessDenied) {
			writeJSONError(w, nil, err.Error(), http.StatusForbidden)
			return
		}

		writeJSONError(w, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeJSONError(w, nil, "unauthorized", http.StatusUnauthorized)
		return
	}
	noteID, err := bson.ObjectIDFromHex(r.PathValue("noteID"))
	if err != nil {
		writeJSONError(w, nil, "invalid note id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), userID, noteID); err != nil {
		if errors.Is(err, entity.ErrAccessDenied) {
			writeJSONError(w, nil, err.Error(), http.StatusForbidden)
			return
		}
		
		writeJSONError(w, nil, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
