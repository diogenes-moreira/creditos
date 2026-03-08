package service

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/messaging"
)

type OTPService struct {
	otpRepo   port.OTPRepository
	userRepo  port.UserRepository
	sender    messaging.OTPSender
	audit     *AuditService
}

func NewOTPService(
	otpRepo port.OTPRepository,
	userRepo port.UserRepository,
	sender messaging.OTPSender,
	audit *AuditService,
) *OTPService {
	return &OTPService{
		otpRepo:  otpRepo,
		userRepo: userRepo,
		sender:   sender,
		audit:    audit,
	}
}

func (s *OTPService) RequestOTP(ctx context.Context, email string) error {
	// Verify the user exists and is a client
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.Role != model.RoleClient && user.Role != model.RoleVendor {
		return fmt.Errorf("OTP login is only available for clients and vendors")
	}

	if !user.IsActive {
		return fmt.Errorf("account is deactivated")
	}

	// Invalidate previous unused OTPs
	_ = s.otpRepo.DeleteExpiredByEmail(ctx, email)

	// Generate new OTP
	otp, err := model.NewOTPCode(email)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.otpRepo.Create(ctx, otp); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Send OTP via messaging service
	if err := s.sender.SendOTP(ctx, email, otp.Code); err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}

	s.audit.Record(ctx, &user.ID, "request_otp", "user", user.ID.String(), fmt.Sprintf("OTP requested for %s", email))

	return nil
}

func (s *OTPService) VerifyOTP(ctx context.Context, email, code string) (*model.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	otp, err := s.otpRepo.FindLatestByEmail(ctx, email)
	if err != nil {
		s.audit.Record(ctx, &user.ID, "otp_verify_failed", "user", user.ID.String(), "No OTP found")
		return nil, fmt.Errorf("invalid or expired code")
	}

	if !otp.IsValid(code) {
		s.audit.Record(ctx, &user.ID, "otp_verify_failed", "user", user.ID.String(), "Invalid OTP code")
		return nil, fmt.Errorf("invalid or expired code")
	}

	// Mark OTP as used
	otp.MarkUsed()
	if err := s.otpRepo.MarkUsed(ctx, otp.ID); err != nil {
		return nil, fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	// Record login
	user.RecordLogin()
	_ = s.userRepo.Update(ctx, user)

	s.audit.Record(ctx, &user.ID, "otp_login", "user", user.ID.String(), fmt.Sprintf("Client %s logged in via OTP", email))

	return user, nil
}
