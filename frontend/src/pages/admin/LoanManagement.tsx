import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Box,
  Typography,
  Alert,
  Button,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  MenuItem,
  InputAdornment,
  List,
  ListItemButton,
  ListItemText,
} from "@mui/material";
import {
  Check as ApproveIcon,
  Send as DisburseIcon,
  Cancel as CancelIcon,
  PaymentOutlined as PrepayIcon,
  Add as AddIcon,
  Search as SearchIcon,
  AccountBalanceWallet as WithdrawalIcon,
} from "@mui/icons-material";
import { format } from "date-fns";
import {
  adminGetPendingLoans,
  adminApproveLoan,
  adminDisburseLoan,
  adminCancelLoan,
  adminPrepayLoan,
  adminCreateLoan,
  adminCreateWithdrawal,
  adminSearchClients,
  adminGetClientCreditLines,
} from "../../api/endpoints";
import { useNotification } from "../../contexts/NotificationContext";
import { getErrorMessage } from "../../api/errorUtils";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import StatusBadge from "../../components/StatusBadge";
import ConfirmDialog from "../../components/ConfirmDialog";
import type { Loan, Client, CreditLine } from "../../api/types";

type ActionType = "approve" | "disburse" | "cancel" | "prepay";

const LoanManagement: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [confirmAction, setConfirmAction] = useState<{ type: ActionType; loan: Loan } | null>(null);
  const { showSuccess, showError } = useNotification();

  // Create loan / withdrawal dialog state
  const [createOpen, setCreateOpen] = useState(false);
  const [dialogMode, setDialogMode] = useState<"loan" | "withdrawal">("loan");
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<Client[]>([]);
  const [selectedClient, setSelectedClient] = useState<Client | null>(null);
  const [clientCreditLines, setClientCreditLines] = useState<CreditLine[]>([]);
  const [selectedCreditLine, setSelectedCreditLine] = useState<CreditLine | null>(null);
  const [loanAmount, setLoanAmount] = useState("");
  const [loanInstallments, setLoanInstallments] = useState(12);
  const [loanAmortType, setLoanAmortType] = useState<"french" | "german">("french");

  const { data: loans, isLoading } = useQuery({
    queryKey: ["admin-pending-loans"],
    queryFn: adminGetPendingLoans,
  });

  const mutationOptions = (successMsg: string, errorMsg: string) => ({
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-pending-loans"] });
      setConfirmAction(null);
      showSuccess(successMsg);
    },
    onError: (err: unknown) => {
      setConfirmAction(null);
      showError(getErrorMessage(err, errorMsg));
    },
  });

  const approveMutation = useMutation({ mutationFn: adminApproveLoan, ...mutationOptions(t("admin.loanApproved"), t("admin.processError")) });
  const disburseMutation = useMutation({ mutationFn: adminDisburseLoan, ...mutationOptions(t("admin.loanDisbursed"), t("admin.processError")) });
  const cancelMutation = useMutation({ mutationFn: adminCancelLoan, ...mutationOptions(t("admin.loanCancelled"), t("admin.processError")) });
  const prepayMutation = useMutation({ mutationFn: ({ id, amount }: { id: string; amount: string }) => adminPrepayLoan(id, amount), ...mutationOptions(t("admin.earlyCancelProcessed"), t("admin.processError")) });

  const createLoanMutation = useMutation({
    mutationFn: adminCreateLoan,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-pending-loans"] });
      handleCloseCreate();
      showSuccess(t("admin.loanCreated"));
    },
    onError: (err: unknown) => {
      showError(getErrorMessage(err, t("admin.loanCreateError")));
    },
  });

  const createWithdrawalMutation = useMutation({
    mutationFn: adminCreateWithdrawal,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-pending-loans"] });
      handleCloseCreate();
      showSuccess(t("admin.withdrawalCreated"));
    },
    onError: (err: unknown) => {
      showError(getErrorMessage(err, t("admin.withdrawalCreateError")));
    },
  });

  const handleConfirm = () => {
    if (!confirmAction) return;
    const { type, loan } = confirmAction;
    switch (type) {
      case "approve": approveMutation.mutate(loan.id); break;
      case "disburse": disburseMutation.mutate(loan.id); break;
      case "cancel": cancelMutation.mutate(loan.id); break;
      case "prepay": prepayMutation.mutate({ id: loan.id, amount: loan.totalRemaining }); break;
    }
  };

  const handleCloseCreate = () => {
    setCreateOpen(false);
    setDialogMode("loan");
    setSearchQuery("");
    setSearchResults([]);
    setSelectedClient(null);
    setClientCreditLines([]);
    setSelectedCreditLine(null);
    setLoanAmount("");
    setLoanInstallments(12);
    setLoanAmortType("french");
  };

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;
    const results = await adminSearchClients(searchQuery.trim());
    setSearchResults(results);
  };

  const handleSelectClient = async (client: Client) => {
    setSelectedClient(client);
    setSearchResults([]);
    const lines = await adminGetClientCreditLines(client.id);
    const approved = lines.filter((cl) => cl.status === "approved");
    setClientCreditLines(approved);
    if (approved.length === 1) {
      setSelectedCreditLine(approved[0]);
    } else {
      setSelectedCreditLine(null);
    }
  };

  const handleCreateLoan = () => {
    if (!selectedClient || !selectedCreditLine || !loanAmount) return;
    const data = {
      clientId: selectedClient.id,
      creditLineId: selectedCreditLine.id,
      amount: loanAmount,
      numInstallments: loanInstallments,
      amortizationType: loanAmortType,
    };
    if (dialogMode === "withdrawal") {
      createWithdrawalMutation.mutate(data);
    } else {
      createLoanMutation.mutate(data);
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
      id: "clientId",
      label: t("admin.client"),
      minWidth: 150,
      render: (row) => row.clientName || row.clientId.slice(0, 8),
    },
    {
      id: "principal",
      label: t("common.amount"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.principal} fontWeight={500} />,
    },
    {
      id: "numInstallments",
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
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
        <Typography variant="h4">{t("admin.loanManagement")}</Typography>
        <Box display="flex" gap={1}>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => { setDialogMode("loan"); setCreateOpen(true); }}>
            {t("admin.createLoan")}
          </Button>
          <Button variant="contained" color="secondary" startIcon={<WithdrawalIcon />} onClick={() => { setDialogMode("withdrawal"); setCreateOpen(true); }}>
            {t("admin.cashWithdrawal")}
          </Button>
        </Box>
      </Box>
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

      {/* Create Loan Dialog */}
      <Dialog open={createOpen} onClose={handleCloseCreate} maxWidth="sm" fullWidth>
        <DialogTitle>{dialogMode === "withdrawal" ? t("admin.cashWithdrawal") : t("admin.createLoan")}</DialogTitle>
        <DialogContent>
          <Box mt={1} display="flex" flexDirection="column" gap={2}>
            {/* Step 1: Search & select client */}
            {!selectedClient ? (
              <>
                <TextField
                  fullWidth
                  size="small"
                  placeholder={t("admin.searchPlaceholder")}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSearch()}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <SearchIcon color="action" />
                      </InputAdornment>
                    ),
                  }}
                />
                <Button size="small" variant="outlined" onClick={handleSearch} disabled={!searchQuery.trim()}>
                  {t("common.search")}
                </Button>
                {searchResults.length > 0 && (
                  <List dense sx={{ border: 1, borderColor: "divider", borderRadius: 1, maxHeight: 200, overflow: "auto" }}>
                    {searchResults.map((c) => (
                      <ListItemButton key={c.id} onClick={() => handleSelectClient(c)}>
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
                {/* Selected client display */}
                <Box display="flex" justifyContent="space-between" alignItems="center">
                  <Typography variant="subtitle2">
                    {t("admin.client")}: {selectedClient.firstName} {selectedClient.lastName} (DNI: {selectedClient.dni})
                  </Typography>
                  <Button size="small" onClick={() => { setSelectedClient(null); setClientCreditLines([]); setSelectedCreditLine(null); }}>
                    {t("common.edit")}
                  </Button>
                </Box>

                {/* Step 2: Credit line */}
                {clientCreditLines.length === 0 ? (
                  <Alert severity="warning">{t("admin.noCreditLineForClient")}</Alert>
                ) : clientCreditLines.length === 1 ? (
                  <Box sx={{ p: 1.5, border: 1, borderColor: "divider", borderRadius: 1 }}>
                    <Typography variant="body2">
                      {t("creditLines.maxAmount")}: <MoneyDisplay amount={selectedCreditLine!.maxAmount} fontWeight={600} variant="body2" />
                      {" | "}{t("creditLines.availableAmount")}: <MoneyDisplay amount={selectedCreditLine!.availableAmount} variant="body2" color="success.main" />
                      {" | "}{t("loans.interestRate")}: {selectedCreditLine!.interestRate}%
                    </Typography>
                  </Box>
                ) : (
                  <TextField
                    select
                    fullWidth
                    label={t("admin.selectCreditLine")}
                    value={selectedCreditLine?.id || ""}
                    onChange={(e) => setSelectedCreditLine(clientCreditLines.find((cl) => cl.id === e.target.value) || null)}
                  >
                    {clientCreditLines.map((cl) => (
                      <MenuItem key={cl.id} value={cl.id}>
                        <MoneyDisplay amount={cl.availableAmount} variant="body2" /> {t("creditLines.availableAmount")} — {cl.interestRate}%
                      </MenuItem>
                    ))}
                  </TextField>
                )}

                {/* Step 3: Loan details */}
                {selectedCreditLine && (
                  <>
                    <TextField
                      fullWidth
                      type="number"
                      label={t("loans.loanAmount")}
                      value={loanAmount}
                      onChange={(e) => setLoanAmount(e.target.value)}
                      inputProps={{ min: 1, max: parseFloat(selectedCreditLine.availableAmount) }}
                    />
                    <TextField
                      fullWidth
                      type="number"
                      label={t("loans.numberOfInstallments")}
                      value={loanInstallments}
                      onChange={(e) => setLoanInstallments(Number(e.target.value))}
                      inputProps={{ min: 1, max: selectedCreditLine.maxInstallments }}
                    />
                    <TextField
                      select
                      fullWidth
                      label={t("loans.amortizationSystem")}
                      value={loanAmortType}
                      onChange={(e) => setLoanAmortType(e.target.value as "french" | "german")}
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
          <Button onClick={handleCloseCreate}>{t("common.cancel")}</Button>
          <Button
            variant="contained"
            onClick={handleCreateLoan}
            disabled={createLoanMutation.isPending || createWithdrawalMutation.isPending || !selectedClient || !selectedCreditLine || !loanAmount}
          >
            {(createLoanMutation.isPending || createWithdrawalMutation.isPending) ? t("common.creating") : t("common.create")}
          </Button>
        </DialogActions>
      </Dialog>

      {confirmAction && currentAction && (
        confirmAction.type === "cancel" && confirmAction.loan.cancellationSettlement ? (
          <Dialog open={true} onClose={() => setConfirmAction(null)} maxWidth="sm" fullWidth>
            <DialogTitle>{`${currentAction.label} ${t("loans.loanNumber")} #${confirmAction.loan.id.slice(0, 8)}`}</DialogTitle>
            <DialogContent>
              <Typography variant="body1" gutterBottom>{currentAction.message}</Typography>
              <Box sx={{ mt: 2, p: 2, bgcolor: "grey.50", borderRadius: 1 }}>
                <Typography variant="subtitle2" gutterBottom>{t("admin.settlementBreakdown")}</Typography>
                <Box display="flex" justifyContent="space-between" mb={0.5}>
                  <Typography variant="body2" color="text.secondary">{t("admin.pendingCapital")}</Typography>
                  <MoneyDisplay amount={confirmAction.loan.cancellationSettlement.pendingCapital} variant="body2" />
                </Box>
                <Box display="flex" justifyContent="space-between" mb={0.5}>
                  <Typography variant="body2" color="text.secondary">{t("admin.accumulatedInterest")}</Typography>
                  <MoneyDisplay amount={confirmAction.loan.cancellationSettlement.accumulatedInterest} variant="body2" />
                </Box>
                <Box display="flex" justifyContent="space-between" mb={0.5}>
                  <Typography variant="body2" color="text.secondary">{t("admin.accumulatedIVA")}</Typography>
                  <MoneyDisplay amount={confirmAction.loan.cancellationSettlement.accumulatedIVA} variant="body2" />
                </Box>
                <Box display="flex" justifyContent="space-between" mt={1} pt={1} sx={{ borderTop: 1, borderColor: "divider" }}>
                  <Typography variant="subtitle2">{t("admin.settlementTotal")}</Typography>
                  <MoneyDisplay amount={confirmAction.loan.cancellationSettlement.total} fontWeight={700} variant="subtitle2" />
                </Box>
              </Box>
            </DialogContent>
            <DialogActions sx={{ px: 3, pb: 2 }}>
              <Button onClick={() => setConfirmAction(null)} disabled={cancelMutation.isPending}>{t("common.cancel")}</Button>
              <Button variant="contained" color="error" onClick={handleConfirm} disabled={cancelMutation.isPending}>
                {cancelMutation.isPending ? t("common.processing") : currentAction.label}
              </Button>
            </DialogActions>
          </Dialog>
        ) : (
          <ConfirmDialog
            open={true}
            title={`${currentAction.label} ${t("loans.loanNumber")} #${confirmAction.loan.id.slice(0, 8)}`}
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
        )
      )}

    </Box>
  );
};

export default LoanManagement;
