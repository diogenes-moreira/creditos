import React, { useState } from "react";
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
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
  Snackbar,
} from "@mui/material";
import { ArrowBack as BackIcon, Payment as PaymentIcon } from "@mui/icons-material";
import { format } from "date-fns";
import { getLoan, recordPayment } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import StatusBadge from "../../components/StatusBadge";
import type { Installment } from "../../api/types";

const LoanDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { t } = useTranslation();
  const [payDialog, setPayDialog] = useState(false);
  const [selectedInstallment, setSelectedInstallment] = useState<Installment | null>(null);
  const [paymentMethod, setPaymentMethod] = useState("transfer");
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" as "success" | "error" });

  const { data: loan, isLoading } = useQuery({
    queryKey: ["loan", id],
    queryFn: () => getLoan(id!),
    enabled: !!id,
  });

  const payMutation = useMutation({
    mutationFn: (data: { installmentNumber: number; amount: number; method: string }) =>
      recordPayment(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["loan", id] });
      setPayDialog(false);
      setSnackbar({ open: true, message: t("loans.paymentSuccess"), severity: "success" });
    },
    onError: () => {
      setSnackbar({ open: true, message: t("loans.paymentError"), severity: "error" });
    },
  });

  const handlePay = (installment: Installment) => {
    setSelectedInstallment(installment);
    setPayDialog(true);
  };

  const confirmPayment = () => {
    if (!selectedInstallment) return;
    payMutation.mutate({
      installmentNumber: selectedInstallment.number,
      amount: selectedInstallment.total,
      method: paymentMethod,
    });
  };

  const columns: Column<Installment>[] = [
    { id: "number", label: "#", align: "center" },
    {
      id: "dueDate",
      label: t("loans.dueDate"),
      render: (row) => format(new Date(row.dueDate), "dd/MM/yyyy"),
    },
    {
      id: "principal",
      label: t("loans.capital"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.principal} />,
    },
    {
      id: "interest",
      label: t("loans.interest"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.interest} />,
    },
    {
      id: "total",
      label: t("loans.installmentTotal"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.total} fontWeight={600} />,
    },
    {
      id: "balance",
      label: t("account.balance"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.balance} />,
    },
    {
      id: "status",
      label: t("common.status"),
      render: (row) => <StatusBadge status={row.status} />,
    },
    {
      id: "actions",
      label: t("common.actions"),
      align: "center",
      render: (row) =>
        row.status === "pending" || row.status === "overdue" ? (
          <Button
            size="small"
            variant="contained"
            startIcon={<PaymentIcon />}
            onClick={() => handlePay(row)}
          >
            {t("loans.pay")}
          </Button>
        ) : row.paidAt ? (
          <Typography variant="caption" color="text.secondary">
            {t("loans.paidOn")} {format(new Date(row.paidAt), "dd/MM/yyyy")}
          </Typography>
        ) : null,
    },
  ];

  if (isLoading) {
    return (
      <Box>
        <Typography>{t("common.loading")}</Typography>
      </Box>
    );
  }

  if (!loan) {
    return (
      <Alert severity="error">{t("loans.notFound")}</Alert>
    );
  }

  return (
    <Box>
      <Button startIcon={<BackIcon />} onClick={() => navigate("/loans")} sx={{ mb: 2 }}>
        {t("loans.backToLoans")}
      </Button>

      <Typography variant="h4" gutterBottom>
        {t("loans.loanNumber")} #{loan.id.slice(0, 8)}
      </Typography>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Grid container spacing={3}>
            <Grid item xs={12} sm={6} md={3}>
              <Typography variant="body2" color="text.secondary">{t("loans.requestedAmount")}</Typography>
              <MoneyDisplay amount={loan.amount} variant="h6" fontWeight={600} />
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <Typography variant="body2" color="text.secondary">{t("loans.totalPayment")}</Typography>
              <MoneyDisplay amount={loan.totalAmount} variant="h6" fontWeight={600} />
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <Typography variant="body2" color="text.secondary">{t("loans.installments")}</Typography>
              <Typography variant="h6" fontWeight={600}>{loan.installments}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <Typography variant="body2" color="text.secondary">{t("common.status")}</Typography>
              <Box mt={0.5}><StatusBadge status={loan.status} size="medium" /></Box>
            </Grid>
          </Grid>
          <Divider sx={{ my: 2 }} />
          <Grid container spacing={3}>
            <Grid item xs={12} sm={4}>
              <Typography variant="body2" color="text.secondary">{t("loans.system")}</Typography>
              <Typography>{loan.amortizationType === "french" ? t("loans.french") : t("loans.german")}</Typography>
            </Grid>
            <Grid item xs={12} sm={4}>
              <Typography variant="body2" color="text.secondary">{t("loans.interestRate")}</Typography>
              <Typography>{loan.interestRate}%</Typography>
            </Grid>
            <Grid item xs={12} sm={4}>
              <Typography variant="body2" color="text.secondary">{t("loans.applicationDate")}</Typography>
              <Typography>{format(new Date(loan.createdAt), "dd/MM/yyyy")}</Typography>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      <Typography variant="h6" gutterBottom>{t("loans.installmentSchedule")}</Typography>
      <DataTable
        columns={columns}
        rows={loan.schedule || []}
        keyExtractor={(row) => String(row.number)}
        emptyMessage={t("loans.noInstallments")}
      />

      <Dialog open={payDialog} onClose={() => setPayDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("loans.registerPayment")} - {t("payments.installmentNumber")}{selectedInstallment?.number}</DialogTitle>
        <DialogContent>
          <Box mt={1}>
            <Typography variant="body1" gutterBottom>
              {t("loans.amountToPay")}: <MoneyDisplay amount={selectedInstallment?.total || 0} fontWeight={600} />
            </Typography>
            <TextField
              select
              fullWidth
              label={t("loans.paymentMethod")}
              value={paymentMethod}
              onChange={(e) => setPaymentMethod(e.target.value)}
              sx={{ mt: 2 }}
            >
              <MenuItem value="transfer">{t("payments.transfer")}</MenuItem>
              <MenuItem value="mercadopago">{t("payments.mercadoPago")}</MenuItem>
              <MenuItem value="cash">{t("payments.cash")}</MenuItem>
            </TextField>
          </Box>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setPayDialog(false)}>{t("common.cancel")}</Button>
          <Button
            variant="contained"
            onClick={confirmPayment}
            disabled={payMutation.isPending}
          >
            {payMutation.isPending ? t("common.processing") : t("loans.confirmPayment")}
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

export default LoanDetail;
