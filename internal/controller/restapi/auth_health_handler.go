package restapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type AuthHealthHandler struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewAuthHealthHandler(db *sqlx.DB, logger *slog.Logger) *AuthHealthHandler {
	return &AuthHealthHandler{
		db:     db,
		logger: logger,
	}
}

func (h *AuthHealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (h *AuthHealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		h.logger.Error("readiness check failed: postgres", slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "postgres is down", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
