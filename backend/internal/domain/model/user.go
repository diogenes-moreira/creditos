package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleClient UserRole = "client"
)

type User struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	FirebaseUID string         `gorm:"uniqueIndex;not null"`
	Email       string         `gorm:"uniqueIndex;not null"`
	Role        UserRole       `gorm:"type:varchar(20);not null;default:'client'"`
	IsActive    bool           `gorm:"not null;default:true"`
	LastLoginAt *time.Time     `gorm:""`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func NewUser(firebaseUID, email string, role UserRole) (*User, error) {
	if firebaseUID == "" {
		return nil, fmt.Errorf("firebase UID is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if role != RoleAdmin && role != RoleClient {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	return &User{
		ID:          uuid.New(),
		FirebaseUID: firebaseUID,
		Email:       email,
		Role:        role,
		IsActive:    true,
	}, nil
}

func (u *User) Activate() {
	u.IsActive = true
}

func (u *User) Deactivate() {
	u.IsActive = false
}

func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
