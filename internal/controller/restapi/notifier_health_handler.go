package restapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type NotifierHealthHandler struct {
	mongoClient *mongo.Client
	rabbitConn  *amqp.Connection
	logger      *slog.Logger
}

func NewNotifierHealthHandler(mongoClient *mongo.Client, rabbitConn *amqp.Connection, logger *slog.Logger) *NotifierHealthHandler {
	return &NotifierHealthHandler{
		mongoClient: mongoClient,
		rabbitConn:  rabbitConn,
		logger:      logger,
	}
}

func (h *NotifierHealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (h *NotifierHealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		h.logger.Error("readiness check failed: mongodb", slog.String("error", err.Error()))
		writeJSONError(w, h.logger, "mongodb is down", http.StatusServiceUnavailable)
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
