package models

import (
	"encoding/json"
	"time"
)

const (
	StatusPending   = "PENDING"
	StatusRunning   = "RUNNING"
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED"
)

type Task struct {
	ID         string          `json:"id"`
	Payload    any             `json:"payload"`
	Status     string          `json:"status"`
	Result     json.RawMessage `json:"result"`
	Error      string          `json:"error"`
	CreatedAt  time.Time       `json:"created_at"`
	FinishedAt time.Time       `json:"finished_at"`
}
