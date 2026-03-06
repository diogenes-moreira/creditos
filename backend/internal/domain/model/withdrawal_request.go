package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WithdrawalStatus string

const (
	WithdrawalStatusPending  WithdrawalStatus = "pending"
	WithdrawalStatusApproved WithdrawalStatus = "approved"
	WithdrawalStatusPaid     WithdrawalStatus = "paid"
	WithdrawalStatusRejected WithdrawalStatus = "rejected"
)

type WithdrawalRequest struct {
	ID              uuid.UUID           `gorm:"type:uuid;primaryKey"`
	VendorID        uuid.UUID           `gorm:"type:uuid;index;not null"`
	Vendor          Vendor              `gorm:"foreignKey:VendorID"`
	Amount          decimal.Decimal     `gorm:"type:decimal(18,2);not null"`
	Method          VendorPaymentMethod `gorm:"type:varchar(20);not null"`
	Reference       string              `gorm:""`
	Status          WithdrawalStatus    `gorm:"type:varchar(20);not null;default:'pending'"`
	RejectionReason string              `gorm:""`
	RequestedAt     time.Time           `gorm:"not null"`
	ProcessedAt     *time.Time          `gorm:""`
	ProcessedBy     *uuid.UUID          `gorm:"type:uuid"`
	PaymentID       *uuid.UUID          `gorm:"type:uuid"`
}

func NewWithdrawalRequest(vendorID uuid.UUID, amount decimal.Decimal, method VendorPaymentMethod) (*WithdrawalRequest, error) {
	if !amount.IsPositive() {
		return nil, fmt.Errorf("withdrawal amount must be positive")
	}
	if method != VendorPaymentCash && method != VendorPaymentTransfer {
		return nil, fmt.Errorf("invalid withdrawal method: %s", method)
	}

	return &WithdrawalRequest{
		ID:          uuid.New(),
		VendorID:    vendorID,
		Amount:      amount,
		Method:      method,
		Status:      WithdrawalStatusPending,
		RequestedAt: time.Now(),
	}, nil
}

func (w *WithdrawalRequest) Approve(adminID uuid.UUID) error {
	if w.Status != WithdrawalStatusPending {
		return fmt.Errorf("can only approve pending withdrawal requests, current status: %s", w.Status)
	}
	now := time.Now()
	w.Status = WithdrawalStatusApproved
	w.ProcessedAt = &now
	w.ProcessedBy = &adminID
	return nil
}

func (w *WithdrawalRequest) MarkPaid(paymentID uuid.UUID, reference string) error {
	if w.Status != WithdrawalStatusApproved {
		return fmt.Errorf("can only mark approved withdrawal requests as paid, current status: %s", w.Status)
	}
	w.Status = WithdrawalStatusPaid
	w.PaymentID = &paymentID
	w.Reference = reference
	return nil
}

func (w *WithdrawalRequest) Reject(adminID uuid.UUID, reason string) error {
	if w.Status != WithdrawalStatusPending {
		return fmt.Errorf("can only reject pending withdrawal requests, current status: %s", w.Status)
	}
	if reason == "" {
		return fmt.Errorf("rejection reason is required")
	}
	now := time.Now()
	w.Status = WithdrawalStatusRejected
	w.RejectionReason = reason
	w.ProcessedAt = &now
	w.ProcessedBy = &adminID
	return nil
}
