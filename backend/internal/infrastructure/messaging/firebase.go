package messaging

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// OTPSender sends OTP codes to users via email.
type OTPSender interface {
	SendOTP(ctx context.Context, email, code string) error
}

// FirebaseOTPSender sends OTP via Firebase Auth email link.
type FirebaseOTPSender struct {
	authClient *auth.Client
}

func NewFirebaseOTPSender(credentialsFile string) (*FirebaseOTPSender, error) {
	if credentialsFile == "" {
		return nil, fmt.Errorf("firebase credentials file not configured")
	}

	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase Auth client: %w", err)
	}

	return &FirebaseOTPSender{authClient: authClient}, nil
}

func (s *FirebaseOTPSender) SendOTP(ctx context.Context, email, code string) error {
	actionCodeSettings := &auth.ActionCodeSettings{
		URL:             os.Getenv("OTP_REDIRECT_URL"),
		HandleCodeInApp: true,
	}

	link, err := s.authClient.EmailSignInLink(ctx, email, actionCodeSettings)
	if err != nil {
		return fmt.Errorf("failed to generate email sign-in link: %w", err)
	}

	// Log the link for debugging; in production Firebase sends the email automatically.
	log.Printf("[Firebase OTP] Email: %s, Code: %s, Link: %s", email, code, link)
	return nil
}

// ConsoleOTPSender logs OTP codes to console for development.
type ConsoleOTPSender struct{}

func NewConsoleOTPSender() *ConsoleOTPSender {
	return &ConsoleOTPSender{}
}

func (s *ConsoleOTPSender) SendOTP(_ context.Context, email, code string) error {
	log.Printf("[OTP] Code for %s: %s", email, code)
	return nil
}

// NewOTPSender creates the appropriate OTP sender based on configuration.
func NewOTPSender() OTPSender {
	credFile := os.Getenv("FIREBASE_CREDENTIALS_FILE")
	if credFile != "" {
		sender, err := NewFirebaseOTPSender(credFile)
		if err != nil {
			log.Printf("Warning: Failed to initialize Firebase OTP sender: %v. Falling back to console.", err)
			return NewConsoleOTPSender()
		}
		return sender
	}
	log.Println("FIREBASE_CREDENTIALS_FILE not set, using console OTP sender")
	return NewConsoleOTPSender()
}
