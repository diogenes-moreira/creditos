package port

import (
	"io"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
)

type PDFService interface {
	GenerateLoanSchedule(loan *model.Loan, installments []model.Installment, client *model.Client) (io.Reader, error)
	GeneratePaymentReceipt(payment *model.Payment, loan *model.Loan, client *model.Client) (io.Reader, error)
	GenerateVendorPaymentReceipt(payment *model.VendorPayment, vendor *model.Vendor) (io.Reader, error)
}
