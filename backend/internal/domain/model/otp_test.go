package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOTPCode(t *testing.T) {
	otp, err := NewOTPCode("test@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, otp.ID)
	assert.Equal(t, "test@example.com", otp.Email)
	assert.Len(t, otp.Code, 6)
	assert.False(t, otp.Used)
	assert.True(t, otp.ExpiresAt.After(time.Now()))
}

func TestNewOTPCode_EmptyEmail(t *testing.T) {
	_, err := NewOTPCode("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestOTPCode_IsExpired(t *testing.T) {
	otp, _ := NewOTPCode("test@example.com")
	assert.False(t, otp.IsExpired())

	otp.ExpiresAt = time.Now().Add(-1 * time.Minute)
	assert.True(t, otp.IsExpired())
}

func TestOTPCode_IsValid(t *testing.T) {
	otp, _ := NewOTPCode("test@example.com")

	assert.True(t, otp.IsValid(otp.Code))
	assert.False(t, otp.IsValid("000000"))

	// Expired OTP
	otp.ExpiresAt = time.Now().Add(-1 * time.Minute)
	assert.False(t, otp.IsValid(otp.Code))

	// Used OTP
	otp2, _ := NewOTPCode("test@example.com")
	otp2.MarkUsed()
	assert.False(t, otp2.IsValid(otp2.Code))
}

func TestOTPCode_MarkUsed(t *testing.T) {
	otp, _ := NewOTPCode("test@example.com")
	assert.False(t, otp.Used)
	otp.MarkUsed()
	assert.True(t, otp.Used)
}

func TestOTPCode_UniqueCode(t *testing.T) {
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		otp, err := NewOTPCode("test@example.com")
		require.NoError(t, err)
		codes[otp.Code] = true
	}
	// With 6-digit codes, 100 samples should produce at least a few unique codes
	assert.Greater(t, len(codes), 1)
}
