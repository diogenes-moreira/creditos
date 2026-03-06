package model

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

const otpExpiration = 5 * time.Minute

type OTPCode struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"not null;index"`
	Code      string    `gorm:"type:varchar(6);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"not null"`
}

func NewOTPCode(email string) (*OTPCode, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	code, err := generateOTPCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP code: %w", err)
	}

	return &OTPCode{
		ID:        uuid.New(),
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(otpExpiration),
		Used:      false,
	}, nil
}

func (o *OTPCode) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

func (o *OTPCode) IsValid(code string) bool {
	return !o.Used && !o.IsExpired() && o.Code == code
}

func (o *OTPCode) MarkUsed() {
	o.Used = true
}

func generateOTPCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}
