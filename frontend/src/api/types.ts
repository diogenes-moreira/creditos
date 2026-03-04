// ---- Auth ----
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  dni: string;
  cuit: string;
  dateOfBirth: string;
  phone: string;
  address: string;
  city: string;
  province: string;
  isPEP: boolean;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface User {
  id: string;
  email: string;
  role: "admin" | "client" | "vendor";
  firstName?: string;
  lastName?: string;
}

// ---- Profile ----
export interface Profile {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  dni: string;
  cuit: string;
  phone: string;
  address: string;
  city: string;
  province: string;
  mercadoPagoLink?: string;
  createdAt: string;
}

export interface UpdateProfileRequest {
  firstName: string;
  lastName: string;
  phone: string;
  address: string;
  city: string;
  province: string;
}

export interface UpdateMercadoPagoRequest {
  alias: string;
  cvu: string;
}

// ---- Account ----
export interface Account {
  id: string;
  clientId: string;
  balance: string;
}

export interface Movement {
  id: string;
  type: string;
  amount: string;
  balanceAfter: string;
  description: string;
  reference?: string;
  createdAt: string;
}

// ---- Credit Lines ----
export interface CreditLine {
  id: string;
  clientId: string;
  clientName?: string;
  maxAmount: string;
  usedAmount: string;
  availableAmount: string;
  interestRate: string;
  maxInstallments: number;
  status: string;
  approvedAt?: string;
  rejectionReason?: string;
  createdAt: string;
}

export interface CreateCreditLineRequest {
  clientId: string;
  maxAmount: string;
  interestRate: string;
  maxInstallments: number;
}

export interface UpdateCreditLineRequest {
  maxAmount: string;
}

// ---- Loans ----
export interface Loan {
  id: string;
  clientId: string;
  clientName?: string;
  creditLineId: string;
  principal: string;
  interestRate: string;
  numInstallments: number;
  amortizationType: string;
  status: string;
  totalPaid: string;
  totalRemaining: string;
  disbursedAt?: string;
  createdAt: string;
}

export interface LoanDetail extends Loan {
  installments: Installment[];
}

export interface Installment {
  id: string;
  number: number;
  dueDate: string;
  capitalAmount: string;
  interestAmount: string;
  totalAmount: string;
  paidAmount: string;
  remainingAmount: string;
  status: string;
  paidAt?: string;
}

export interface LoanRequest {
  creditLineId: string;
  amount: number;
  installments: number;
  amortizationType: "french" | "german";
}

export interface SimulateRequest {
  amount: number;
  installments: number;
  interestRate: number;
  amortizationType: "french" | "german";
}

export interface SimulationResult {
  principal: string;
  interestRate: string;
  totalInterest: string;
  totalPayment: string;
  installments: Installment[];
}

// ---- Payments ----
export interface Payment {
  id: string;
  loanId: string;
  installmentId?: string;
  amount: string;
  method: string;
  reference?: string;
  isAdjustment: boolean;
  adjustmentNote?: string;
  createdAt: string;
}

export interface RecordPaymentRequest {
  amount: number;
  method: string;
}

export interface AdjustPaymentRequest {
  amount: number;
  reason: string;
}

// ---- Admin: Clients ----
export interface Client {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  dni: string;
  cuit: string;
  dateOfBirth: string;
  phone: string;
  address: string;
  city: string;
  province: string;
  isPEP: boolean;
  isBlocked: boolean;
  createdAt: string;
}

// ---- Admin: Dashboard ----
export interface PortfolioSummary {
  totalClients: number;
  activeLoans: number;
  totalDisbursed: number;
  totalCollected: number;
  pendingApprovals: number;
}

export interface DelinquencySummary {
  delinquencyRate: number;
  overdueCount: number;
  totalOverdue: number;
}

export interface KPIs {
  totalClients: number;
  activeLoans: number;
  totalDisbursed: number;
  delinquencyRate: number;
  collectionRate: number;
  averageLoanAmount: number;
}

export interface TrendData {
  month: string;
  amount: number;
  count: number;
}

// ---- Admin: Audit ----
export interface AuditEntry {
  id: string;
  userId?: string;
  action: string;
  entityType: string;
  entityId: string;
  description: string;
  ip: string;
  userAgent: string;
  createdAt: string;
}

// ---- Vendor ----
export interface Vendor {
  id: string;
  businessName: string;
  cuit: string;
  phone: string;
  address: string;
  city: string;
  province: string;
  isActive: boolean;
  email: string;
  createdAt: string;
}

export interface VendorAccount {
  id: string;
  vendorId: string;
  balance: string;
}

export interface VendorMovement {
  id: string;
  type: string;
  amount: string;
  balanceAfter: string;
  description: string;
  reference?: string;
  createdAt: string;
}

export interface Purchase {
  id: string;
  vendorId: string;
  vendorName?: string;
  clientId: string;
  clientName?: string;
  creditLineId: string;
  amount: string;
  description: string;
  createdAt: string;
}

export interface VendorPayment {
  id: string;
  vendorId: string;
  amount: string;
  method: string;
  reference?: string;
  paidBy: string;
  createdAt: string;
}

export interface RegisterVendorRequest {
  email: string;
  password: string;
  businessName: string;
  cuit: string;
  phone: string;
  address: string;
  city: string;
  province: string;
}

export interface UpdateVendorProfileRequest {
  phone: string;
  address: string;
  city: string;
  province: string;
}

export interface RecordPurchaseRequest {
  clientId: string;
  creditLineId: string;
  amount: string;
  description: string;
}

export interface RecordVendorPaymentRequest {
  vendorId: string;
  amount: string;
  method: string;
  reference?: string;
}

export interface RegisterClientByVendorRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  dni: string;
  cuit: string;
  dateOfBirth: string;
  phone: string;
  address: string;
  city: string;
  province: string;
  isPEP: boolean;
}

export interface RequestCreditLineByVendorRequest {
  clientId: string;
  maxAmount: string;
  interestRate: string;
  maxInstallments: number;
}

// ---- Pagination ----
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  offset: number;
  limit: number;
}
