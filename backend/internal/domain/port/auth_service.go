package port

import "context"

type FirebaseUser struct {
	UID   string
	Email string
}

type AuthService interface {
	CreateUser(ctx context.Context, email, password string) (*FirebaseUser, error)
	VerifyToken(ctx context.Context, token string) (*FirebaseUser, error)
	DeleteUser(ctx context.Context, uid string) error
}
