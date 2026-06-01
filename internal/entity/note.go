package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Note struct {
	ID        bson.ObjectID          `json:"id" bson:"_id,omitempty"`
	TaskID    int                    `json:"task_id" bson:"task_id"`
	AuthorID  int                    `json:"author_id" bson:"author_id"`
	Text      string                 `json:"text" bson:"text"`
	CreatedAt time.Time              `json:"created_at" bson:"created_at"`
	Meta      map[string]interface{} `json:"meta,omitempty" bson:"meta,omitempty"`
}
