package entity

import "time"

type TaskEvent struct {
	EventType string                 `json:"event_type"`
	TaskID    int64                  `json:"task_id"`
	UserID    int64                  `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}
