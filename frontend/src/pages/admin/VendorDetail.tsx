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
  Alert,
  CircularProgress,
} from "@mui/material";
import { ArrowBack as BackIcon } from "@mui/icons-material";
import {
  adminGetVendor,
  adminActivateVendor,
  adminDeactivateVendor,
  adminGetVendorPurchases,
  adminGetVendorPayments,
  adminRecordVendorPayment,
} from "../../api/endpoints";
import type { Vendor, Purchase, VendorPayment } from "../../api/types";

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
  const [openPayment, setOpenPayment] = useState(false);
  const [payAmount, setPayAmount] = useState("");
  const [payMethod, setPayMethod] = useState("transfer");
  const [payRef, setPayRef] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

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

  useEffect(() => { fetchVendor(); }, [id]);
  useEffect(() => { if (id) fetchPurchases(); }, [id, purchasePage]);
  useEffect(() => { if (id) fetchPayments(); }, [id, paymentPage]);

  const handleToggleActive = async () => {
    try {
      if (vendor?.isActive) {
        await adminDeactivateVendor(id!);
        setSuccess(t("adminVendor.deactivated"));
      } else {
        await adminActivateVendor(id!);
        setSuccess(t("adminVendor.activated"));
      }
      fetchVendor();
    } catch { setError(t("admin.processError")); }
  };

  const handleRecordPayment = async () => {
    try {
      await adminRecordVendorPayment(id!, { vendorId: id!, amount: payAmount, method: payMethod, reference: payRef });
      setOpenPayment(false);
      setPayAmount(""); setPayRef("");
      setSuccess(t("adminVendor.paymentSuccess"));
      fetchPayments();
      fetchVendor();
    } catch { setError(t("adminVendor.paymentError")); }
  };

  if (loading) return <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>;
  if (!vendor) return <Typography>{t("adminVendor.vendorNotFound")}</Typography>;

  return (
    <Box>
      <Button startIcon={<BackIcon />} onClick={() => navigate("/admin/vendors")} sx={{ mb: 2 }}>
        {t("adminVendor.backToVendors")}
      </Button>

      <Typography variant="h4" fontWeight={700} gutterBottom>{t("adminVendor.vendorDetail")}</Typography>

      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError("")}>{error}</Alert>}
      {success && <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess("")}>{success}</Alert>}

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
            </Box>
          </Paper>
        </Grid>
      </Grid>

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
            </TableRow>
          </TableHead>
          <TableBody>
            {payments.map((p) => (
              <TableRow key={p.id}>
                <TableCell>{new Date(p.createdAt).toLocaleDateString()}</TableCell>
                <TableCell>${p.amount}</TableCell>
                <TableCell>{t(`payments.${p.method}`)}</TableCell>
                <TableCell>{p.reference}</TableCell>
              </TableRow>
            ))}
            {payments.length === 0 && <TableRow><TableCell colSpan={4} align="center">{t("payments.noPayments")}</TableCell></TableRow>}
          </TableBody>
        </Table>
        <TablePagination component="div" count={paymentTotal} page={paymentPage} onPageChange={(_, p) => setPaymentPage(p)} rowsPerPage={20} rowsPerPageOptions={[20]} />
      </TableContainer>

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
    </Box>
  );
};

export default VendorDetail;
