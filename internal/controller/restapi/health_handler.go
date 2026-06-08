package restapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type HealthHandler struct {
	db          *sqlx.DB
	mongoClient *mongo.Client
	rdb         *redis.Client
	rabbitConn  *amqp.Connection
	logger      *slog.Logger
}

func NewHealthHandler(db *sqlx.DB, mongoClient *mongo.Client, rdb *redis.Client, rabbitConn *amqp.Connection, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:          db,
		mongoClient: mongoClient,
		rdb:         rdb,
		rabbitConn:  rabbitConn,
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
		writeJSONError(w, h.logger, "postgres is down", http.StatusServiceUnavailable)
		return
	}

	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		h.logger.Error("readiness check failed: mongodb", slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "mongodb is down", http.StatusServiceUnavailable)
		return
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		h.logger.Error("readiness check failed: redis", slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "redis is down", http.StatusServiceUnavailable)
		return
	}

	if h.rabbitConn.IsClosed() {
		h.logger.Error("readiness check failed: rabbitmq connection is closed")
		writeJSONError(w, h.logger, "rabbitmq is down", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
