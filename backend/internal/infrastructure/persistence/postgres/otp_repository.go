package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTPRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) *OTPRepository {
	return &OTPRepository{db: db}
}

func (r *OTPRepository) Create(ctx context.Context, otp *model.OTPCode) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *OTPRepository) FindLatestByEmail(ctx context.Context, email string) (*model.OTPCode, error) {
	var otp model.OTPCode
	if err := r.db.WithContext(ctx).
		Where("email = ? AND used = false", email).
		Order("created_at DESC").
		First(&otp).Error; err != nil {
		return nil, fmt.Errorf("OTP not found: %w", err)
	}
	return &otp, nil
}

func (r *OTPRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.OTPCode{}).Where("id = ?", id).Update("used", true).Error
}

func (r *OTPRepository) DeleteExpiredByEmail(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).Where("email = ? AND used = false", email).Delete(&model.OTPCode{}).Error
}
