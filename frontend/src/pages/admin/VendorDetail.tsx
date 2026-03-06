import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  Box,
  Typography,
  Paper,
  Grid,
  Chip,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
  CircularProgress,
  IconButton,
  Tooltip,
  InputAdornment,
  List,
  ListItemButton,
  ListItemText,
  Alert,
} from "@mui/material";
import {
  ArrowBack as BackIcon,
  Download as DownloadIcon,
  Search as SearchIcon,
  ShoppingCart as SaleIcon,
} from "@mui/icons-material";
import {
  adminGetVendor,
  adminActivateVendor,
  adminDeactivateVendor,
  adminGetVendorPurchases,
  adminGetVendorPayments,
  adminRecordVendorPayment,
  adminGetVendorWithdrawals,
  adminApproveWithdrawal,
  adminRejectWithdrawal,
  downloadAdminVendorPaymentReceipt,
  adminRecordVendorPurchase,
  adminSearchClients,
  adminGetClientCreditLines,
} from "../../api/endpoints";
import type { Vendor, Purchase, VendorPayment, WithdrawalRequest, Client, CreditLine } from "../../api/types";
import { useNotification } from "../../contexts/NotificationContext";
import { getErrorMessage } from "../../api/errorUtils";

const statusColors: Record<string, "warning" | "success" | "error" | "info" | "default"> = {
  pending: "warning",
  approved: "info",
  paid: "success",
  rejected: "error",
};

const VendorDetail: React.FC = () => {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [vendor, setVendor] = useState<Vendor | null>(null);
  const [loading, setLoading] = useState(true);
  const [purchases, setPurchases] = useState<Purchase[]>([]);
  const [purchaseTotal, setPurchaseTotal] = useState(0);
  const [purchasePage, setPurchasePage] = useState(0);
  const [payments, setPayments] = useState<VendorPayment[]>([]);
  const [paymentTotal, setPaymentTotal] = useState(0);
  const [paymentPage, setPaymentPage] = useState(0);
  const [withdrawals, setWithdrawals] = useState<WithdrawalRequest[]>([]);
  const [withdrawalTotal, setWithdrawalTotal] = useState(0);
  const [withdrawalPage, setWithdrawalPage] = useState(0);
  const [openPayment, setOpenPayment] = useState(false);
  const [payAmount, setPayAmount] = useState("");
  const [payMethod, setPayMethod] = useState("transfer");
  const [payRef, setPayRef] = useState("");
  const [openApprove, setOpenApprove] = useState<string | null>(null);
  const [approveRef, setApproveRef] = useState("");
  const [openReject, setOpenReject] = useState<string | null>(null);
  const [rejectReason, setRejectReason] = useState("");
  // Register sale dialog state
  const [openPurchase, setOpenPurchase] = useState(false);
  const [purchaseSearchQuery, setPurchaseSearchQuery] = useState("");
  const [purchaseSearchResults, setPurchaseSearchResults] = useState<Client[]>([]);
  const [purchaseSelectedClient, setPurchaseSelectedClient] = useState<Client | null>(null);
  const [purchaseClientCreditLines, setPurchaseClientCreditLines] = useState<CreditLine[]>([]);
  const [purchaseSelectedCreditLine, setPurchaseSelectedCreditLine] = useState<CreditLine | null>(null);
  const [purchaseAmount, setPurchaseAmount] = useState("");
  const [purchaseDescription, setPurchaseDescription] = useState("");
  const [purchaseNumInstallments, setPurchaseNumInstallments] = useState(1);
  const [purchaseAmortType, setPurchaseAmortType] = useState<"french" | "german">("french");
  const [purchaseSubmitting, setPurchaseSubmitting] = useState(false);
  const { showSuccess, showError } = useNotification();

  const fetchVendor = async () => {
    try {
      const v = await adminGetVendor(id!);
      setVendor(v);
    } catch {
      setVendor(null);
    } finally {
      setLoading(false);
    }
  };

  const fetchPurchases = async () => {
    try {
      const res = await adminGetVendorPurchases(id!, purchasePage + 1);
      setPurchases(res.data || []);
      setPurchaseTotal(res.total || 0);
    } catch { /* ignore */ }
  };

  const fetchPayments = async () => {
    try {
      const res = await adminGetVendorPayments(id!, paymentPage + 1);
      setPayments(res.data || []);
      setPaymentTotal(res.total || 0);
    } catch { /* ignore */ }
  };

  const fetchWithdrawals = async () => {
    try {
      const res = await adminGetVendorWithdrawals(id!, withdrawalPage + 1);
      setWithdrawals(res.data || []);
      setWithdrawalTotal(res.total || 0);
    } catch { /* ignore */ }
  };

  useEffect(() => { fetchVendor(); }, [id]);
  useEffect(() => { if (id) fetchPurchases(); }, [id, purchasePage]);
  useEffect(() => { if (id) fetchPayments(); }, [id, paymentPage]);
  useEffect(() => { if (id) fetchWithdrawals(); }, [id, withdrawalPage]);

  const handleToggleActive = async () => {
    try {
      if (vendor?.isActive) {
        await adminDeactivateVendor(id!);
        showSuccess(t("adminVendor.deactivated"));
      } else {
        await adminActivateVendor(id!);
        showSuccess(t("adminVendor.activated"));
      }
      fetchVendor();
    } catch (err) { showError(getErrorMessage(err, t("admin.processError"))); }
  };

  const handleRecordPayment = async () => {
    try {
      await adminRecordVendorPayment(id!, { vendorId: id!, amount: payAmount, method: payMethod, reference: payRef });
      setOpenPayment(false);
      setPayAmount(""); setPayRef("");
      showSuccess(t("adminVendor.paymentSuccess"));
      fetchPayments();
      fetchVendor();
    } catch (err) { showError(getErrorMessage(err, t("adminVendor.paymentError"))); }
  };

  const handleApproveWithdrawal = async () => {
    if (!openApprove) return;
    try {
      await adminApproveWithdrawal(openApprove, { reference: approveRef });
      setOpenApprove(null);
      setApproveRef("");
      showSuccess(t("adminVendor.approveSuccess"));
      fetchWithdrawals();
      fetchPayments();
    } catch (err) { showError(getErrorMessage(err, t("admin.processError"))); }
  };

  const handleRejectWithdrawal = async () => {
    if (!openReject) return;
    try {
      await adminRejectWithdrawal(openReject, { reason: rejectReason });
      setOpenReject(null);
      setRejectReason("");
      showSuccess(t("adminVendor.rejectSuccess"));
      fetchWithdrawals();
    } catch (err) { showError(getErrorMessage(err, t("admin.processError"))); }
  };

  const handleClosePurchase = () => {
    setOpenPurchase(false);
    setPurchaseSearchQuery("");
    setPurchaseSearchResults([]);
    setPurchaseSelectedClient(null);
    setPurchaseClientCreditLines([]);
    setPurchaseSelectedCreditLine(null);
    setPurchaseAmount("");
    setPurchaseDescription("");
    setPurchaseNumInstallments(1);
    setPurchaseAmortType("french");
  };

  const handlePurchaseSearch = async () => {
    if (!purchaseSearchQuery.trim()) return;
    const results = await adminSearchClients(purchaseSearchQuery.trim());
    setPurchaseSearchResults(results);
  };

  const handlePurchaseSelectClient = async (client: Client) => {
    setPurchaseSelectedClient(client);
    setPurchaseSearchResults([]);
    const lines = await adminGetClientCreditLines(client.id);
    const approved = lines.filter((cl) => cl.status === "approved");
    setPurchaseClientCreditLines(approved);
    if (approved.length === 1) {
      setPurchaseSelectedCreditLine(approved[0]);
    } else {
      setPurchaseSelectedCreditLine(null);
    }
  };

  const handleRecordPurchase = async () => {
    if (!purchaseSelectedClient || !purchaseSelectedCreditLine || !purchaseAmount) return;
    setPurchaseSubmitting(true);
    try {
      await adminRecordVendorPurchase(id!, {
        clientId: purchaseSelectedClient.id,
        creditLineId: purchaseSelectedCreditLine.id,
        amount: purchaseAmount,
        description: purchaseDescription,
        numInstallments: purchaseNumInstallments,
        amortizationType: purchaseAmortType,
      });
      handleClosePurchase();
      showSuccess(t("adminVendor.saleSuccess"));
      fetchPurchases();
    } catch (err) {
      showError(getErrorMessage(err, t("adminVendor.saleError")));
    } finally {
      setPurchaseSubmitting(false);
    }
  };

  const handleDownloadReceipt = async (paymentId: string) => {
    try {
      const blob = await downloadAdminVendorPaymentReceipt(id!, paymentId);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `receipt-${paymentId.slice(0, 8)}.pdf`;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (err) { showError(getErrorMessage(err, t("common.error"))); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;
  if (!vendor) return <Typography>{t("adminVendor.vendorNotFound")}</Typography>;

  return (
    <Box>
      <Button startIcon={<BackIcon />} onClick={() => navigate("/admin/vendors")} sx={{ mb: 2 }}>
        {t("adminVendor.backToVendors")}
      </Button>

      <Typography variant="h4" fontWeight={700} gutterBottom>{t("adminVendor.vendorDetail")}</Typography>

      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>{t("admin.personalInfo")}</Typography>
            <Grid container spacing={2}>
              <Grid item xs={6}><Typography color="text.secondary">{t("adminVendor.businessName")}</Typography><Typography fontWeight={600}>{vendor.businessName}</Typography></Grid>
              <Grid item xs={6}><Typography color="text.secondary">{t("adminVendor.cuit")}</Typography><Typography fontWeight={600}>{vendor.cuit}</Typography></Grid>
              <Grid item xs={6}><Typography color="text.secondary">{t("adminVendor.email")}</Typography><Typography>{vendor.email}</Typography></Grid>
              <Grid item xs={6}><Typography color="text.secondary">{t("adminVendor.phone")}</Typography><Typography>{vendor.phone}</Typography></Grid>
              <Grid item xs={6}><Typography color="text.secondary">{t("adminVendor.address")}</Typography><Typography>{vendor.address}</Typography></Grid>
              <Grid item xs={3}><Typography color="text.secondary">{t("adminVendor.city")}</Typography><Typography>{vendor.city}</Typography></Grid>
              <Grid item xs={3}><Typography color="text.secondary">{t("adminVendor.province")}</Typography><Typography>{vendor.province}</Typography></Grid>
            </Grid>
          </Paper>
        </Grid>
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3 }}>
            <Typography color="text.secondary">{t("common.status")}</Typography>
            <Chip label={vendor.isActive ? t("status.active") : t("status.inactive")} color={vendor.isActive ? "success" : "default"} sx={{ mb: 2 }} />
            <Box display="flex" flexDirection="column" gap={1}>
              <Button variant="outlined" color={vendor.isActive ? "warning" : "success"} onClick={handleToggleActive}>
                {vendor.isActive ? t("adminVendor.deactivate") : t("adminVendor.activate")}
              </Button>
              <Button variant="contained" onClick={() => setOpenPayment(true)}>
                {t("adminVendor.recordPayment")}
              </Button>
              <Button variant="contained" color="secondary" startIcon={<SaleIcon />} onClick={() => setOpenPurchase(true)}>
                {t("adminVendor.registerSale")}
              </Button>
            </Box>
          </Paper>
        </Grid>
      </Grid>

      {/* Withdrawals Section */}
      <Typography variant="h6" gutterBottom>{t("adminVendor.withdrawals")}</Typography>
      <TableContainer component={Paper} sx={{ mb: 3 }}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t("common.date")}</TableCell>
              <TableCell>{t("common.amount")}</TableCell>
              <TableCell>{t("payments.method")}</TableCell>
              <TableCell>{t("common.status")}</TableCell>
              <TableCell>{t("adminVendor.rejectionReason")}</TableCell>
              <TableCell>{t("common.actions")}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {withdrawals.map((w) => (
              <TableRow key={w.id}>
                <TableCell>{new Date(w.requestedAt).toLocaleDateString()}</TableCell>
                <TableCell>${w.amount}</TableCell>
                <TableCell>{t(`payments.${w.method}`)}</TableCell>
                <TableCell>
                  <Chip label={t(`status.${w.status}`)} size="small" color={statusColors[w.status] || "default"} />
                </TableCell>
                <TableCell>{w.rejectionReason || "-"}</TableCell>
                <TableCell>
                  {w.status === "pending" && (
                    <Box display="flex" gap={1}>
                      <Button size="small" variant="contained" color="success" onClick={() => setOpenApprove(w.id)}>
                        {t("admin.approve")}
                      </Button>
                      <Button size="small" variant="outlined" color="error" onClick={() => setOpenReject(w.id)}>
                        {t("creditLines.reject")}
                      </Button>
                    </Box>
                  )}
                </TableCell>
              </TableRow>
            ))}
            {withdrawals.length === 0 && <TableRow><TableCell colSpan={6} align="center">{t("vendor.noWithdrawals")}</TableCell></TableRow>}
          </TableBody>
        </Table>
        <TablePagination component="div" count={withdrawalTotal} page={withdrawalPage} onPageChange={(_, p) => setWithdrawalPage(p)} rowsPerPage={20} rowsPerPageOptions={[20]} />
      </TableContainer>

      <Typography variant="h6" gutterBottom>{t("adminVendor.purchases")}</Typography>
      <TableContainer component={Paper} sx={{ mb: 3 }}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t("common.date")}</TableCell>
              <TableCell>{t("admin.client")}</TableCell>
              <TableCell>{t("common.amount")}</TableCell>
              <TableCell>{t("common.description")}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {purchases.map((p) => (
              <TableRow key={p.id}>
                <TableCell>{new Date(p.createdAt).toLocaleDateString()}</TableCell>
                <TableCell>{p.clientName || p.clientId}</TableCell>
                <TableCell>${p.amount}</TableCell>
                <TableCell>{p.description}</TableCell>
              </TableRow>
            ))}
            {purchases.length === 0 && <TableRow><TableCell colSpan={4} align="center">{t("vendor.noPurchases")}</TableCell></TableRow>}
          </TableBody>
        </Table>
        <TablePagination component="div" count={purchaseTotal} page={purchasePage} onPageChange={(_, p) => setPurchasePage(p)} rowsPerPage={20} rowsPerPageOptions={[20]} />
      </TableContainer>

      <Typography variant="h6" gutterBottom>{t("adminVendor.payments")}</Typography>
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t("common.date")}</TableCell>
              <TableCell>{t("common.amount")}</TableCell>
              <TableCell>{t("payments.method")}</TableCell>
              <TableCell>{t("payments.reference")}</TableCell>
              <TableCell></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {payments.map((p) => (
              <TableRow key={p.id}>
                <TableCell>{new Date(p.createdAt).toLocaleDateString()}</TableCell>
                <TableCell>${p.amount}</TableCell>
                <TableCell>{t(`payments.${p.method}`)}</TableCell>
                <TableCell>{p.reference}</TableCell>
                <TableCell>
                  <Tooltip title={t("adminVendor.downloadReceipt")}>
                    <IconButton size="small" onClick={() => handleDownloadReceipt(p.id)}>
                      <DownloadIcon />
                    </IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
            {payments.length === 0 && <TableRow><TableCell colSpan={5} align="center">{t("payments.noPayments")}</TableCell></TableRow>}
          </TableBody>
        </Table>
        <TablePagination component="div" count={paymentTotal} page={paymentPage} onPageChange={(_, p) => setPaymentPage(p)} rowsPerPage={20} rowsPerPageOptions={[20]} />
      </TableContainer>

      {/* Record Payment Dialog */}
      <Dialog open={openPayment} onClose={() => setOpenPayment(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("adminVendor.recordPayment")}</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} mt={1}>
            <TextField label={t("adminVendor.paymentAmount")} value={payAmount} onChange={(e) => setPayAmount(e.target.value)} type="number" />
            <TextField label={t("adminVendor.paymentMethod")} select value={payMethod} onChange={(e) => setPayMethod(e.target.value)}>
              <MenuItem value="cash">{t("payments.cash")}</MenuItem>
              <MenuItem value="transfer">{t("payments.transfer")}</MenuItem>
            </TextField>
            <TextField label={t("adminVendor.paymentReference")} value={payRef} onChange={(e) => setPayRef(e.target.value)} />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenPayment(false)}>{t("common.cancel")}</Button>
          <Button variant="contained" onClick={handleRecordPayment}>{t("common.confirm")}</Button>
        </DialogActions>
      </Dialog>

      {/* Approve Withdrawal Dialog */}
      <Dialog open={!!openApprove} onClose={() => setOpenApprove(null)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("adminVendor.approveWithdrawal")}</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} mt={1}>
            <TextField label={t("adminVendor.reference")} value={approveRef} onChange={(e) => setApproveRef(e.target.value)} />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenApprove(null)}>{t("common.cancel")}</Button>
          <Button variant="contained" color="success" onClick={handleApproveWithdrawal}>{t("common.confirm")}</Button>
        </DialogActions>
      </Dialog>

      {/* Reject Withdrawal Dialog */}
      <Dialog open={!!openReject} onClose={() => setOpenReject(null)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("adminVendor.rejectWithdrawal")}</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} mt={1}>
            <TextField
              label={t("adminVendor.rejectionReason")}
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
              multiline
              rows={3}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenReject(null)}>{t("common.cancel")}</Button>
          <Button variant="contained" color="error" onClick={handleRejectWithdrawal} disabled={!rejectReason}>{t("common.confirm")}</Button>
        </DialogActions>
      </Dialog>

      {/* Register Sale Dialog */}
      <Dialog open={openPurchase} onClose={handleClosePurchase} maxWidth="sm" fullWidth>
        <DialogTitle>{t("adminVendor.registerSale")}</DialogTitle>
        <DialogContent>
          <Box mt={1} display="flex" flexDirection="column" gap={2}>
            {!purchaseSelectedClient ? (
              <>
                <TextField
                  fullWidth
                  size="small"
                  placeholder={t("adminVendor.searchClient")}
                  value={purchaseSearchQuery}
                  onChange={(e) => setPurchaseSearchQuery(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handlePurchaseSearch()}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <SearchIcon color="action" />
                      </InputAdornment>
                    ),
                  }}
                />
                <Button size="small" variant="outlined" onClick={handlePurchaseSearch} disabled={!purchaseSearchQuery.trim()}>
                  {t("common.search")}
                </Button>
                {purchaseSearchResults.length > 0 && (
                  <List dense sx={{ border: 1, borderColor: "divider", borderRadius: 1, maxHeight: 200, overflow: "auto" }}>
                    {purchaseSearchResults.map((c) => (
                      <ListItemButton key={c.id} onClick={() => handlePurchaseSelectClient(c)}>
                        <ListItemText
                          primary={`${c.firstName} ${c.lastName}`}
                          secondary={`DNI: ${c.dni} | ${c.email}`}
                        />
                      </ListItemButton>
                    ))}
                  </List>
                )}
              </>
            ) : (
              <>
                <Box display="flex" justifyContent="space-between" alignItems="center">
                  <Typography variant="subtitle2">
                    {t("admin.client")}: {purchaseSelectedClient.firstName} {purchaseSelectedClient.lastName} (DNI: {purchaseSelectedClient.dni})
                  </Typography>
                  <Button size="small" onClick={() => { setPurchaseSelectedClient(null); setPurchaseClientCreditLines([]); setPurchaseSelectedCreditLine(null); }}>
                    {t("common.edit")}
                  </Button>
                </Box>

                {purchaseClientCreditLines.length === 0 ? (
                  <Alert severity="warning">{t("admin.noCreditLineForClient")}</Alert>
                ) : purchaseClientCreditLines.length === 1 ? (
                  <Box sx={{ p: 1.5, border: 1, borderColor: "divider", borderRadius: 1 }}>
                    <Typography variant="body2">
                      {t("adminVendor.selectCreditLine")}: ${purchaseSelectedCreditLine?.availableAmount} {t("creditLines.availableAmount")}
                    </Typography>
                  </Box>
                ) : (
                  <TextField
                    select
                    fullWidth
                    label={t("adminVendor.selectCreditLine")}
                    value={purchaseSelectedCreditLine?.id || ""}
                    onChange={(e) => setPurchaseSelectedCreditLine(purchaseClientCreditLines.find((cl) => cl.id === e.target.value) || null)}
                  >
                    {purchaseClientCreditLines.map((cl) => (
                      <MenuItem key={cl.id} value={cl.id}>
                        ${cl.availableAmount} {t("creditLines.availableAmount")}
                      </MenuItem>
                    ))}
                  </TextField>
                )}

                {purchaseSelectedCreditLine && (
                  <>
                    <TextField
                      fullWidth
                      type="number"
                      label={t("adminVendor.saleAmount")}
                      value={purchaseAmount}
                      onChange={(e) => setPurchaseAmount(e.target.value)}
                    />
                    <TextField
                      fullWidth
                      label={t("adminVendor.saleDescription")}
                      value={purchaseDescription}
                      onChange={(e) => setPurchaseDescription(e.target.value)}
                    />
                    <TextField
                      select
                      fullWidth
                      label={t("vendor.installments")}
                      value={purchaseNumInstallments}
                      onChange={(e) => setPurchaseNumInstallments(Number(e.target.value))}
                    >
                      {Array.from(
                        { length: purchaseSelectedCreditLine.maxInstallments },
                        (_, i) => i + 1
                      ).map((n) => (
                        <MenuItem key={n} value={n}>
                          {n}
                        </MenuItem>
                      ))}
                    </TextField>
                    <TextField
                      select
                      fullWidth
                      label={t("vendor.amortizationSystem")}
                      value={purchaseAmortType}
                      onChange={(e) => setPurchaseAmortType(e.target.value as "french" | "german")}
                    >
                      <MenuItem value="french">{t("loans.frenchFull")}</MenuItem>
                      <MenuItem value="german">{t("loans.germanFull")}</MenuItem>
                    </TextField>
                  </>
                )}
              </>
            )}
          </Box>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={handleClosePurchase}>{t("common.cancel")}</Button>
          <Button
            variant="contained"
            onClick={handleRecordPurchase}
            disabled={purchaseSubmitting || !purchaseSelectedClient || !purchaseSelectedCreditLine || !purchaseAmount || !purchaseDescription}
          >
            {purchaseSubmitting ? t("common.creating") : t("common.confirm")}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default VendorDetail;
