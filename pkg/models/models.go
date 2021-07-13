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

type EngineeringWorkOrder struct {
	ID            int
	Title         string
	Created       time.Time
	Location      *Location
	CreatedBy     *SysUser
	RequestStatus *RequestStatus
	Notes         []*EngineeringWorkOrderNote
}

type SysUser struct {
	ID             int
	FullName       string
	Username       string
	HashedPassword []byte
	Created        time.Time
	CreatedBy      *SysUser
	Position       *Position
	Manager        *SysUser
	ActiveUser     bool
}

type Department struct {
	ID        int
	Title     string
	Manager   *SysUser
	Created   time.Time
	CreatedBy *SysUser
}

type EngineeringWorkOrderChange struct {
	ID          int
	WorkOrderID int
	Field       string
	OldValue    string
	NewValue    string
	Created     time.Time
	CreatedBy   *SysUser
}

type EngineeringWorkOrderNote struct {
	ID        int
	Content   string
	Created   time.Time
	CreatedBy *SysUser
}

type Location struct {
	ID        int
	Title     string
	Created   time.Time
	CreatedBy *SysUser
}

type Position struct {
	ID           int
	Title        string
	Created      time.Time
	CreatedByID  *SysUser
	DepartmentID *Department
}

type RequestStatus struct {
	ID        int
	Title     string
	Created   time.Time
	CreatedBy *SysUser
	IsClosed  bool
}
