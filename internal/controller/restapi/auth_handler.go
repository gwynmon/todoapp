package restapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"todoapp/internal/controller/restapi/middleware"
	"todoapp/internal/entity"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	svc      *middleware.AuthService
	validate *validator.Validate
}

func NewAuthHandler(svc *middleware.AuthService) *AuthHandler {
	return &AuthHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input entity.RegisterInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		http.Error(w, "validation failed", http.StatusUnprocessableEntity)
		return
	}

	if err := h.svc.Register(r.Context(), input); err != nil {
		switch {
		case errors.Is(err, entity.ErrUserAlreadyExists):
			http.Error(w, "email already exists", http.StatusConflict) // 409
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError) // 500
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input entity.LoginInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(input); err != nil {
		http.Error(w, "validation failed", http.StatusUnprocessableEntity)
		return
	}

	token, err := h.svc.Login(r.Context(), input)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized) // 401
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
