package auth

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	fbauth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// FirebaseVerifiedToken holds the decoded claims from a Firebase ID token.
type FirebaseVerifiedToken struct {
	UID            string
	Email          string
	Phone          string
	SignInProvider  string
}

// FirebaseTokenVerifier verifies Firebase ID tokens using the Admin SDK.
type FirebaseTokenVerifier struct {
	client *fbauth.Client
}

// NewFirebaseTokenVerifier creates a verifier from a credentials JSON file path.
func NewFirebaseTokenVerifier(credentialsFile string) (*FirebaseTokenVerifier, error) {
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase auth client: %w", err)
	}
	return &FirebaseTokenVerifier{client: client}, nil
}

// NewFirebaseTokenVerifierOrNil creates a verifier if FIREBASE_CREDENTIALS_FILE is set,
// otherwise returns nil (dev mode).
func NewFirebaseTokenVerifierOrNil() *FirebaseTokenVerifier {
	credsFile := os.Getenv("FIREBASE_CREDENTIALS_FILE")
	if credsFile == "" {
		log.Println("FIREBASE_CREDENTIALS_FILE not set — Firebase token verification disabled (dev mode)")
		return nil
	}
	v, err := NewFirebaseTokenVerifier(credsFile)
	if err != nil {
		log.Printf("WARNING: Failed to init Firebase verifier: %v — running without it", err)
		return nil
	}
	log.Println("Firebase token verifier initialized")
	return v
}

// VerifyIDToken verifies a Firebase ID token and extracts claims.
func (v *FirebaseTokenVerifier) VerifyIDToken(ctx context.Context, idToken string) (*FirebaseVerifiedToken, error) {
	token, err := v.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid Firebase ID token: %w", err)
	}

	result := &FirebaseVerifiedToken{
		UID: token.UID,
	}

	if email, ok := token.Claims["email"].(string); ok {
		result.Email = email
	}
	if phone, ok := token.Claims["phone_number"].(string); ok {
		result.Phone = phone
	}
	result.SignInProvider = token.Firebase.SignInProvider

	return result, nil
}
