package postgres

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Create(ctx context.Context, client *model.Client) error {
	return r.db.WithContext(ctx).Create(client).Error
}

func (r *ClientRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Client, error) {
	var client model.Client
	if err := r.db.WithContext(ctx).Preload("User").First(&client, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}
	return &client, nil
}

func (r *ClientRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Client, error) {
	var client model.Client
	if err := r.db.WithContext(ctx).Preload("User").First(&client, "user_id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}
	return &client, nil
}

func (r *ClientRepository) FindByDNI(ctx context.Context, dni string) (*model.Client, error) {
	var client model.Client
	if err := r.db.WithContext(ctx).First(&client, "dni = ?", dni).Error; err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}
	return &client, nil
}

func (r *ClientRepository) FindByCUIT(ctx context.Context, cuit string) (*model.Client, error) {
	var client model.Client
	if err := r.db.WithContext(ctx).First(&client, "cuit = ?", cuit).Error; err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}
	return &client, nil
}

func (r *ClientRepository) Search(ctx context.Context, query string, offset, limit int) ([]model.Client, int64, error) {
	var clients []model.Client
	var total int64
	q := "%" + query + "%"
	base := r.db.WithContext(ctx).Model(&model.Client{}).
		Where("first_name ILIKE ? OR last_name ILIKE ? OR dni ILIKE ? OR cuit ILIKE ? OR phone ILIKE ?", q, q, q, q, q)
	base.Count(&total)
	if err := base.Preload("User").Offset(offset).Limit(limit).Order("created_at DESC").Find(&clients).Error; err != nil {
		return nil, 0, err
	}
	return clients, total, nil
}

func (r *ClientRepository) FindAll(ctx context.Context, offset, limit int) ([]model.Client, int64, error) {
	var clients []model.Client
	var total int64
	r.db.WithContext(ctx).Model(&model.Client{}).Count(&total)
	if err := r.db.WithContext(ctx).Preload("User").Offset(offset).Limit(limit).Order("created_at DESC").Find(&clients).Error; err != nil {
		return nil, 0, err
	}
	return clients, total, nil
}

func (r *ClientRepository) Update(ctx context.Context, client *model.Client) error {
	return r.db.WithContext(ctx).Save(client).Error
}
