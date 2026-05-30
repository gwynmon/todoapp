package restapi

import (
	"encoding/json"
	"net/http"
	"todoapp/internal/controller/restapi/middleware"
	"todoapp/internal/entity"
	"todoapp/internal/usecases/task"

	"github.com/go-playground/validator/v10"
)

type TaskHandler struct {
	svc      *task.Service
	validate *validator.Validate
}

func NewTaskHandler(svc *task.Service) *TaskHandler {
	return &TaskHandler{svc: svc, validate: validator.New()}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input entity.Task
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.svc.Create(r.Context(), userID, input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *TaskHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tasks, err := h.svc.GetByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}
