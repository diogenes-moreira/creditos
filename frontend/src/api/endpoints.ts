import apiClient from "./client";
import type {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  Profile,
  UpdateProfileRequest,
  UpdateMercadoPagoRequest,
  Account,
  Movement,
  Loan,
  LoanDetail,
  LoanRequest,
  SimulateRequest,
  SimulationResult,
  Payment,
  RecordPaymentRequest,
  AdjustPaymentRequest,
  Client,
  CreditLine,
  CreateCreditLineRequest,
  PortfolioSummary,
  DelinquencySummary,
  KPIs,
  TrendData,
  AuditEntry,
  PaginatedResponse,
  Vendor,
  VendorAccount,
  VendorMovement,
  Purchase,
  VendorPayment,
  RegisterVendorRequest,
  UpdateVendorProfileRequest,
  RecordPurchaseRequest,
  RecordVendorPaymentRequest,
  RegisterClientByVendorRequest,
  RequestCreditLineByVendorRequest,
} from "./types";

// ==================== Auth ====================

export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  const res = await apiClient.post("/auth/login", data);
  return res.data;
};

export const register = async (data: RegisterRequest): Promise<AuthResponse> => {
  const res = await apiClient.post("/auth/register", data);
  return res.data;
};

// ==================== Profile ====================

export const getProfile = async (): Promise<Profile> => {
  const res = await apiClient.get("/me/profile");
  return res.data;
};

export const updateProfile = async (data: UpdateProfileRequest): Promise<Profile> => {
  const res = await apiClient.put("/me/profile", data);
  return res.data;
};

export const updateMercadoPago = async (data: UpdateMercadoPagoRequest): Promise<Profile> => {
  const res = await apiClient.put("/me/mercadopago", data);
  return res.data;
};

// ==================== Account ====================

export const getAccount = async (): Promise<Account> => {
  const res = await apiClient.get("/me/account");
  return res.data;
};

export const getMovements = async (page = 1, pageSize = 20): Promise<PaginatedResponse<Movement>> => {
  const res = await apiClient.get("/me/account/movements", { params: { page, pageSize } });
  return res.data;
};

// ==================== Loans ====================

export const getLoans = async (): Promise<Loan[]> => {
  const res = await apiClient.get("/me/loans");
  return res.data;
};

export const getLoan = async (id: string): Promise<LoanDetail> => {
  const res = await apiClient.get(`/me/loans/${id}`);
  return res.data;
};

export const requestLoan = async (data: LoanRequest): Promise<Loan> => {
  const res = await apiClient.post("/me/loans", data);
  return res.data;
};

export const simulateLoan = async (data: SimulateRequest): Promise<SimulationResult> => {
  const res = await apiClient.post("/loans/simulate", data);
  return res.data;
};

// ==================== Payments ====================

export const recordPayment = async (loanId: string, data: RecordPaymentRequest): Promise<Payment> => {
  const res = await apiClient.post(`/me/loans/${loanId}/payments`, data);
  return res.data;
};

export const getPayments = async (): Promise<Payment[]> => {
  const res = await apiClient.get("/me/payments");
  return res.data;
};

// ==================== Admin: Clients ====================

export const adminGetClients = async (page = 1, pageSize = 20): Promise<PaginatedResponse<Client>> => {
  const res = await apiClient.get("/admin/clients", { params: { page, pageSize } });
  return res.data;
};

export const adminGetClient = async (id: string): Promise<Client> => {
  const res = await apiClient.get(`/admin/clients/${id}`);
  return res.data;
};

export const adminSearchClients = async (query: string): Promise<Client[]> => {
  const res = await apiClient.get("/admin/clients/search", { params: { q: query } });
  return res.data;
};

// ==================== Admin: Credit Lines ====================

export const adminCreateCreditLine = async (data: CreateCreditLineRequest): Promise<CreditLine> => {
  const res = await apiClient.post("/admin/credit-lines", data);
  return res.data;
};

export const adminApproveCreditLine = async (id: string): Promise<CreditLine> => {
  const res = await apiClient.post(`/admin/credit-lines/${id}/approve`);
  return res.data;
};

export const adminRejectCreditLine = async (id: string): Promise<CreditLine> => {
  const res = await apiClient.post(`/admin/credit-lines/${id}/reject`);
  return res.data;
};

// ==================== Admin: Loans ====================

export const adminGetPendingLoans = async (): Promise<Loan[]> => {
  const res = await apiClient.get("/admin/loans/pending");
  return res.data;
};

export const adminApproveLoan = async (id: string): Promise<Loan> => {
  const res = await apiClient.post(`/admin/loans/${id}/approve`);
  return res.data;
};

export const adminDisburseLoan = async (id: string): Promise<Loan> => {
  const res = await apiClient.post(`/admin/loans/${id}/disburse`);
  return res.data;
};

export const adminCancelLoan = async (id: string): Promise<Loan> => {
  const res = await apiClient.post(`/admin/loans/${id}/cancel`);
  return res.data;
};

export const adminPrepayLoan = async (id: string): Promise<Loan> => {
  const res = await apiClient.post(`/admin/loans/${id}/prepay`);
  return res.data;
};

// ==================== Admin: Payments ====================

export const adminAdjustPayment = async (id: string, data: AdjustPaymentRequest): Promise<Payment> => {
  const res = await apiClient.put(`/admin/payments/${id}/adjust`, data);
  return res.data;
};

// ==================== Admin: Dashboard ====================

export const getDashboard = async (): Promise<PortfolioSummary> => {
  const res = await apiClient.get("/admin/dashboard/portfolio");
  return res.data;
};

export const getDelinquency = async (): Promise<DelinquencySummary> => {
  const res = await apiClient.get("/admin/dashboard/delinquency");
  return res.data;
};

export const getKPIs = async (): Promise<KPIs> => {
  const res = await apiClient.get("/admin/dashboard/kpis");
  return res.data;
};

export const getDisbursementTrends = async (): Promise<TrendData[]> => {
  const res = await apiClient.get("/admin/dashboard/trends/disbursements");
  return res.data;
};

export const getCollectionTrends = async (): Promise<TrendData[]> => {
  const res = await apiClient.get("/admin/dashboard/trends/collections");
  return res.data;
};

// ==================== Vendor: Self-Service ====================

export const getVendorProfile = async (): Promise<Vendor> => {
  const res = await apiClient.get("/me/vendor/profile");
  return res.data;
};

export const updateVendorProfile = async (data: UpdateVendorProfileRequest): Promise<Vendor> => {
  const res = await apiClient.put("/me/vendor/profile", data);
  return res.data;
};

export const getVendorAccount = async (): Promise<VendorAccount> => {
  const res = await apiClient.get("/me/vendor/account");
  return res.data;
};

export const getVendorMovements = async (page = 1, pageSize = 20): Promise<PaginatedResponse<VendorMovement>> => {
  const res = await apiClient.get("/me/vendor/account/movements", { params: { offset: (page - 1) * pageSize, limit: pageSize } });
  return res.data;
};

export const getVendorPurchases = async (page = 1, pageSize = 20): Promise<PaginatedResponse<Purchase>> => {
  const res = await apiClient.get("/me/vendor/purchases", { params: { offset: (page - 1) * pageSize, limit: pageSize } });
  return res.data;
};

export const vendorSearchClients = async (query: string, page = 1, pageSize = 20): Promise<PaginatedResponse<Client>> => {
  const res = await apiClient.get("/me/vendor/clients/search", { params: { q: query, offset: (page - 1) * pageSize, limit: pageSize } });
  return res.data;
};

export const vendorGetClientCreditLines = async (clientId: string): Promise<CreditLine[]> => {
  const res = await apiClient.get(`/me/vendor/clients/${clientId}/credit-lines`);
  return res.data;
};

export const vendorRegisterClient = async (data: RegisterClientByVendorRequest): Promise<Client> => {
  const res = await apiClient.post("/me/vendor/clients/register", data);
  return res.data;
};

export const vendorRequestCreditLine = async (clientId: string, data: RequestCreditLineByVendorRequest): Promise<CreditLine> => {
  const res = await apiClient.post(`/me/vendor/clients/${clientId}/credit-lines`, data);
  return res.data;
};

export const vendorRecordPurchase = async (data: RecordPurchaseRequest): Promise<Purchase> => {
  const res = await apiClient.post("/me/vendor/purchases", data);
  return res.data;
};

// ==================== Admin: Vendors ====================

export const adminGetVendors = async (page = 1, pageSize = 20, query = ""): Promise<PaginatedResponse<Vendor>> => {
  const res = await apiClient.get("/admin/vendors", { params: { q: query, offset: (page - 1) * pageSize, limit: pageSize } });
  return res.data;
};

export const adminGetVendor = async (id: string): Promise<Vendor> => {
  const res = await apiClient.get(`/admin/vendors/${id}`);
  return res.data;
};

export const adminRegisterVendor = async (data: RegisterVendorRequest): Promise<Vendor> => {
  const res = await apiClient.post("/admin/vendors", data);
  return res.data;
};

export const adminActivateVendor = async (id: string): Promise<void> => {
  await apiClient.post(`/admin/vendors/${id}/activate`);
};

export const adminDeactivateVendor = async (id: string): Promise<void> => {
  await apiClient.post(`/admin/vendors/${id}/deactivate`);
};

export const adminGetVendorPurchases = async (id: string, page = 1, pageSize = 20): Promise<PaginatedResponse<Purchase>> => {
  const res = await apiClient.get(`/admin/vendors/${id}/purchases`, { params: { offset: (page - 1) * pageSize, limit: pageSize } });
  return res.data;
};

export const adminGetVendorPayments = async (id: string, page = 1, pageSize = 20): Promise<PaginatedResponse<VendorPayment>> => {
  const res = await apiClient.get(`/admin/vendors/${id}/payments`, { params: { offset: (page - 1) * pageSize, limit: pageSize } });
  return res.data;
};

export const adminRecordVendorPayment = async (id: string, data: RecordVendorPaymentRequest): Promise<VendorPayment> => {
  const res = await apiClient.post(`/admin/vendors/${id}/payments`, data);
  return res.data;
};

// ==================== Admin: Audit ====================

export const getAuditLogs = async (page = 1, pageSize = 20): Promise<PaginatedResponse<AuditEntry>> => {
  const res = await apiClient.get("/admin/audit", { params: { page, pageSize } });
  return res.data;
};
