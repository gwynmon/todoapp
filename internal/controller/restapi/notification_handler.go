package restapi

import (
	"encoding/json"
	"net/http"
	"todoapp/internal/controller/restapi/middleware"
	notificationuc "todoapp/internal/usecases/notification"
)

type NotificationHandler struct {
	service *notificationuc.Service
}

func NewNotificationHandler(service *notificationuc.Service) *NotificationHandler {
	return &NotificationHandler{
		service: service,
	}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	notifications, err := h.service.ListByUserID(
		r.Context(),
		int64(userID),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(notifications)
}
