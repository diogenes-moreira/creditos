package pdf

import (
	"bytes"
	"fmt"
	"io"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/jung-kurt/gofpdf"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateLoanSchedule(loan *model.Loan, installments []model.Installment, client *model.Client) (io.Reader, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Prestia - Loan Schedule")
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 7, fmt.Sprintf("Client: %s", client.FullName()))
	pdf.Cell(95, 7, fmt.Sprintf("DNI: %s", client.DNI))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("CUIT: %s", client.CUIT))
	pdf.Cell(95, 7, fmt.Sprintf("Loan ID: %s", loan.ID.String()[:8]))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("Principal: $%s", loan.Principal.StringFixed(2)))
	pdf.Cell(95, 7, fmt.Sprintf("Rate: %s%%", loan.InterestRate.StringFixed(2)))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("Installments: %d", loan.NumInstallments))
	pdf.Cell(95, 7, fmt.Sprintf("Type: %s", loan.AmortizationType))
	pdf.Ln(12)

	// Table header
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(15, 8, "#", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Due Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Capital", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Interest", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Total", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Status", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 9)
	for _, inst := range installments {
		pdf.CellFormat(15, 7, fmt.Sprintf("%d", inst.Number), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 7, inst.DueDate.Format("02/01/2006"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 7, "$"+inst.CapitalAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, "$"+inst.InterestAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, "$"+inst.TotalAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 7, string(inst.Status), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return &buf, nil
}

func (g *Generator) GeneratePaymentReceipt(payment *model.Payment, loan *model.Loan, client *model.Client) (io.Reader, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Prestia - Payment Receipt")
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(190, 7, fmt.Sprintf("Receipt #: %s", payment.ID.String()[:8]))
	pdf.Ln(7)
	pdf.Cell(190, 7, fmt.Sprintf("Date: %s", payment.CreatedAt.Format("02/01/2006 15:04")))
	pdf.Ln(12)

	pdf.Cell(95, 7, fmt.Sprintf("Client: %s", client.FullName()))
	pdf.Cell(95, 7, fmt.Sprintf("DNI: %s", client.DNI))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("Loan: %s", loan.ID.String()[:8]))
	pdf.Cell(95, 7, fmt.Sprintf("Method: %s", payment.Method))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, fmt.Sprintf("Amount Paid: $%s", payment.Amount.StringFixed(2)))
	pdf.Ln(10)

	if payment.Reference != "" {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(190, 7, fmt.Sprintf("Reference: %s", payment.Reference))
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return &buf, nil
}

func (g *Generator) GenerateVendorPaymentReceipt(payment *model.VendorPayment, vendor *model.Vendor) (io.Reader, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Prestia - Comprobante de Pago a Vendedor")
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(190, 7, fmt.Sprintf("Comprobante #: %s", payment.ID.String()[:8]))
	pdf.Ln(7)
	pdf.Cell(190, 7, fmt.Sprintf("Fecha: %s", payment.CreatedAt.Format("02/01/2006 15:04")))
	pdf.Ln(12)

	pdf.Cell(95, 7, fmt.Sprintf("Vendedor: %s", vendor.BusinessName))
	pdf.Cell(95, 7, fmt.Sprintf("CUIT: %s", vendor.CUIT))
	pdf.Ln(7)
	pdf.Cell(190, 7, fmt.Sprintf("Direccion: %s, %s, %s", vendor.Address, vendor.City, vendor.Province))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, fmt.Sprintf("Monto Pagado: $%s", payment.Amount.StringFixed(2)))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(190, 7, fmt.Sprintf("Metodo: %s", string(payment.Method)))
	pdf.Ln(7)

	if payment.Reference != "" {
		pdf.Cell(190, 7, fmt.Sprintf("Referencia: %s", payment.Reference))
		pdf.Ln(7)
	}

	pdf.Ln(20)
	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(190, 7, "Este comprobante es valido como constancia de pago.")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return &buf, nil
}
