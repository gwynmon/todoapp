package restapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"todoapp/internal/controller/restapi/middleware"
	"todoapp/internal/entity"
	"todoapp/internal/usecases/task"
	"todoapp/pkg/cache"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

type TaskHandler struct {
	svc      *task.Service
	cache    *cache.Cache
	validate *validator.Validate
	logger   *slog.Logger
}

func NewTaskHandler(svc *task.Service, cache *cache.Cache, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{
		svc:      svc,
		cache:    cache,
		validate: validator.New(),
		logger:   logger,
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeJSONError(w, h.logger, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input entity.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, h.logger, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		writeJSONError(w, h.logger, "validation failed", http.StatusUnprocessableEntity)
		return
	}

	if err := h.svc.Create(r.Context(), userID, input); err != nil {
		writeJSONError(w, h.logger, "internal server error", http.StatusInternalServerError)
		return
	}

	// Инвалидация кеша
	h.cache.Delete(r.Context(), cache.TasksKey(userID))

	w.WriteHeader(http.StatusCreated)
}

func (h *TaskHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeJSONError(w, h.logger, "unauthorized", http.StatusUnauthorized)
		return
	}

	cacheKey := cache.TasksKey(userID)
	var tasks []entity.Task

	err = h.cache.Get(r.Context(), cacheKey, &tasks)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tasks)
		return
	}

	if !errors.Is(err, redis.Nil) {
		h.logger.Warn("cache get failed", slog.String("error", err.Error()))
	}

	tasks, err = h.svc.GetByUser(r.Context(), userID)
	if err != nil {
		writeJSONError(w, h.logger, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.cache.Set(r.Context(), cacheKey, tasks); err != nil {
		h.logger.Warn("cache set failed", slog.String("error", err.Error()))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeJSONError(w, h.logger, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskIDStr := r.PathValue("taskID")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		writeJSONError(w, h.logger, "invalid task id", http.StatusBadRequest)
		return
	}

	task, err := h.svc.GetByID(r.Context(), userID, taskID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFoundOrAccessDenied) {
			writeJSONError(w, h.logger, "task not found or access denied", http.StatusNotFound)
			return
		}
		writeJSONError(w, h.logger, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeJSONError(w, h.logger, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.Atoi(r.PathValue("taskID"))
	if err != nil {
		writeJSONError(w, h.logger, "invalid task id", http.StatusBadRequest)
		return
	}

	var input entity.UpdateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSONError(w, h.logger, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.svc.Update(r.Context(), userID, taskID, input); err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFoundOrAccessDenied):
			writeJSONError(w, h.logger, "not found or access denied", http.StatusNotFound)
		case errors.Is(err, entity.ErrInvalidStatus):
			writeJSONError(w, h.logger, "invalid status", http.StatusBadRequest)
		default:
			writeJSONError(w, h.logger, "internal error", http.StatusInternalServerError)
		}
		return
	}

	h.cache.Delete(r.Context(), cache.TasksKey(userID))

	w.WriteHeader(http.StatusOK)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeJSONError(w, h.logger, "unauthorized", http.StatusUnauthorized)
		return
	}

	taskID, err := strconv.Atoi(r.PathValue("taskID"))
	if err != nil {
		writeJSONError(w, h.logger, "invalid task id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), userID, taskID); err != nil {
		if errors.Is(err, entity.ErrNotFoundOrAccessDenied) {
			writeJSONError(w, h.logger, "not found or access denied", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to delete task", slog.Int("task_id", taskID), slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "internal error", http.StatusInternalServerError)
		return
	}

	h.cache.Delete(r.Context(), cache.TasksKey(userID))

	w.WriteHeader(http.StatusNoContent)
}
