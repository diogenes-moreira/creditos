package service

import (
	"context"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/google/uuid"
)

type AuditService struct {
	repo port.AuditLogRepository
}

func NewAuditService(repo port.AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Record(ctx context.Context, userID *uuid.UUID, action, entityType, entityID, description string) {
	ip, _ := ctx.Value("ip").(string)
	userAgent, _ := ctx.Value("userAgent").(string)
	log := model.NewAuditLog(userID, action, entityType, entityID, description, ip, userAgent)
	_ = s.repo.Create(ctx, log)
}

func (s *AuditService) FindByEntity(ctx context.Context, entityType, entityID string, offset, limit int) ([]model.AuditLog, int64, error) {
	return s.repo.FindByEntity(ctx, entityType, entityID, offset, limit)
}

func (s *AuditService) FindByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]model.AuditLog, int64, error) {
	return s.repo.FindByUser(ctx, userID, offset, limit)
}

func (s *AuditService) FindAll(ctx context.Context, offset, limit int) ([]model.AuditLog, int64, error) {
	return s.repo.FindAll(ctx, offset, limit)
}
