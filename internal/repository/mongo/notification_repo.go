package mongo

import (
	"context"

	"todoapp/internal/entity"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type NotificationRepo struct {
	collection *mongo.Collection
}

func NewNotificationRepo(db *mongo.Database) *NotificationRepo {
	return &NotificationRepo{
		collection: db.Collection("notifications"),
	}
}

func (r *NotificationRepo) Create(ctx context.Context, notification *entity.Notification) error {
	_, err := r.collection.InsertOne(ctx, notification)
	return err
}
