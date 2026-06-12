package restapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"todoapp/internal/entity"
	"todoapp/internal/usecases/auth"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	svc      *auth.Service
	validate *validator.Validate
	logger   *slog.Logger
}

func NewAuthHandler(svc *auth.Service, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		svc:      svc,
		validate: validator.New(),
		logger:   logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input entity.RegisterInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request body", slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.logger.Warn("validation failed", slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "validation failed", http.StatusUnprocessableEntity)
		return
	}

	if err := h.svc.Register(r.Context(), input); err != nil {
		switch {
		case errors.Is(err, entity.ErrUserAlreadyExists):
			h.logger.Info("registration failed: user exists", slog.String("email", input.Email))
			writeJSONError(w, h.logger, "email already exists", http.StatusConflict) // 409
		default:
			h.logger.Error("registration failed", slog.String("error", err.Error()))
			writeJSONError(w, h.logger, "internal server error", http.StatusInternalServerError) // 500
		}
		return
	}

	h.logger.Info("user registered", slog.String("email", input.Email))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input entity.LoginInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("invalid request body", slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		h.logger.Warn("validation failed", slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "validation failed", http.StatusUnprocessableEntity)
		return
	}

	token, err := h.svc.Login(r.Context(), input)
	if err != nil {
		h.logger.Warn("login failed", slog.String("email", input.Email), slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "invalid credentials", http.StatusUnauthorized) // 401
		return
	}

	h.logger.Info("user logged in", slog.String("email", input.Email))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
