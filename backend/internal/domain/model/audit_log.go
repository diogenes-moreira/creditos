package model

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID      *uuid.UUID `gorm:"type:uuid;index"`
	Action      string     `gorm:"not null"`
	EntityType  string     `gorm:"index;not null"`
	EntityID    string     `gorm:"index;not null"`
	Description string     `gorm:"not null"`
	IP          string     `gorm:""`
	UserAgent   string     `gorm:""`
	CreatedAt   time.Time  `gorm:"not null;index"`
}

func NewAuditLog(userID *uuid.UUID, action, entityType, entityID, description, ip, userAgent string) *AuditLog {
	return &AuditLog{
		ID:          uuid.New(),
		UserID:      userID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		IP:          ip,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}
}
