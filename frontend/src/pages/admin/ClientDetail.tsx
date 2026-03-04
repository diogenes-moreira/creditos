import React, { useState, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Button,
  Divider,
  CircularProgress,
  Alert,
  Tabs,
  Tab,
  Snackbar,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Chip,
  IconButton,
} from "@mui/material";
import {
  ArrowBack as BackIcon,
  AccountBalance as BalanceIcon,
  CreditCard as LoanIcon,
  Payment as PaymentIcon,
  ShoppingCart as PurchaseIcon,
  Edit as EditIcon,
} from "@mui/icons-material";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { format } from "date-fns";
import {
  adminGetClient,
  adminGetClientLoans,
  adminGetClientCreditLines,
  adminGetClientPayments,
  adminGetClientPurchases,
  adminGetClientAccount,
  adminGetClientMovements,
  adminUpdateCreditLine,
} from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import StatusBadge from "../../components/StatusBadge";
import KPICard from "../../components/KPICard";
import type { Loan, Payment, Purchase, Movement, CreditLine } from "../../api/types";

const ClientDetail: React.FC = () => {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState(0);
  const [loansPage, setLoansPage] = useState(1);
  const [paymentsPage, setPaymentsPage] = useState(1);
  const [purchasesPage, setPurchasesPage] = useState(1);
  const [movementsPage, setMovementsPage] = useState(1);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editingCreditLine, setEditingCreditLine] = useState<CreditLine | null>(null);
  const [newMaxAmount, setNewMaxAmount] = useState("");
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" as "success" | "error" });

  const { data: client, isLoading } = useQuery({
    queryKey: ["admin-client", id],
    queryFn: () => adminGetClient(id!),
    enabled: !!id,
  });

  const { data: account } = useQuery({
    queryKey: ["admin-client-account", id],
    queryFn: () => adminGetClientAccount(id!),
    enabled: !!id,
  });

  const { data: creditLines } = useQuery({
    queryKey: ["admin-client-credit-lines", id],
    queryFn: () => adminGetClientCreditLines(id!),
    enabled: !!id,
  });

  const { data: loansData, isLoading: loansLoading } = useQuery({
    queryKey: ["admin-client-loans", id, loansPage],
    queryFn: () => adminGetClientLoans(id!, loansPage),
    enabled: !!id && activeTab === 0,
  });

  const { data: paymentsData, isLoading: paymentsLoading } = useQuery({
    queryKey: ["admin-client-payments", id, paymentsPage],
    queryFn: () => adminGetClientPayments(id!, paymentsPage),
    enabled: !!id && activeTab === 1,
  });

  const { data: purchasesData, isLoading: purchasesLoading } = useQuery({
    queryKey: ["admin-client-purchases", id, purchasesPage],
    queryFn: () => adminGetClientPurchases(id!, purchasesPage),
    enabled: !!id && activeTab === 2,
  });

  const { data: movementsData, isLoading: movementsLoading } = useQuery({
    queryKey: ["admin-client-movements", id, movementsPage],
    queryFn: () => adminGetClientMovements(id!, movementsPage),
    enabled: !!id && activeTab === 3,
  });

  const updateCreditLineMutation = useMutation({
    mutationFn: ({ creditLineId, maxAmount }: { creditLineId: string; maxAmount: string }) =>
      adminUpdateCreditLine(creditLineId, { maxAmount }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-client-credit-lines", id] });
      setEditDialogOpen(false);
      setEditingCreditLine(null);
      setNewMaxAmount("");
      setSnackbar({ open: true, message: t("admin.creditLineUpdated"), severity: "success" });
    },
    onError: () => {
      setSnackbar({ open: true, message: t("admin.creditLineUpdateError"), severity: "error" });
    },
  });

  const formatMoney = (amount: number) =>
    new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS", maximumFractionDigits: 0 }).format(amount);

  // Chart data: aggregate payments by month
  const chartData = useMemo(() => {
    if (!paymentsData?.data) return [];
    const byMonth: Record<string, number> = {};
    paymentsData.data.forEach((p) => {
      const month = new Date(p.createdAt).toLocaleDateString("es-AR", { month: "short", year: "2-digit" });
      byMonth[month] = (byMonth[month] || 0) + parseFloat(p.amount);
    });
    return Object.entries(byMonth).map(([month, amount]) => ({ month, pagos: amount }));
  }, [paymentsData]);

  // KPI computations
  const activeLoansCount = loansData?.data?.filter((l) => l.status === "active" || l.status === "disbursed").length || 0;
  const totalPaid = paymentsData?.data?.reduce((sum, p) => sum + parseFloat(p.amount), 0) || 0;
  const totalPurchasesCount = purchasesData?.total || 0;

  const loanColumns: Column<Loan>[] = [
    { id: "id", label: "ID", minWidth: 90, render: (row) => `#${row.id.slice(0, 8)}` },
    { id: "principal", label: t("loans.principal"), align: "right", render: (row) => <MoneyDisplay amount={row.principal} fontWeight={500} /> },
    { id: "interestRate", label: t("admin.rate"), align: "center", render: (row) => `${row.interestRate}%` },
    { id: "numInstallments", label: t("loans.installments"), align: "center" },
    { id: "amortizationType", label: t("loans.system"), render: (row) => <Chip label={row.amortizationType === "french" ? t("loans.french") : t("loans.german")} size="small" variant="outlined" /> },
    { id: "status", label: t("common.status"), render: (row) => <StatusBadge status={row.status} /> },
    { id: "createdAt", label: t("common.date"), render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy") },
  ];

  const paymentColumns: Column<Payment>[] = [
    { id: "createdAt", label: t("common.date"), minWidth: 120, render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy HH:mm") },
    { id: "amount", label: t("common.amount"), align: "right", render: (row) => <MoneyDisplay amount={row.amount} fontWeight={500} /> },
    { id: "method", label: t("payments.method"), render: (row) => <Chip label={row.method} size="small" variant="outlined" /> },
    { id: "loanId", label: t("payments.loan"), render: (row) => `#${row.loanId.slice(0, 8)}` },
    { id: "isAdjustment", label: t("common.status"), render: (row) => row.isAdjustment ? <Chip label={t("payments.adjustment")} size="small" color="warning" /> : <Chip label="OK" size="small" color="success" /> },
  ];

  const purchaseColumns: Column<Purchase>[] = [
    { id: "createdAt", label: t("common.date"), minWidth: 120, render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy") },
    { id: "vendorName", label: t("nav.vendors"), render: (row) => row.vendorName || row.vendorId.slice(0, 8) },
    { id: "amount", label: t("common.amount"), align: "right", render: (row) => <MoneyDisplay amount={row.amount} fontWeight={500} /> },
    { id: "description", label: t("common.description") },
  ];

  const movementColumns: Column<Movement>[] = [
    { id: "createdAt", label: t("common.date"), minWidth: 120, render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy HH:mm") },
    { id: "type", label: t("common.type"), render: (row) => <Chip label={row.type === "credit" ? t("account.credit") : t("account.debit")} size="small" color={row.type === "credit" ? "success" : "error"} /> },
    { id: "amount", label: t("common.amount"), align: "right", render: (row) => <MoneyDisplay amount={row.amount} fontWeight={500} color={row.type === "credit" ? "success.main" : "error.main"} /> },
    { id: "balanceAfter", label: t("account.balance"), align: "right", render: (row) => <MoneyDisplay amount={row.balanceAfter} /> },
    { id: "description", label: t("common.description") },
  ];

  const handleEditCreditLine = (cl: CreditLine) => {
    setEditingCreditLine(cl);
    setNewMaxAmount(cl.maxAmount);
    setEditDialogOpen(true);
  };

  const handleSaveCreditLine = () => {
    if (!editingCreditLine) return;
    updateCreditLineMutation.mutate({ creditLineId: editingCreditLine.id, maxAmount: newMaxAmount });
  };

  if (isLoading) {
    return <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>;
  }

  if (!client) {
    return <Alert severity="error">{t("admin.clientNotFound")}</Alert>;
  }

  const activeCreditLine = creditLines?.find((cl) => cl.status === "approved");

  return (
    <Box>
      <Button startIcon={<BackIcon />} onClick={() => navigate("/admin/clients")} sx={{ mb: 2 }}>
        {t("admin.backToClients")}
      </Button>

      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">
          {client.firstName} {client.lastName}
        </Typography>
        <StatusBadge status={client.isBlocked ? "blocked" : "active"} size="medium" />
      </Box>

      {/* KPI Cards */}
      <Grid container spacing={3} mb={3}>
        <Grid item xs={12} sm={6} md={3}>
          <KPICard
            icon={<BalanceIcon sx={{ fontSize: 28 }} />}
            label={t("admin.clientDetailBalance")}
            value={account ? formatMoney(parseFloat(account.balance)) : "$0"}
            color="primary.main"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <KPICard
            icon={<LoanIcon sx={{ fontSize: 28 }} />}
            label={t("admin.clientDetailActiveLoans")}
            value={activeLoansCount}
            color="info.main"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <KPICard
            icon={<PaymentIcon sx={{ fontSize: 28 }} />}
            label={t("admin.clientDetailTotalPaid")}
            value={formatMoney(totalPaid)}
            color="success.main"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <KPICard
            icon={<PurchaseIcon sx={{ fontSize: 28 }} />}
            label={t("admin.clientDetailTotalPurchases")}
            value={totalPurchasesCount}
            color="warning.main"
          />
        </Grid>
      </Grid>

      {/* Credit Line Card */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
            <Typography variant="h6">{t("admin.clientDetailCreditLine")}</Typography>
          </Box>
          <Divider sx={{ mb: 2 }} />
          {creditLines && creditLines.length > 0 ? (
            creditLines.map((cl) => (
              <Box key={cl.id} mb={2}>
                <Grid container spacing={2} alignItems="center">
                  <Grid item xs={12} sm={6} md={2}>
                    <Typography variant="body2" color="text.secondary">{t("creditLines.maxAmount")}</Typography>
                    <Box display="flex" alignItems="center" gap={1}>
                      <MoneyDisplay amount={cl.maxAmount} fontWeight={600} />
                      <IconButton size="small" onClick={() => handleEditCreditLine(cl)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Box>
                  </Grid>
                  <Grid item xs={12} sm={6} md={2}>
                    <Typography variant="body2" color="text.secondary">{t("creditLines.usedAmount")}</Typography>
                    <MoneyDisplay amount={cl.usedAmount} color="error.main" />
                  </Grid>
                  <Grid item xs={12} sm={6} md={2}>
                    <Typography variant="body2" color="text.secondary">{t("creditLines.availableAmount")}</Typography>
                    <MoneyDisplay amount={cl.availableAmount} color="success.main" />
                  </Grid>
                  <Grid item xs={12} sm={6} md={2}>
                    <Typography variant="body2" color="text.secondary">{t("loans.interestRate")}</Typography>
                    <Typography fontWeight={500}>{cl.interestRate}%</Typography>
                  </Grid>
                  <Grid item xs={12} sm={6} md={2}>
                    <Typography variant="body2" color="text.secondary">{t("creditLines.maxInstallments")}</Typography>
                    <Typography fontWeight={500}>{cl.maxInstallments}</Typography>
                  </Grid>
                  <Grid item xs={12} sm={6} md={2}>
                    <StatusBadge status={cl.status} />
                  </Grid>
                </Grid>
                {creditLines.length > 1 && <Divider sx={{ mt: 2 }} />}
              </Box>
            ))
          ) : (
            <Typography color="text.secondary">{t("admin.noCreditLines")}</Typography>
          )}
        </CardContent>
      </Card>

      {/* Payment Evolution Chart */}
      {chartData.length > 0 && (
        <Card sx={{ mb: 3 }}>
          <CardContent>
            <Typography variant="h6" gutterBottom>{t("admin.clientDetailPaymentEvolution")}</Typography>
            <Box sx={{ width: "100%", height: 300 }}>
              <ResponsiveContainer>
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="month" />
                  <YAxis tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`} />
                  <Tooltip formatter={(value: number) => formatMoney(value)} />
                  <Bar dataKey="pagos" name={t("admin.clientDetailPaymentsTab")} fill="#2E7D32" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </Box>
          </CardContent>
        </Card>
      )}

      {/* Personal Info Card */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" mb={2}>{t("admin.personalInfo")}</Typography>
          <Divider sx={{ mb: 2 }} />
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("auth.email")}</Typography>
              <Typography>{client.email}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">DNI</Typography>
              <Typography>{client.dni}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("registration.phone")}</Typography>
              <Typography>{client.phone}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("registration.address")}</Typography>
              <Typography>{client.address}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("registration.city")}</Typography>
              <Typography>{client.city}, {client.province}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("admin.registrationDate")}</Typography>
              <Typography>{format(new Date(client.createdAt), "dd/MM/yyyy")}</Typography>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Tabbed Tables */}
      <Card>
        <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
          <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)}>
            <Tab label={t("admin.clientDetailLoansTab")} />
            <Tab label={t("admin.clientDetailPaymentsTab")} />
            <Tab label={t("admin.clientDetailPurchasesTab")} />
            <Tab label={t("admin.clientDetailMovementsTab")} />
          </Tabs>
        </Box>
        <CardContent>
          {activeTab === 0 && (
            <DataTable
              columns={loanColumns}
              rows={loansData?.data || []}
              total={loansData?.total || 0}
              page={loansPage - 1}
              pageSize={20}
              onPageChange={(p) => setLoansPage(p + 1)}
              loading={loansLoading}
              keyExtractor={(row) => row.id}
              emptyMessage={t("loans.noLoans")}
            />
          )}
          {activeTab === 1 && (
            <DataTable
              columns={paymentColumns}
              rows={paymentsData?.data || []}
              total={paymentsData?.total || 0}
              page={paymentsPage - 1}
              pageSize={20}
              onPageChange={(p) => setPaymentsPage(p + 1)}
              loading={paymentsLoading}
              keyExtractor={(row) => row.id}
              emptyMessage={t("payments.noPayments")}
            />
          )}
          {activeTab === 2 && (
            <DataTable
              columns={purchaseColumns}
              rows={purchasesData?.data || []}
              total={purchasesData?.total || 0}
              page={purchasesPage - 1}
              pageSize={20}
              onPageChange={(p) => setPurchasesPage(p + 1)}
              loading={purchasesLoading}
              keyExtractor={(row) => row.id}
              emptyMessage={t("vendor.noPurchases")}
            />
          )}
          {activeTab === 3 && (
            <DataTable
              columns={movementColumns}
              rows={movementsData?.data || []}
              total={movementsData?.total || 0}
              page={movementsPage - 1}
              pageSize={20}
              onPageChange={(p) => setMovementsPage(p + 1)}
              loading={movementsLoading}
              keyExtractor={(row) => row.id}
              emptyMessage={t("account.noMovements")}
            />
          )}
        </CardContent>
      </Card>

      {/* Edit Credit Line Dialog */}
      <Dialog open={editDialogOpen} onClose={() => setEditDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("admin.clientDetailEditLimit")}</DialogTitle>
        <DialogContent>
          {editingCreditLine && (
            <Box mt={1}>
              <Typography variant="body2" color="text.secondary" mb={2}>
                {t("creditLines.usedAmount")}: <MoneyDisplay amount={editingCreditLine.usedAmount} />
              </Typography>
              <TextField
                fullWidth
                label={t("admin.clientDetailNewLimit")}
                type="number"
                value={newMaxAmount}
                onChange={(e) => setNewMaxAmount(e.target.value)}
                inputProps={{ min: parseFloat(editingCreditLine.usedAmount) }}
              />
            </Box>
          )}
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setEditDialogOpen(false)} disabled={updateCreditLineMutation.isPending}>
            {t("common.cancel")}
          </Button>
          <Button
            onClick={handleSaveCreditLine}
            variant="contained"
            disabled={updateCreditLineMutation.isPending || !newMaxAmount}
          >
            {updateCreditLineMutation.isPending ? t("common.processing") : t("common.save")}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert severity={snackbar.severity} onClose={() => setSnackbar({ ...snackbar, open: false })}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default ClientDetail;
