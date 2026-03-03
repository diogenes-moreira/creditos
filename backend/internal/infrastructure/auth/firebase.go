package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// LocalAuthService implements AuthService for local development without Firebase.
type LocalAuthService struct {
	jwtSecret       string
	expirationHours int
}

func NewLocalAuthService(jwtSecret string, expirationHours int) *LocalAuthService {
	return &LocalAuthService{
		jwtSecret:       jwtSecret,
		expirationHours: expirationHours,
	}
}

func (s *LocalAuthService) CreateUser(_ context.Context, email, password string) (*port.FirebaseUser, error) {
	_ = password
	return &port.FirebaseUser{
		UID:   uuid.New().String(),
		Email: email,
	}, nil
}

func (s *LocalAuthService) VerifyToken(_ context.Context, tokenStr string) (*port.FirebaseUser, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	uid, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	return &port.FirebaseUser{UID: uid, Email: email}, nil
}

func (s *LocalAuthService) DeleteUser(_ context.Context, uid string) error {
	_ = uid
	return nil
}

func (s *LocalAuthService) GenerateToken(userID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(time.Duration(s.expirationHours) * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
