package restapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type HealthHandler struct {
	db          *sqlx.DB
	mongoClient *mongo.Client
	logger      *slog.Logger
}

func NewHealthHandler(db *sqlx.DB, mongoClient *mongo.Client, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:          db,
		mongoClient: mongoClient,
		logger:      logger,
	}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		h.logger.Error("readiness check failed: postgres", slog.String("error", err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable) // 503
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "postgres is down", "details": err.Error()})
		return
	}

	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		h.logger.Error("readiness check failed: mongodb", slog.String("error", err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable) // 503
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "mongodb is down", "details": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
