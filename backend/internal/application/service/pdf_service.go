package service

import (
	"context"
	"fmt"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/google/uuid"
)

type PDFAppService struct {
	pdfService  port.PDFService
	storage     port.StorageService
	loanRepo    port.LoanRepository
	clientRepo  port.ClientRepository
	paymentRepo port.PaymentRepository
}

func NewPDFAppService(
	pdfService port.PDFService,
	storage port.StorageService,
	loanRepo port.LoanRepository,
	clientRepo port.ClientRepository,
	paymentRepo port.PaymentRepository,
) *PDFAppService {
	return &PDFAppService{
		pdfService:  pdfService,
		storage:     storage,
		loanRepo:    loanRepo,
		clientRepo:  clientRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *PDFAppService) GenerateLoanSchedulePDF(ctx context.Context, loanID uuid.UUID) (string, error) {
	loan, err := s.loanRepo.FindByIDWithInstallments(ctx, loanID)
	if err != nil {
		return "", err
	}
	client, err := s.clientRepo.FindByID(ctx, loan.ClientID)
	if err != nil {
		return "", err
	}

	reader, err := s.pdfService.GenerateLoanSchedule(loan, loan.Installments, client)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	path := fmt.Sprintf("pdfs/loan-schedule-%s.pdf", loanID.String())
	url, err := s.storage.Upload(ctx, path, reader, "application/pdf")
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}
	return url, nil
}

func (s *PDFAppService) GeneratePaymentReceiptPDF(ctx context.Context, paymentID uuid.UUID) (string, error) {
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return "", err
	}
	loan, err := s.loanRepo.FindByID(ctx, payment.LoanID)
	if err != nil {
		return "", err
	}
	client, err := s.clientRepo.FindByID(ctx, loan.ClientID)
	if err != nil {
		return "", err
	}

	reader, err := s.pdfService.GeneratePaymentReceipt(payment, loan, client)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	path := fmt.Sprintf("pdfs/payment-receipt-%s.pdf", paymentID.String())
	url, err := s.storage.Upload(ctx, path, reader, "application/pdf")
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}
	return url, nil
}
