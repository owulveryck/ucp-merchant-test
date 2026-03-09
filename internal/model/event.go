package model

import "time"

// DashboardEvent represents an activity event for the live dashboard.
type DashboardEvent struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	Summary   string    `json:"summary"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}
