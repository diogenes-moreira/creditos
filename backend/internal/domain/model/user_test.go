package model_test

import (
	"testing"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		firebaseUID string
		email       string
		role        model.UserRole
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid admin user",
			firebaseUID: "firebase-uid-123",
			email:       "admin@example.com",
			role:        model.RoleAdmin,
		},
		{
			name:        "valid client user",
			firebaseUID: "firebase-uid-456",
			email:       "client@example.com",
			role:        model.RoleClient,
		},
		{
			name:        "empty firebase UID",
			firebaseUID: "",
			email:       "test@example.com",
			role:        model.RoleClient,
			wantErr:     true,
			errMsg:      "firebase UID is required",
		},
		{
			name:        "empty email",
			firebaseUID: "firebase-uid-789",
			email:       "",
			role:        model.RoleClient,
			wantErr:     true,
			errMsg:      "email is required",
		},
		{
			name:        "invalid role",
			firebaseUID: "firebase-uid-101",
			email:       "test@example.com",
			role:        model.UserRole("superadmin"),
			wantErr:     true,
			errMsg:      "invalid role: superadmin",
		},
		{
			name:        "empty role string",
			firebaseUID: "firebase-uid-102",
			email:       "test@example.com",
			role:        model.UserRole(""),
			wantErr:     true,
			errMsg:      "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := model.NewUser(tt.firebaseUID, tt.email, tt.role)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.NotEqual(t, user.ID.String(), "00000000-0000-0000-0000-000000000000")
				assert.Equal(t, tt.firebaseUID, user.FirebaseUID)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.role, user.Role)
				assert.True(t, user.IsActive)
				assert.Nil(t, user.LastLoginAt)
			}
		})
	}
}

func TestUser_Activate(t *testing.T) {
	user, err := model.NewUser("uid", "test@example.com", model.RoleClient)
	require.NoError(t, err)

	user.Deactivate()
	assert.False(t, user.IsActive)

	user.Activate()
	assert.True(t, user.IsActive)
}

func TestUser_Deactivate(t *testing.T) {
	user, err := model.NewUser("uid", "test@example.com", model.RoleClient)
	require.NoError(t, err)
	assert.True(t, user.IsActive)

	user.Deactivate()
	assert.False(t, user.IsActive)

	// Deactivating again should stay false
	user.Deactivate()
	assert.False(t, user.IsActive)
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     model.UserRole
		expected bool
	}{
		{name: "admin role returns true", role: model.RoleAdmin, expected: true},
		{name: "client role returns false", role: model.RoleClient, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := model.NewUser("uid", "test@example.com", tt.role)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, user.IsAdmin())
		})
	}
}

func TestUser_RecordLogin(t *testing.T) {
	user, err := model.NewUser("uid", "test@example.com", model.RoleClient)
	require.NoError(t, err)
	assert.Nil(t, user.LastLoginAt)

	user.RecordLogin()
	require.NotNil(t, user.LastLoginAt)
}
