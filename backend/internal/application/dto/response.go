package dto

import "time"

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ClientResponse struct {
	ID              string `json:"id"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	DNI             string `json:"dni"`
	CUIT            string `json:"cuit"`
	DateOfBirth     string `json:"dateOfBirth"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	City            string `json:"city"`
	Province        string `json:"province"`
	IsPEP           bool   `json:"isPEP"`
	MercadoPagoLink string `json:"mercadoPagoLink,omitempty"`
	IsBlocked       bool   `json:"isBlocked"`
	Email           string `json:"email"`
	CreatedAt       string `json:"createdAt"`
}

type AccountResponse struct {
	ID       string `json:"id"`
	Balance  string `json:"balance"`
	ClientID string `json:"clientId"`
}

type MovementResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Amount       string `json:"amount"`
	BalanceAfter string `json:"balanceAfter"`
	Description  string `json:"description"`
	Reference    string `json:"reference,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

type CreditLineResponse struct {
	ID              string  `json:"id"`
	ClientID        string  `json:"clientId"`
	ClientName      string  `json:"clientName,omitempty"`
	MaxAmount       string  `json:"maxAmount"`
	UsedAmount      string  `json:"usedAmount"`
	AvailableAmount string  `json:"availableAmount"`
	InterestRate    string  `json:"interestRate"`
	MaxInstallments int     `json:"maxInstallments"`
	Status          string  `json:"status"`
	ApprovedAt      *string `json:"approvedAt,omitempty"`
	RejectionReason string  `json:"rejectionReason,omitempty"`
	CreatedAt       string  `json:"createdAt"`
}

type LoanResponse struct {
	ID               string                `json:"id"`
	ClientID         string                `json:"clientId"`
	CreditLineID     string                `json:"creditLineId"`
	Principal        string                `json:"principal"`
	InterestRate     string                `json:"interestRate"`
	NumInstallments  int                   `json:"numInstallments"`
	AmortizationType string                `json:"amortizationType"`
	Status           string                `json:"status"`
	DisbursedAt      *string               `json:"disbursedAt,omitempty"`
	TotalPaid        string                `json:"totalPaid"`
	TotalRemaining   string                `json:"totalRemaining"`
	Installments     []InstallmentResponse `json:"installments,omitempty"`
	CreatedAt        string                `json:"createdAt"`
}

type InstallmentResponse struct {
	ID              string  `json:"id"`
	Number          int     `json:"number"`
	DueDate         string  `json:"dueDate"`
	CapitalAmount   string  `json:"capitalAmount"`
	InterestAmount  string  `json:"interestAmount"`
	TotalAmount     string  `json:"totalAmount"`
	PaidAmount      string  `json:"paidAmount"`
	RemainingAmount string  `json:"remainingAmount"`
	Status          string  `json:"status"`
	PaidAt          *string `json:"paidAt,omitempty"`
}

type SimulationResponse struct {
	Principal     string                `json:"principal"`
	InterestRate  string                `json:"interestRate"`
	TotalInterest string                `json:"totalInterest"`
	TotalPayment  string                `json:"totalPayment"`
	Installments  []InstallmentResponse `json:"installments"`
}

type PaymentResponse struct {
	ID             string  `json:"id"`
	LoanID         string  `json:"loanId"`
	InstallmentID  *string `json:"installmentId,omitempty"`
	Amount         string  `json:"amount"`
	Method         string  `json:"method"`
	Reference      string  `json:"reference,omitempty"`
	IsAdjustment   bool    `json:"isAdjustment"`
	AdjustmentNote string  `json:"adjustmentNote,omitempty"`
	CreatedAt      string  `json:"createdAt"`
}

type AuditLogResponse struct {
	ID          string  `json:"id"`
	UserID      *string `json:"userId,omitempty"`
	Action      string  `json:"action"`
	EntityType  string  `json:"entityType"`
	EntityID    string  `json:"entityId"`
	Description string  `json:"description"`
	IP          string  `json:"ip"`
	UserAgent   string  `json:"userAgent"`
	CreatedAt   string  `json:"createdAt"`
}

type PortfolioResponse struct {
	TotalClients     int64  `json:"totalClients"`
	ActiveLoans      int64  `json:"activeLoans"`
	TotalDisbursed   string `json:"totalDisbursed"`
	TotalOutstanding string `json:"totalOutstanding"`
	TotalCollected   string `json:"totalCollected"`
	PendingApprovals int64  `json:"pendingApprovals"`
}

type DelinquencyResponse struct {
	PAR30           string `json:"par30"`
	PAR60           string `json:"par60"`
	PAR90           string `json:"par90"`
	TotalOverdue    string `json:"totalOverdue"`
	OverdueCount    int64  `json:"overdueCount"`
	DelinquencyRate string `json:"delinquencyRate"`
}

type KPIResponse struct {
	Portfolio   PortfolioResponse   `json:"portfolio"`
	Delinquency DelinquencyResponse `json:"delinquency"`
}

type TrendPointResponse struct {
	Date   time.Time `json:"date"`
	Amount string    `json:"amount"`
	Count  int64     `json:"count"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Offset     int         `json:"offset"`
	Limit      int         `json:"limit"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
