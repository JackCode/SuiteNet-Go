package models

import (
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: mo matching record found")

type MaintenanceRequest struct {
	ID          int
	Title       string
	Description string
	Created     time.Time
	Status      string
}
