package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Notification struct {
	ID     bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID int64         `bson:"user_id" json:"user_id"`
	TaskID int64         `bson:"task_id" json:"task_id"`

	Type    string `bson:"type" json:"type"`
	Message string `bson:"message" json:"message"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
