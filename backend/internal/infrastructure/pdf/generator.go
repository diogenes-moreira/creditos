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
	pdf.Cell(190, 10, "Prestia - Plan de Cuotas")
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 7, fmt.Sprintf("Cliente: %s", client.FullName()))
	pdf.Cell(95, 7, fmt.Sprintf("DNI: %s", client.DNI))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("CUIT: %s", client.CUIT))
	pdf.Cell(95, 7, fmt.Sprintf("Prestamo: %s", loan.ID.String()[:8]))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("Capital: $%s", loan.Principal.StringFixed(2)))
	pdf.Cell(95, 7, fmt.Sprintf("Tasa: %s%%", loan.InterestRate.StringFixed(2)))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("Cuotas: %d", loan.NumInstallments))
	pdf.Cell(95, 7, fmt.Sprintf("Sistema: %s", loan.AmortizationType))
	pdf.Ln(12)

	// Status labels
	statusLabels := map[string]string{
		"pending": "Pend.",
		"partial": "Parcial",
		"paid":    "Pagada",
		"overdue": "Vencida",
	}

	// Table header
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(12, 8, "#", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Vencimiento", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 8, "Capital", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 8, "Interes", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 8, "IVA", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 8, "Total", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 8, "Pagado", "1", 0, "C", true, 0, "")
	pdf.CellFormat(19, 8, "Estado", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)
	for _, inst := range installments {
		status := string(inst.Status)
		if label, ok := statusLabels[status]; ok {
			status = label
		}
		pdf.CellFormat(12, 6, fmt.Sprintf("%d", inst.Number), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, inst.DueDate.Format("02/01/2006"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(28, 6, "$"+inst.CapitalAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(28, 6, "$"+inst.InterestAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(22, 6, "$"+inst.IVAAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(28, 6, "$"+inst.TotalAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(28, 6, "$"+inst.PaidAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(19, 6, status, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(190, 7, "Este documento es informativo y no constituye un contrato.")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return &buf, nil
}

func (g *Generator) GeneratePaymentReceipt(payment *model.Payment, loan *model.Loan, client *model.Client) (io.Reader, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Prestia - Comprobante de Pago")
	pdf.Ln(15)

	// Receipt info
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 7, fmt.Sprintf("Comprobante #: %s", payment.ID.String()[:8]))
	pdf.Cell(95, 7, fmt.Sprintf("Fecha: %s", payment.CreatedAt.Format("02/01/2006 15:04")))
	pdf.Ln(10)

	// Client info
	pdf.Cell(95, 7, fmt.Sprintf("Cliente: %s", client.FullName()))
	pdf.Cell(95, 7, fmt.Sprintf("DNI: %s", client.DNI))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("CUIT: %s", client.CUIT))
	pdf.Cell(95, 7, fmt.Sprintf("Prestamo: %s", loan.ID.String()[:8]))
	pdf.Ln(10)

	// Payment details
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 7, fmt.Sprintf("Metodo: %s", payment.Method))
	if payment.Reference != "" {
		pdf.Cell(95, 7, fmt.Sprintf("Referencia: %s", payment.Reference))
	}
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, fmt.Sprintf("Monto Pagado: $%s", payment.Amount.StringFixed(2)))
	pdf.Ln(15)

	// Loan summary
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Plan de Cuotas")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(95, 6, fmt.Sprintf("Capital: $%s", loan.Principal.StringFixed(2)))
	pdf.Cell(95, 6, fmt.Sprintf("Tasa: %s%%", loan.InterestRate.StringFixed(2)))
	pdf.Ln(6)
	pdf.Cell(95, 6, fmt.Sprintf("Cuotas: %d", loan.NumInstallments))
	pdf.Cell(95, 6, fmt.Sprintf("Sistema: %s", loan.AmortizationType))
	pdf.Ln(8)

	// Installments table header
	pdf.SetFont("Arial", "B", 8)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(12, 7, "#", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Vencimiento", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 7, "Capital", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 7, "Interes", "1", 0, "C", true, 0, "")
	pdf.CellFormat(22, 7, "IVA", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 7, "Total", "1", 0, "C", true, 0, "")
	pdf.CellFormat(28, 7, "Pagado", "1", 0, "C", true, 0, "")
	pdf.CellFormat(19, 7, "Estado", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	// Installments table rows
	pdf.SetFont("Arial", "", 8)
	statusLabels := map[string]string{
		"pending": "Pend.",
		"partial": "Parcial",
		"paid":    "Pagada",
		"overdue": "Vencida",
	}
	for _, inst := range loan.Installments {
		status := string(inst.Status)
		if label, ok := statusLabels[status]; ok {
			status = label
		}
		pdf.CellFormat(12, 6, fmt.Sprintf("%d", inst.Number), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, inst.DueDate.Format("02/01/2006"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(28, 6, "$"+inst.CapitalAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(28, 6, "$"+inst.InterestAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(22, 6, "$"+inst.IVAAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(28, 6, "$"+inst.TotalAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(28, 6, "$"+inst.PaidAmount.StringFixed(2), "1", 0, "R", false, 0, "")
		pdf.CellFormat(19, 6, status, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(190, 7, "Este comprobante es valido como constancia de pago.")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return &buf, nil
}

func (g *Generator) GenerateCancellationReceipt(payment *model.Payment, loan *model.Loan, client *model.Client, pendingCapital, accruedInterest, accruedIVA, total string) (io.Reader, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Prestia - Comprobante de Cancelacion Anticipada")
	pdf.Ln(15)

	// Receipt info
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 7, fmt.Sprintf("Comprobante #: %s", payment.ID.String()[:8]))
	pdf.Cell(95, 7, fmt.Sprintf("Fecha: %s", payment.CreatedAt.Format("02/01/2006 15:04")))
	pdf.Ln(10)

	// Client info
	pdf.Cell(95, 7, fmt.Sprintf("Cliente: %s", client.FullName()))
	pdf.Cell(95, 7, fmt.Sprintf("DNI: %s", client.DNI))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("CUIT: %s", client.CUIT))
	pdf.Cell(95, 7, fmt.Sprintf("Prestamo: %s", loan.ID.String()[:8]))
	pdf.Ln(10)

	// Loan info
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(95, 7, fmt.Sprintf("Capital original: $%s", loan.Principal.StringFixed(2)))
	pdf.Cell(95, 7, fmt.Sprintf("Tasa: %s%%", loan.InterestRate.StringFixed(2)))
	pdf.Ln(7)
	pdf.Cell(95, 7, fmt.Sprintf("Cuotas: %d", loan.NumInstallments))
	pdf.Cell(95, 7, fmt.Sprintf("Sistema: %s", loan.AmortizationType))
	pdf.Ln(7)
	if loan.DisbursedAt != nil {
		pdf.Cell(190, 7, fmt.Sprintf("Fecha de desembolso: %s", loan.DisbursedAt.Format("02/01/2006")))
		pdf.Ln(7)
	}
	pdf.Ln(5)

	// Settlement breakdown
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "Detalle de Cancelacion")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(95, 8, "Concepto", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 8, "Monto", "1", 0, "R", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(95, 8, "Capital pendiente", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, "$"+pendingCapital, "1", 0, "R", false, 0, "")
	pdf.Ln(-1)
	pdf.CellFormat(95, 8, "Interes devengado", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, "$"+accruedInterest, "1", 0, "R", false, 0, "")
	pdf.Ln(-1)
	pdf.CellFormat(95, 8, "IVA", "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 8, "$"+accruedIVA, "1", 0, "R", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(95, 9, "TOTAL CANCELACION", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 9, "$"+total, "1", 0, "R", true, 0, "")
	pdf.Ln(-1)

	pdf.Ln(7)
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(95, 7, fmt.Sprintf("Metodo de pago: %s", payment.Method))
	if payment.Reference != "" {
		pdf.Cell(95, 7, fmt.Sprintf("Referencia: %s", payment.Reference))
	}
	pdf.Ln(15)

	// Footer
	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(190, 7, "Este comprobante es valido como constancia de pago por cancelacion anticipada del credito.")

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
