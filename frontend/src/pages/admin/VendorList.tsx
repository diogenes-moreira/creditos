import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  TextField,
  InputAdornment,
  Chip,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Grid,
} from "@mui/material";
import { Search as SearchIcon, Add as AddIcon } from "@mui/icons-material";
import { adminGetVendors, adminRegisterVendor } from "../../api/endpoints";
import type { Vendor, RegisterVendorRequest } from "../../api/types";
import { useNotification } from "../../contexts/NotificationContext";
import { getErrorMessage } from "../../api/errorUtils";
import LocationSelector from "../../components/LocationSelector";

const VendorList: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [vendors, setVendors] = useState<Vendor[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(20);
  const [search, setSearch] = useState("");
  const [openRegister, setOpenRegister] = useState(false);
  const { showError } = useNotification();
  const [formData, setFormData] = useState<RegisterVendorRequest>({
    email: "", password: "", businessName: "", cuit: "", phone: "", address: "", country: "Argentina", city: "", province: "",
  });

  const fetchVendors = async () => {
    try {
      const res = await adminGetVendors(page + 1, rowsPerPage, search);
      setVendors(res.data || []);
      setTotal(res.total || 0);
    } catch {
      setVendors([]);
    }
  };

  useEffect(() => { fetchVendors(); }, [page, rowsPerPage, search]);

  const handleRegister = async () => {
    try {
      await adminRegisterVendor(formData);
      setOpenRegister(false);
      setFormData({ email: "", password: "", businessName: "", cuit: "", phone: "", address: "", country: "Argentina", city: "", province: "" });
      fetchVendors();
    } catch (err) {
      showError(getErrorMessage(err, t("adminVendor.registerError")));
    }
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Box>
          <Typography variant="h4" fontWeight={700} gutterBottom>{t("adminVendor.title")}</Typography>
          <Typography color="text.secondary">{t("adminVendor.description")}</Typography>
        </Box>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setOpenRegister(true)}>
          {t("adminVendor.registerVendor")}
        </Button>
      </Box>

      <Paper sx={{ mb: 2, p: 2 }}>
        <TextField
          fullWidth size="small" placeholder={t("admin.searchPlaceholder")}
          value={search} onChange={(e) => { setSearch(e.target.value); setPage(0); }}
          InputProps={{ startAdornment: <InputAdornment position="start"><SearchIcon /></InputAdornment> }}
        />
      </Paper>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>{t("adminVendor.businessName")}</TableCell>
              <TableCell>{t("adminVendor.cuit")}</TableCell>
              <TableCell>{t("adminVendor.city")}</TableCell>
              <TableCell>{t("common.status")}</TableCell>
              <TableCell>{t("common.date")}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {vendors.length === 0 ? (
              <TableRow><TableCell colSpan={5} align="center">{t("adminVendor.noVendors")}</TableCell></TableRow>
            ) : vendors.map((v) => (
              <TableRow key={v.id} hover sx={{ cursor: "pointer" }} onClick={() => navigate(`/admin/vendors/${v.id}`)}>
                <TableCell>{v.businessName}</TableCell>
                <TableCell>{v.cuit}</TableCell>
                <TableCell>{v.city}</TableCell>
                <TableCell>
                  <Chip label={v.isActive ? t("status.active") : t("status.inactive")} color={v.isActive ? "success" : "default"} size="small" />
                </TableCell>
                <TableCell>{new Date(v.createdAt).toLocaleDateString()}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        <TablePagination
          component="div" count={total} page={page} onPageChange={(_, p) => setPage(p)}
          rowsPerPage={rowsPerPage} onRowsPerPageChange={(e) => { setRowsPerPage(parseInt(e.target.value)); setPage(0); }}
          rowsPerPageOptions={[10, 20, 50]}
        />
      </TableContainer>

      <Dialog open={openRegister} onClose={() => setOpenRegister(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("adminVendor.registerVendor")}</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} mt={1}>
            <TextField label={t("adminVendor.email")} value={formData.email} onChange={(e) => setFormData({ ...formData, email: e.target.value })} />
            <TextField label={t("adminVendor.password")} type="password" value={formData.password} onChange={(e) => setFormData({ ...formData, password: e.target.value })} />
            <TextField label={t("adminVendor.businessName")} value={formData.businessName} onChange={(e) => setFormData({ ...formData, businessName: e.target.value })} />
            <TextField label={t("adminVendor.cuit")} value={formData.cuit} onChange={(e) => setFormData({ ...formData, cuit: e.target.value })} />
            <TextField label={t("adminVendor.phone")} value={formData.phone} onChange={(e) => setFormData({ ...formData, phone: e.target.value })} />
            <TextField label={t("adminVendor.address")} value={formData.address} onChange={(e) => setFormData({ ...formData, address: e.target.value })} />
            <Grid container spacing={2}>
              <LocationSelector
                country={formData.country}
                province={formData.province}
                city={formData.city}
                onChange={(field, value) => setFormData(prev => ({ ...prev, [field]: value }))}
              />
            </Grid>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenRegister(false)}>{t("common.cancel")}</Button>
          <Button variant="contained" onClick={handleRegister}>{t("common.create")}</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default VendorList;
