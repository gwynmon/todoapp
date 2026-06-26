package mongo

import (
	"context"
	"errors"

	"todoapp/internal/entity"

	"go.mongodb.org/mongo-driver/v2/bson"
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

func (r *NotificationRepo) ListByUserID(ctx context.Context, userID int64) ([]*entity.Notification, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{"user_id": userID},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []*entity.Notification

	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepo) ExistsDeadlineNotification(ctx context.Context, taskID int64) (bool, error) {
	filter := bson.M{
		"task_id": taskID,
		"type":    "task.deadline_approaching",
	}

	err := r.collection.FindOne(ctx, filter).Err()

	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
