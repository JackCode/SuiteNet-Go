package models

import (
	"errors"
	"time"
)

var (
	ErrNoRecord           = errors.New("models: mo matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateUsername  = errors.New("models: duplicate username")
)

type WorkOrder struct {
	ID          int
	Title       string
	Description string
	Created     time.Time
	Status      string
}

type User struct {
	ID             int
	Name           string
	Username       string
	HashedPassword []byte
	Created        time.Time
}
