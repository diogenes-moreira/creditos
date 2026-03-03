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
  phone: string;
  address: string;
  city: string;
  province: string;
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
  phone: string;
  address: string;
  city: string;
  province: string;
  mercadoPagoAlias?: string;
  mercadoPagoCvu?: string;
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
  balance: number;
  currency: string;
  status: string;
  createdAt: string;
}

export interface Movement {
  id: string;
  accountId: string;
  type: string;
  amount: number;
  balance: number;
  description: string;
  createdAt: string;
}

// ---- Credit Lines ----
export interface CreditLine {
  id: string;
  clientId: string;
  clientName?: string;
  maxAmount: number;
  currentAmount: number;
  interestRate: number;
  status: string;
  approvedAt?: string;
  createdAt: string;
}

export interface CreateCreditLineRequest {
  clientId: string;
  maxAmount: number;
  interestRate: number;
}

// ---- Loans ----
export interface Loan {
  id: string;
  clientId: string;
  clientName?: string;
  creditLineId: string;
  amount: number;
  totalAmount: number;
  interestRate: number;
  installments: number;
  amortizationType: "french" | "german";
  status: string;
  disbursedAt?: string;
  createdAt: string;
}

export interface LoanDetail extends Loan {
  schedule: Installment[];
}

export interface Installment {
  number: number;
  dueDate: string;
  principal: number;
  interest: number;
  total: number;
  balance: number;
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
  schedule: Installment[];
  totalAmount: number;
  totalInterest: number;
  monthlyPayment?: number;
}

// ---- Payments ----
export interface Payment {
  id: string;
  loanId: string;
  installmentNumber: number;
  amount: number;
  method: string;
  status: string;
  adjustedAmount?: number;
  adjustmentReason?: string;
  paidAt: string;
  createdAt: string;
}

export interface RecordPaymentRequest {
  installmentNumber: number;
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
  phone: string;
  address: string;
  city: string;
  province: string;
  status: string;
  createdAt: string;
}

// ---- Admin: Dashboard ----
export interface PortfolioSummary {
  totalClients: number;
  activeLoans: number;
  totalDisbursed: number;
  totalCollected: number;
  pendingAmount: number;
}

export interface DelinquencySummary {
  delinquencyRate: number;
  overdueLoans: number;
  overdueAmount: number;
  averageDaysOverdue: number;
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
  userId: string;
  userEmail: string;
  action: string;
  description: string;
  ipAddress: string;
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
  page: number;
  pageSize: number;
  totalPages: number;
}
