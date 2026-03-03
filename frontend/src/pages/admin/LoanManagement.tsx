import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Box,
  Typography,
  Snackbar,
  Alert,
  Button,
  Chip,
} from "@mui/material";
import {
  Check as ApproveIcon,
  Send as DisburseIcon,
  Cancel as CancelIcon,
  PaymentOutlined as PrepayIcon,
} from "@mui/icons-material";
import { format } from "date-fns";
import {
  adminGetPendingLoans,
  adminApproveLoan,
  adminDisburseLoan,
  adminCancelLoan,
  adminPrepayLoan,
} from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import StatusBadge from "../../components/StatusBadge";
import ConfirmDialog from "../../components/ConfirmDialog";
import type { Loan } from "../../api/types";

type ActionType = "approve" | "disburse" | "cancel" | "prepay";

const LoanManagement: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [confirmAction, setConfirmAction] = useState<{ type: ActionType; loan: Loan } | null>(null);
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" as "success" | "error" });

  const { data: loans, isLoading } = useQuery({
    queryKey: ["admin-pending-loans"],
    queryFn: adminGetPendingLoans,
  });

  const createMutation = (fn: (id: string) => Promise<Loan>, successMsg: string) =>
    useMutation({
      mutationFn: fn,
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ["admin-pending-loans"] });
        setConfirmAction(null);
        setSnackbar({ open: true, message: successMsg, severity: "success" });
      },
      onError: () => {
        setConfirmAction(null);
        setSnackbar({ open: true, message: t("admin.processError"), severity: "error" });
      },
    });

  const approveMutation = createMutation(adminApproveLoan, t("admin.loanApproved"));
  const disburseMutation = createMutation(adminDisburseLoan, t("admin.loanDisbursed"));
  const cancelMutation = createMutation(adminCancelLoan, t("admin.loanCancelled"));
  const prepayMutation = createMutation(adminPrepayLoan, t("admin.earlyCancelProcessed"));

  const handleConfirm = () => {
    if (!confirmAction) return;
    const { type, loan } = confirmAction;
    switch (type) {
      case "approve": approveMutation.mutate(loan.id); break;
      case "disburse": disburseMutation.mutate(loan.id); break;
      case "cancel": cancelMutation.mutate(loan.id); break;
      case "prepay": prepayMutation.mutate(loan.id); break;
    }
  };

  const actionLabels: Record<ActionType, { label: string; color: "success" | "primary" | "error" | "warning"; message: string }> = {
    approve: { label: t("admin.approve"), color: "success", message: t("confirm.approveLoan") },
    disburse: { label: t("admin.disburse"), color: "primary", message: t("confirm.disburseLoan") },
    cancel: { label: t("common.cancel"), color: "error", message: t("confirm.cancelLoan") },
    prepay: { label: t("admin.earlyCancellation"), color: "warning", message: t("confirm.earlyCancelLoan") },
  };

  const columns: Column<Loan>[] = [
    {
      id: "id",
      label: "ID",
      minWidth: 100,
      render: (row) => `#${row.id.slice(0, 8)}`,
    },
    {
      id: "clientName",
      label: t("admin.client"),
      minWidth: 150,
      render: (row) => row.clientName || row.clientId.slice(0, 8),
    },
    {
      id: "amount",
      label: t("common.amount"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.amount} fontWeight={500} />,
    },
    {
      id: "installments",
      label: t("loans.installments"),
      align: "center",
    },
    {
      id: "amortizationType",
      label: t("loans.system"),
      render: (row) => (
        <Chip
          label={row.amortizationType === "french" ? t("loans.french") : t("loans.german")}
          size="small"
          variant="outlined"
        />
      ),
    },
    {
      id: "interestRate",
      label: t("admin.rate"),
      align: "center",
      render: (row) => `${row.interestRate}%`,
    },
    {
      id: "status",
      label: t("common.status"),
      render: (row) => <StatusBadge status={row.status} />,
    },
    {
      id: "createdAt",
      label: t("common.date"),
      render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy"),
    },
    {
      id: "actions",
      label: t("common.actions"),
      minWidth: 280,
      render: (row) => (
        <Box display="flex" gap={0.5} flexWrap="wrap">
          {row.status === "pending" && (
            <>
              <Button size="small" variant="contained" color="success" startIcon={<ApproveIcon />}
                onClick={() => setConfirmAction({ type: "approve", loan: row })}>
                {t("admin.approve")}
              </Button>
              <Button size="small" variant="outlined" color="error" startIcon={<CancelIcon />}
                onClick={() => setConfirmAction({ type: "cancel", loan: row })}>
                {t("common.cancel")}
              </Button>
            </>
          )}
          {row.status === "approved" && (
            <Button size="small" variant="contained" color="primary" startIcon={<DisburseIcon />}
              onClick={() => setConfirmAction({ type: "disburse", loan: row })}>
              {t("admin.disburse")}
            </Button>
          )}
          {(row.status === "active" || row.status === "disbursed") && (
            <>
              <Button size="small" variant="outlined" color="warning" startIcon={<PrepayIcon />}
                onClick={() => setConfirmAction({ type: "prepay", loan: row })}>
                {t("admin.prepay")}
              </Button>
              <Button size="small" variant="outlined" color="error" startIcon={<CancelIcon />}
                onClick={() => setConfirmAction({ type: "cancel", loan: row })}>
                {t("common.cancel")}
              </Button>
            </>
          )}
        </Box>
      ),
    },
  ];

  const currentAction = confirmAction ? actionLabels[confirmAction.type] : null;

  return (
    <Box>
      <Typography variant="h4" gutterBottom>{t("admin.loanManagement")}</Typography>
      <Typography variant="body1" color="text.secondary" mb={3}>
        {t("admin.pendingAction")}
      </Typography>

      <DataTable
        columns={columns}
        rows={loans || []}
        loading={isLoading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("admin.noPendingLoans")}
      />

      {confirmAction && currentAction && (
        <ConfirmDialog
          open={true}
          title={`${currentAction.label} Prestamo #${confirmAction.loan.id.slice(0, 8)}`}
          message={currentAction.message}
          confirmLabel={currentAction.label}
          confirmColor={currentAction.color}
          onConfirm={handleConfirm}
          onCancel={() => setConfirmAction(null)}
          loading={
            approveMutation.isPending ||
            disburseMutation.isPending ||
            cancelMutation.isPending ||
            prepayMutation.isPending
          }
        />
      )}

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

export default LoanManagement;
