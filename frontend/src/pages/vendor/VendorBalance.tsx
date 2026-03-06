import React, { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import {
  Box,
  Typography,
  Card,
  CardContent,
  Chip,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
  IconButton,
  Tooltip,
} from "@mui/material";
import {
  AccountBalance as AccountIcon,
  Add as AddIcon,
  Download as DownloadIcon,
} from "@mui/icons-material";
import { format } from "date-fns";
import {
  getVendorAccount,
  getVendorMovements,
  vendorRequestWithdrawal,
  vendorGetWithdrawals,
  downloadVendorPaymentReceipt,
} from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import type { VendorMovement, WithdrawalRequest } from "../../api/types";
import { useNotification } from "../../contexts/NotificationContext";
import { getErrorMessage } from "../../api/errorUtils";

const statusColors: Record<string, "warning" | "success" | "error" | "info" | "default"> = {
  pending: "warning",
  approved: "info",
  paid: "success",
  rejected: "error",
};

const VendorBalance: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [withdrawalPage, setWithdrawalPage] = useState(0);
  const [openWithdrawal, setOpenWithdrawal] = useState(false);
  const [withdrawalAmount, setWithdrawalAmount] = useState("");
  const [withdrawalMethod, setWithdrawalMethod] = useState("transfer");
  const { showSuccess, showError } = useNotification();

  const { data: account, isLoading: accountLoading } = useQuery({
    queryKey: ["vendor-account"],
    queryFn: getVendorAccount,
  });

  const { data: movementsData, isLoading: movementsLoading } = useQuery({
    queryKey: ["vendor-movements", page, pageSize],
    queryFn: () => getVendorMovements(page + 1, pageSize),
  });

  const { data: withdrawalsData, isLoading: withdrawalsLoading } = useQuery({
    queryKey: ["vendor-withdrawals", withdrawalPage],
    queryFn: () => vendorGetWithdrawals(withdrawalPage + 1),
  });

  const withdrawalMutation = useMutation({
    mutationFn: vendorRequestWithdrawal,
    onSuccess: () => {
      setOpenWithdrawal(false);
      setWithdrawalAmount("");
      showSuccess(t("vendor.withdrawalSuccess"));
      queryClient.invalidateQueries({ queryKey: ["vendor-withdrawals"] });
    },
    onError: (err: unknown) => {
      showError(getErrorMessage(err, t("vendor.withdrawalError")));
    },
  });

  const handleRequestWithdrawal = () => {
    withdrawalMutation.mutate({ amount: withdrawalAmount, method: withdrawalMethod });
  };

  const handleDownloadReceipt = async (paymentId: string) => {
    try {
      const blob = await downloadVendorPaymentReceipt(paymentId);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `receipt-${paymentId.slice(0, 8)}.pdf`;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (err) {
      showError(getErrorMessage(err, t("common.error")));
    }
  };

  const movementColumns: Column<VendorMovement>[] = [
    {
      id: "createdAt",
      label: t("common.date"),
      minWidth: 130,
      render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy HH:mm"),
    },
    {
      id: "type",
      label: t("common.type"),
      render: (row) => {
        const numAmount = parseFloat(row.amount);
        return (
          <Chip
            label={row.type}
            size="small"
            color={numAmount > 0 ? "success" : "error"}
            variant="outlined"
          />
        );
      },
    },
    {
      id: "amount",
      label: t("common.amount"),
      align: "right",
      render: (row) => {
        const numAmount = parseFloat(row.amount);
        return (
          <MoneyDisplay
            amount={numAmount}
            color={numAmount > 0 ? "success.main" : "error.main"}
            fontWeight={500}
          />
        );
      },
    },
    {
      id: "balanceAfter",
      label: t("vendor.balanceAfter"),
      align: "right",
      render: (row) => (
        <MoneyDisplay amount={parseFloat(row.balanceAfter)} fontWeight={500} />
      ),
    },
    {
      id: "description",
      label: t("common.description"),
      minWidth: 200,
    },
  ];

  const withdrawalColumns: Column<WithdrawalRequest>[] = [
    {
      id: "requestedAt",
      label: t("common.date"),
      minWidth: 130,
      render: (row) => format(new Date(row.requestedAt), "dd/MM/yyyy HH:mm"),
    },
    {
      id: "amount",
      label: t("common.amount"),
      align: "right",
      render: (row) => <MoneyDisplay amount={parseFloat(row.amount)} fontWeight={500} />,
    },
    {
      id: "method",
      label: t("payments.method"),
      render: (row) => t(`payments.${row.method}`),
    },
    {
      id: "status",
      label: t("common.status"),
      render: (row) => (
        <Chip
          label={t(`status.${row.status}`)}
          size="small"
          color={statusColors[row.status] || "default"}
        />
      ),
    },
    {
      id: "rejectionReason",
      label: t("adminVendor.rejectionReason"),
      render: (row) => row.rejectionReason || "-",
    },
    {
      id: "actions",
      label: "",
      render: (row) =>
        row.status === "paid" && row.paymentId ? (
          <Tooltip title={t("vendor.downloadReceipt")}>
            <IconButton size="small" onClick={() => handleDownloadReceipt(row.paymentId!)}>
              <DownloadIcon />
            </IconButton>
          </Tooltip>
        ) : null,
    },
  ];

  return (
    <Box>
      <Typography variant="h4" fontWeight={700} gutterBottom>
        {t("vendor.balance")}
      </Typography>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box display="flex" alignItems="center" justifyContent="space-between">
            <Box display="flex" alignItems="center" gap={2}>
              <Box
                sx={{
                  width: 64,
                  height: 64,
                  borderRadius: 2,
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  bgcolor: "primary.main",
                  color: "white",
                }}
              >
                <AccountIcon sx={{ fontSize: 32 }} />
              </Box>
              <Box>
                <Typography variant="body2" color="text.secondary">
                  {t("vendor.currentBalance")}
                </Typography>
                {accountLoading ? (
                  <Typography variant="h4">{t("common.loading")}</Typography>
                ) : (
                  <MoneyDisplay
                    amount={parseFloat(account?.balance || "0")}
                    variant="h4"
                    fontWeight={700}
                    color="primary.main"
                  />
                )}
              </Box>
            </Box>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => setOpenWithdrawal(true)}
            >
              {t("vendor.requestWithdrawal")}
            </Button>
          </Box>
        </CardContent>
      </Card>

      <Typography variant="h6" gutterBottom>
        {t("vendor.withdrawals")}
      </Typography>
      <DataTable
        columns={withdrawalColumns}
        rows={withdrawalsData?.data || []}
        total={withdrawalsData?.total || 0}
        page={withdrawalPage}
        pageSize={20}
        onPageChange={setWithdrawalPage}
        onPageSizeChange={() => {}}
        loading={withdrawalsLoading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("vendor.noWithdrawals")}
      />

      <Typography variant="h6" gutterBottom sx={{ mt: 3 }}>
        {t("vendor.movements")}
      </Typography>
      <DataTable
        columns={movementColumns}
        rows={movementsData?.data || []}
        total={movementsData?.total || 0}
        page={page}
        pageSize={pageSize}
        onPageChange={setPage}
        onPageSizeChange={(size) => {
          setPageSize(size);
          setPage(0);
        }}
        loading={movementsLoading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("vendor.noMovements")}
      />

      <Dialog open={openWithdrawal} onClose={() => setOpenWithdrawal(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("vendor.requestWithdrawal")}</DialogTitle>
        <DialogContent>
          <Box display="flex" flexDirection="column" gap={2} mt={1}>
            <TextField
              label={t("vendor.withdrawalAmount")}
              value={withdrawalAmount}
              onChange={(e) => setWithdrawalAmount(e.target.value)}
              type="number"
            />
            <TextField
              label={t("vendor.withdrawalMethod")}
              select
              value={withdrawalMethod}
              onChange={(e) => setWithdrawalMethod(e.target.value)}
            >
              <MenuItem value="cash">{t("payments.cash")}</MenuItem>
              <MenuItem value="transfer">{t("payments.transfer")}</MenuItem>
            </TextField>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenWithdrawal(false)}>{t("common.cancel")}</Button>
          <Button
            variant="contained"
            onClick={handleRequestWithdrawal}
            disabled={!withdrawalAmount || withdrawalMutation.isPending}
          >
            {t("common.confirm")}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default VendorBalance;
