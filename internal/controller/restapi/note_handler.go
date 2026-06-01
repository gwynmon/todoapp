package restapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"todoapp/internal/controller/restapi/middleware"
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
	userID, _ := middleware.GetUserID(r.Context())
	taskID, _ := strconv.Atoi(r.PathValue("taskID"))

	var input struct {
		Text string         `json:"text"`
		Meta map[string]any `json:"meta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.svc.Create(r.Context(), userID, taskID, input.Text, input.Meta); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	taskID, _ := strconv.Atoi(r.PathValue("taskID"))
	notes, err := h.svc.GetByTaskID(r.Context(), taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	noteID, err := bson.ObjectIDFromHex(r.PathValue("noteID"))
	if err != nil {
		http.Error(w, "invalid note id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), userID, noteID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
