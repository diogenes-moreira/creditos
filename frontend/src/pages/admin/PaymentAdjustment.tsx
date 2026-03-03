import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Box,
  Typography,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Snackbar,
  Alert,
} from "@mui/material";
import { Edit as EditIcon } from "@mui/icons-material";
import { format } from "date-fns";
import { getPayments, adminAdjustPayment } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import StatusBadge from "../../components/StatusBadge";
import type { Payment } from "../../api/types";

const PaymentAdjustment: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [adjustDialog, setAdjustDialog] = useState<Payment | null>(null);
  const [adjustAmount, setAdjustAmount] = useState("");
  const [adjustReason, setAdjustReason] = useState("");
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" as "success" | "error" });

  const { data: payments, isLoading } = useQuery({
    queryKey: ["admin-payments"],
    queryFn: getPayments,
  });

  const adjustMutation = useMutation({
    mutationFn: ({ id, amount, reason }: { id: string; amount: number; reason: string }) =>
      adminAdjustPayment(id, { amount, reason }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-payments"] });
      setAdjustDialog(null);
      setAdjustAmount("");
      setAdjustReason("");
      setSnackbar({ open: true, message: t("payments.paymentAdjusted"), severity: "success" });
    },
    onError: () => setSnackbar({ open: true, message: t("payments.adjustError"), severity: "error" }),
  });

  const handleAdjust = () => {
    if (!adjustDialog) return;
    adjustMutation.mutate({
      id: adjustDialog.id,
      amount: parseFloat(adjustAmount),
      reason: adjustReason,
    });
  };

  const columns: Column<Payment>[] = [
    {
      id: "paidAt",
      label: t("common.date"),
      minWidth: 130,
      render: (row) => format(new Date(row.paidAt), "dd/MM/yyyy HH:mm"),
    },
    {
      id: "loanId",
      label: t("payments.loan"),
      render: (row) => `#${row.loanId.slice(0, 8)}`,
    },
    {
      id: "installmentNumber",
      label: t("payments.installmentNumber"),
      align: "center",
    },
    {
      id: "amount",
      label: t("payments.originalAmount"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.amount} fontWeight={500} />,
    },
    {
      id: "adjustedAmount",
      label: t("payments.adjustedAmount"),
      align: "right",
      render: (row) =>
        row.adjustedAmount ? (
          <MoneyDisplay amount={row.adjustedAmount} color="warning.main" fontWeight={500} />
        ) : (
          <Typography variant="body2" color="text.secondary">-</Typography>
        ),
    },
    {
      id: "method",
      label: t("payments.method"),
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
      render: (row) => (
        <Button
          size="small"
          variant="outlined"
          startIcon={<EditIcon />}
          onClick={() => {
            setAdjustDialog(row);
            setAdjustAmount(String(row.amount));
            setAdjustReason(row.adjustmentReason || "");
          }}
        >
          {t("payments.adjust")}
        </Button>
      ),
    },
  ];

  return (
    <Box>
      <Typography variant="h4" gutterBottom>{t("nav.paymentAdjustment")}</Typography>
      <Typography variant="body1" color="text.secondary" mb={3}>
        {t("payments.managePayments")}
      </Typography>

      <DataTable
        columns={columns}
        rows={payments || []}
        loading={isLoading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("payments.noPayments")}
      />

      <Dialog open={!!adjustDialog} onClose={() => setAdjustDialog(null)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {t("payments.adjustPayment")} - {t("payments.installmentNumber")}{adjustDialog?.installmentNumber}
        </DialogTitle>
        <DialogContent>
          <Box mt={1} display="flex" flexDirection="column" gap={2}>
            <Typography variant="body2" color="text.secondary">
              {t("payments.originalAmount")}: <MoneyDisplay amount={adjustDialog?.amount || 0} fontWeight={600} />
            </Typography>
            <TextField
              fullWidth
              type="number"
              label={t("payments.newAmount")}
              value={adjustAmount}
              onChange={(e) => setAdjustAmount(e.target.value)}
            />
            <TextField
              fullWidth
              multiline
              rows={3}
              label={t("payments.adjustReason")}
              value={adjustReason}
              onChange={(e) => setAdjustReason(e.target.value)}
              placeholder={t("payments.adjustReasonPlaceholder")}
            />
          </Box>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setAdjustDialog(null)}>{t("common.cancel")}</Button>
          <Button
            variant="contained"
            onClick={handleAdjust}
            disabled={!adjustAmount || !adjustReason || adjustMutation.isPending}
          >
            {adjustMutation.isPending ? t("common.processing") : t("payments.applyAdjust")}
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

export default PaymentAdjustment;
