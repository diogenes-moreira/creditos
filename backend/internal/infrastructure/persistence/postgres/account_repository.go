package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *model.CurrentAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *AccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.CurrentAccount, error) {
	var account model.CurrentAccount
	if err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	return &account, nil
}

func (r *AccountRepository) FindByClientID(ctx context.Context, clientID uuid.UUID) (*model.CurrentAccount, error) {
	var account model.CurrentAccount
	if err := r.db.WithContext(ctx).First(&account, "client_id = ?", clientID).Error; err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	return &account, nil
}

func (r *AccountRepository) Update(ctx context.Context, account *model.CurrentAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}
