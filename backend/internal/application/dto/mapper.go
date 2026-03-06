package dto

import (
	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
)

func ToUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:    u.ID.String(),
		Email: u.Email,
		Role:  string(u.Role),
	}
}

func ToClientResponse(c *model.Client, email string) ClientResponse {
	return ClientResponse{
		ID:              c.ID.String(),
		FirstName:       c.FirstName,
		LastName:        c.LastName,
		DNI:             c.DNI,
		CUIT:            c.CUIT,
		DateOfBirth:     c.DateOfBirth.Format("2006-01-02"),
		Phone:           c.Phone,
		Address:         c.Address,
		City:            c.City,
		Province:        c.Province,
		IsPEP:           c.IsPEP,
		IVARate:         c.IVARate.StringFixed(2),
		MercadoPagoLink: c.MercadoPagoLink,
		IsBlocked:       c.IsBlocked,
		Email:           email,
		CreatedAt:       c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToAccountResponse(a *model.CurrentAccount) AccountResponse {
	return AccountResponse{
		ID:       a.ID.String(),
		Balance:  a.Balance.StringFixed(2),
		ClientID: a.ClientID.String(),
	}
}

func ToMovementResponse(m *model.Movement) MovementResponse {
	return MovementResponse{
		ID:           m.ID.String(),
		Type:         string(m.Type),
		Amount:       m.Amount.StringFixed(2),
		BalanceAfter: m.BalanceAfter.StringFixed(2),
		Description:  m.Description,
		Reference:    m.Reference,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToMovementResponses(movements []model.Movement) []MovementResponse {
	result := make([]MovementResponse, len(movements))
	for i, m := range movements {
		result[i] = ToMovementResponse(&m)
	}
	return result
}

func ToCreditLineResponse(cl *model.CreditLine) CreditLineResponse {
	resp := CreditLineResponse{
		ID:                  cl.ID.String(),
		ClientID:            cl.ClientID.String(),
		MaxAmount:           cl.MaxAmount.StringFixed(2),
		UsedAmount:          cl.UsedAmount.StringFixed(2),
		AvailableAmount:     cl.AvailableAmount().StringFixed(2),
		InterestRate:        cl.InterestRate.StringFixed(4),
		MaxInstallments:     cl.MaxInstallments,
		RecalculateOnPrepay: cl.RecalculateOnPrepay,
		Status:              string(cl.Status),
		RejectionReason:     cl.RejectionReason,
		CreatedAt:           cl.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if cl.ApprovedAt != nil {
		s := cl.ApprovedAt.Format("2006-01-02T15:04:05Z")
		resp.ApprovedAt = &s
	}
	return resp
}

func ToCreditLineResponses(cls []model.CreditLine) []CreditLineResponse {
	result := make([]CreditLineResponse, len(cls))
	for i, cl := range cls {
		result[i] = ToCreditLineResponse(&cl)
	}
	return result
}

func ToLoanResponse(l *model.Loan) LoanResponse {
	clientName := ""
	if l.Client.FirstName != "" || l.Client.LastName != "" {
		clientName = l.Client.FirstName + " " + l.Client.LastName
	}
	resp := LoanResponse{
		ID:               l.ID.String(),
		ClientID:         l.ClientID.String(),
		ClientName:       clientName,
		CreditLineID:     l.CreditLineID.String(),
		Principal:        l.Principal.StringFixed(2),
		InterestRate:     l.InterestRate.StringFixed(4),
		NumInstallments:  l.NumInstallments,
		AmortizationType: string(l.AmortizationType),
		Status:           string(l.Status),
		TotalPaid:        l.TotalPaid().StringFixed(2),
		TotalRemaining:   l.TotalRemaining().StringFixed(2),
		CreatedAt:        l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if l.DisbursedAt != nil {
		s := l.DisbursedAt.Format("2006-01-02T15:04:05Z")
		resp.DisbursedAt = &s
	}
	if l.Status == model.LoanActive {
		pc, ai, aiva, total := l.CancellationSettlement()
		resp.CancellationSettlement = &CancellationSettlementResponse{
			PendingCapital:      pc.StringFixed(2),
			AccumulatedInterest: ai.StringFixed(2),
			AccumulatedIVA:      aiva.StringFixed(2),
			Total:               total.StringFixed(2),
		}
	}
	resp.Installments = ToInstallmentResponses(l.Installments)
	return resp
}

func ToLoanResponses(loans []model.Loan) []LoanResponse {
	result := make([]LoanResponse, len(loans))
	for i, l := range loans {
		result[i] = ToLoanResponse(&l)
	}
	return result
}

func ToInstallmentResponse(inst *model.Installment) InstallmentResponse {
	resp := InstallmentResponse{
		ID:              inst.ID.String(),
		Number:          inst.Number,
		DueDate:         inst.DueDate.Format("2006-01-02"),
		CapitalAmount:   inst.CapitalAmount.StringFixed(2),
		InterestAmount:  inst.InterestAmount.StringFixed(2),
		IVAAmount:       inst.IVAAmount.StringFixed(2),
		TotalAmount:     inst.TotalAmount.StringFixed(2),
		PaidAmount:      inst.PaidAmount.StringFixed(2),
		RemainingAmount: inst.RemainingAmount.StringFixed(2),
		PenaltyApplied:  inst.PenaltyApplied,
		Status:          string(inst.Status),
	}
	if inst.PaidAt != nil {
		s := inst.PaidAt.Format("2006-01-02T15:04:05Z")
		resp.PaidAt = &s
	}
	return resp
}

func ToInstallmentResponses(installments []model.Installment) []InstallmentResponse {
	result := make([]InstallmentResponse, len(installments))
	for i, inst := range installments {
		result[i] = ToInstallmentResponse(&inst)
	}
	return result
}

func ToPaymentResponse(p *model.Payment) PaymentResponse {
	resp := PaymentResponse{
		ID:             p.ID.String(),
		LoanID:         p.LoanID.String(),
		Amount:         p.Amount.StringFixed(2),
		Method:         string(p.Method),
		Reference:      p.Reference,
		IsAdjustment:   p.IsAdjustment,
		AdjustmentNote: p.AdjustmentNote,
		CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if p.InstallmentID != nil {
		s := p.InstallmentID.String()
		resp.InstallmentID = &s
	}
	return resp
}

func ToPaymentResponses(payments []model.Payment) []PaymentResponse {
	result := make([]PaymentResponse, len(payments))
	for i, p := range payments {
		result[i] = ToPaymentResponse(&p)
	}
	return result
}

func ToAuditLogResponse(a *model.AuditLog) AuditLogResponse {
	resp := AuditLogResponse{
		ID:          a.ID.String(),
		Action:      a.Action,
		EntityType:  a.EntityType,
		EntityID:    a.EntityID,
		Description: a.Description,
		IP:          a.IP,
		UserAgent:   a.UserAgent,
		CreatedAt:   a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if a.UserID != nil {
		s := a.UserID.String()
		resp.UserID = &s
	}
	return resp
}

func ToAuditLogResponses(logs []model.AuditLog) []AuditLogResponse {
	result := make([]AuditLogResponse, len(logs))
	for i, l := range logs {
		result[i] = ToAuditLogResponse(&l)
	}
	return result
}

func ToVendorResponse(v *model.Vendor, email string) VendorResponse {
	return VendorResponse{
		ID:           v.ID.String(),
		BusinessName: v.BusinessName,
		CUIT:         v.CUIT,
		Phone:        v.Phone,
		Address:      v.Address,
		City:         v.City,
		Province:     v.Province,
		IsActive:     v.IsActive,
		Email:        email,
		CreatedAt:    v.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToVendorAccountResponse(a *model.VendorAccount) VendorAccountResponse {
	return VendorAccountResponse{
		ID:       a.ID.String(),
		VendorID: a.VendorID.String(),
		Balance:  a.Balance.StringFixed(2),
	}
}

func ToVendorMovementResponse(m *model.VendorMovement) VendorMovementResponse {
	return VendorMovementResponse{
		ID:           m.ID.String(),
		Type:         string(m.Type),
		Amount:       m.Amount.StringFixed(2),
		BalanceAfter: m.BalanceAfter.StringFixed(2),
		Description:  m.Description,
		Reference:    m.Reference,
		CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToVendorMovementResponses(movements []model.VendorMovement) []VendorMovementResponse {
	result := make([]VendorMovementResponse, len(movements))
	for i, m := range movements {
		result[i] = ToVendorMovementResponse(&m)
	}
	return result
}

func ToPurchaseResponse(p *model.Purchase) PurchaseResponse {
	resp := PurchaseResponse{
		ID:           p.ID.String(),
		VendorID:     p.VendorID.String(),
		ClientID:     p.ClientID.String(),
		CreditLineID: p.CreditLineID.String(),
		LoanID:       p.LoanID.String(),
		Amount:       p.Amount.StringFixed(2),
		Description:  p.Description,
		CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if p.Vendor.BusinessName != "" {
		resp.VendorName = p.Vendor.BusinessName
	}
	if p.Client.FirstName != "" {
		resp.ClientName = p.Client.FullName()
	}
	return resp
}

func ToPurchaseResponses(purchases []model.Purchase) []PurchaseResponse {
	result := make([]PurchaseResponse, len(purchases))
	for i, p := range purchases {
		result[i] = ToPurchaseResponse(&p)
	}
	return result
}

func ToWithdrawalRequestResponse(w *model.WithdrawalRequest) WithdrawalRequestResponse {
	resp := WithdrawalRequestResponse{
		ID:              w.ID.String(),
		VendorID:        w.VendorID.String(),
		Amount:          w.Amount.StringFixed(2),
		Method:          string(w.Method),
		Reference:       w.Reference,
		Status:          string(w.Status),
		RejectionReason: w.RejectionReason,
		RequestedAt:     w.RequestedAt.Format("2006-01-02T15:04:05Z"),
	}
	if w.Vendor.BusinessName != "" {
		resp.VendorName = w.Vendor.BusinessName
	}
	if w.ProcessedAt != nil {
		s := w.ProcessedAt.Format("2006-01-02T15:04:05Z")
		resp.ProcessedAt = &s
	}
	if w.PaymentID != nil {
		s := w.PaymentID.String()
		resp.PaymentID = &s
	}
	return resp
}

func ToWithdrawalRequestResponses(requests []model.WithdrawalRequest) []WithdrawalRequestResponse {
	result := make([]WithdrawalRequestResponse, len(requests))
	for i, w := range requests {
		result[i] = ToWithdrawalRequestResponse(&w)
	}
	return result
}

func ToVendorPaymentResponse(p *model.VendorPayment) VendorPaymentResponse {
	return VendorPaymentResponse{
		ID:        p.ID.String(),
		VendorID:  p.VendorID.String(),
		Amount:    p.Amount.StringFixed(2),
		Method:    string(p.Method),
		Reference: p.Reference,
		PaidBy:    p.PaidBy.String(),
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToVendorPaymentResponses(payments []model.VendorPayment) []VendorPaymentResponse {
	result := make([]VendorPaymentResponse, len(payments))
	for i, p := range payments {
		result[i] = ToVendorPaymentResponse(&p)
	}
	return result
}

func ToPortfolioResponse(p *PortfolioData) PortfolioResponse {
	return PortfolioResponse{
		TotalClients:     p.TotalClients,
		ActiveLoans:      p.ActiveLoans,
		TotalDisbursed:   p.TotalDisbursed,
		TotalOutstanding: p.TotalOutstanding,
		TotalCollected:   p.TotalCollected,
		PendingApprovals: p.PendingApprovals,
	}
}

type PortfolioData struct {
	TotalClients     int64
	ActiveLoans      int64
	TotalDisbursed   string
	TotalOutstanding string
	TotalCollected   string
	PendingApprovals int64
}
