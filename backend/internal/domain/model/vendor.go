package model

import (
	"fmt"
	"time"

	"github.com/diogenes-moreira/creditos/backend/pkg/validator"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Vendor struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null"`
	User         User           `gorm:"foreignKey:UserID"`
	BusinessName string         `gorm:"not null"`
	CUIT         string         `gorm:"uniqueIndex;not null"`
	Phone        string         `gorm:"not null"`
	Address      string         `gorm:"not null"`
	City         string         `gorm:"not null"`
	Province     string         `gorm:"not null"`
	Country      string         `gorm:"not null;default:'Argentina'"`
	IsActive     bool           `gorm:"not null;default:true"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func NewVendor(userID uuid.UUID, businessName, cuit, phone, address, city, province, country string) (*Vendor, error) {
	if businessName == "" {
		return nil, fmt.Errorf("business name is required")
	}

	if err := validator.ValidateCUIT(cuit); err != nil {
		return nil, fmt.Errorf("invalid CUIT: %w", err)
	}

	if phone == "" {
		return nil, fmt.Errorf("phone is required")
	}

	return &Vendor{
		ID:           uuid.New(),
		UserID:       userID,
		BusinessName: businessName,
		CUIT:         cuit,
		Phone:        phone,
		Address:      address,
		City:         city,
		Province:     province,
		Country:      country,
		IsActive:     true,
	}, nil
}

func (v *Vendor) Activate() {
	v.IsActive = true
}

func (v *Vendor) Deactivate() {
	v.IsActive = false
}

func (v *Vendor) UpdateProfile(phone, address, city, province, country string) {
	if phone != "" {
		v.Phone = phone
	}
	if address != "" {
		v.Address = address
	}
	if city != "" {
		v.City = city
	}
	if province != "" {
		v.Province = province
	}
	if country != "" {
		v.Country = country
	}
}
