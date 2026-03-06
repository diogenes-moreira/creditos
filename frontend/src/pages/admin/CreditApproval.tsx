import React, { useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  IconButton,
  Divider,
  Skeleton,
  InputAdornment,
} from "@mui/material";
import {
  Check as ApproveIcon,
  Close as RejectIcon,
  Add as AddIcon,
  Edit as EditIcon,
  Search as SearchIcon,
} from "@mui/icons-material";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  adminGetClients,
  adminGetClientCreditLines,
  adminCreateCreditLine,
  adminApproveCreditLine,
  adminRejectCreditLine,
  adminUpdateCreditLine,
} from "../../api/endpoints";
import { useNotification } from "../../contexts/NotificationContext";
import { getErrorMessage } from "../../api/errorUtils";
import MoneyDisplay from "../../components/MoneyDisplay";
import StatusBadge from "../../components/StatusBadge";
import ConfirmDialog from "../../components/ConfirmDialog";
import type { Client, CreditLine } from "../../api/types";

const createSchema = z.object({
  clientId: z.string().min(1, "ID de cliente requerido"),
  maxAmount: z.number().min(1000, "Monto minimo: $1.000"),
  interestRate: z.number().min(0.1, "Tasa requerida"),
  maxInstallments: z.number().min(1, "Minimo 1 cuota").max(60, "Maximo 60 cuotas"),
});

type CreateForm = z.infer<typeof createSchema>;

// Sub-component that fetches and displays credit lines for a single client
const ClientCreditCard: React.FC<{
  client: Client;
  onApprove: (clId: string) => void;
  onReject: (clId: string) => void;
  onEdit: (cl: CreditLine) => void;
  onAssign: (clientId: string) => void;
}> = ({ client, onApprove, onReject, onEdit, onAssign }) => {
  const { t } = useTranslation();

  const { data: creditLines, isLoading } = useQuery({
    queryKey: ["admin-client-credit-lines", client.id],
    queryFn: () => adminGetClientCreditLines(client.id),
  });

  return (
    <Card>
      <CardContent>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
          <Typography variant="subtitle1" fontWeight={600}>
            {client.firstName} {client.lastName}
          </Typography>
          <StatusBadge status={client.isBlocked ? "blocked" : "active"} />
        </Box>
        <Typography variant="body2" color="text.secondary" mb={2}>
          DNI: {client.dni} | {client.email}
        </Typography>

        {isLoading ? (
          <Skeleton variant="rectangular" height={60} sx={{ borderRadius: 1 }} />
        ) : creditLines && creditLines.length > 0 ? (
          creditLines.map((cl) => (
            <Box key={cl.id} mb={1}>
              <Divider sx={{ mb: 1 }} />
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={0.5}>
                <StatusBadge status={cl.status} />
                <Box display="flex" alignItems="center" gap={0.5}>
                  <Typography variant="body2" color="text.secondary">{t("creditLines.maxAmount")}:</Typography>
                  <MoneyDisplay amount={cl.maxAmount} fontWeight={600} variant="body2" />
                  <IconButton size="small" onClick={() => onEdit(cl)} title={t("admin.clientDetailEditLimit")}>
                    <EditIcon fontSize="small" />
                  </IconButton>
                </Box>
              </Box>
              <Grid container spacing={1}>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">{t("creditLines.usedAmount")}</Typography>
                  <Typography variant="body2"><MoneyDisplay amount={cl.usedAmount} variant="body2" color="error.main" /></Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">{t("creditLines.availableAmount")}</Typography>
                  <Typography variant="body2"><MoneyDisplay amount={cl.availableAmount} variant="body2" color="success.main" /></Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">{t("loans.interestRate")}</Typography>
                  <Typography variant="body2">{cl.interestRate}%</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" color="text.secondary">{t("creditLines.maxInstallments")}</Typography>
                  <Typography variant="body2">{cl.maxInstallments}</Typography>
                </Grid>
              </Grid>
              {cl.status === "pending" && (
                <Box display="flex" gap={1} mt={1}>
                  <Button size="small" variant="contained" color="success" startIcon={<ApproveIcon />}
                    onClick={() => onApprove(cl.id)}>
                    {t("creditLines.approve")}
                  </Button>
                  <Button size="small" variant="outlined" color="error" startIcon={<RejectIcon />}
                    onClick={() => onReject(cl.id)}>
                    {t("creditLines.reject")}
                  </Button>
                </Box>
              )}
            </Box>
          ))
        ) : (
          <Box textAlign="center" py={1}>
            <Typography variant="body2" color="text.secondary" sx={{ fontStyle: "italic", mb: 1 }}>
              {t("admin.noCreditLines")}
            </Typography>
            <Button size="small" variant="outlined" startIcon={<AddIcon />} onClick={() => onAssign(client.id)}>
              {t("creditLines.assignLine")}
            </Button>
          </Box>
        )}
      </CardContent>
    </Card>
  );
};

const CreditApproval: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editingCreditLine, setEditingCreditLine] = useState<CreditLine | null>(null);
  const [newMaxAmount, setNewMaxAmount] = useState("");
  const [confirmApproveId, setConfirmApproveId] = useState<string | null>(null);
  const [rejectDialogId, setRejectDialogId] = useState<string | null>(null);
  const [rejectReason, setRejectReason] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [createForClient, setCreateForClient] = useState<Client | null>(null);
  const { showSuccess, showError } = useNotification();

  const { data: clientsData } = useQuery({
    queryKey: ["admin-clients-for-credit"],
    queryFn: () => adminGetClients(1, 100),
  });

  const filteredClients = useMemo(() => {
    const clients = clientsData?.data || [];
    if (!searchQuery.trim()) return clients;
    const q = searchQuery.toLowerCase().trim();
    return clients.filter((c) =>
      `${c.firstName} ${c.lastName}`.toLowerCase().includes(q) ||
      c.dni.toLowerCase().includes(q) ||
      c.email.toLowerCase().includes(q)
    );
  }, [clientsData, searchQuery]);

  const { control, handleSubmit, reset, formState: { errors } } = useForm<CreateForm>({
    resolver: zodResolver(createSchema),
    defaultValues: { clientId: "", maxAmount: 100000, interestRate: 5, maxInstallments: 12 },
  });

  const invalidateCreditLines = () => {
    queryClient.invalidateQueries({ queryKey: ["admin-client-credit-lines"] });
  };

  const createMutation = useMutation({
    mutationFn: adminCreateCreditLine,
    onSuccess: () => {
      invalidateCreditLines();
      setCreateOpen(false);
      reset();
      showSuccess(t("creditLines.created"));
    },
    onError: (err: unknown) => showError(getErrorMessage(err, t("creditLines.createError"))),
  });

  const approveMutation = useMutation({
    mutationFn: adminApproveCreditLine,
    onSuccess: () => {
      invalidateCreditLines();
      setConfirmApproveId(null);
      showSuccess(t("creditLines.approved"));
    },
    onError: (err: unknown) => showError(getErrorMessage(err, t("creditLines.approveError"))),
  });

  const rejectMutation = useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => adminRejectCreditLine(id, reason),
    onSuccess: () => {
      invalidateCreditLines();
      setRejectDialogId(null);
      setRejectReason("");
      showSuccess(t("creditLines.rejected"));
    },
    onError: (err: unknown) => showError(getErrorMessage(err, t("creditLines.rejectError"))),
  });

  const updateCreditLineMutation = useMutation({
    mutationFn: ({ creditLineId, maxAmount }: { creditLineId: string; maxAmount: string }) =>
      adminUpdateCreditLine(creditLineId, { maxAmount }),
    onSuccess: () => {
      invalidateCreditLines();
      setEditDialogOpen(false);
      setEditingCreditLine(null);
      setNewMaxAmount("");
      showSuccess(t("admin.creditLineUpdated"));
    },
    onError: (err: unknown) => {
      showError(getErrorMessage(err, t("admin.creditLineUpdateError")));
    },
  });

  const handleApproveConfirm = () => {
    if (!confirmApproveId) return;
    approveMutation.mutate(confirmApproveId);
  };

  const handleRejectConfirm = () => {
    if (!rejectDialogId || !rejectReason.trim()) return;
    rejectMutation.mutate({ id: rejectDialogId, reason: rejectReason.trim() });
  };

  const handleEdit = (cl: CreditLine) => {
    setEditingCreditLine(cl);
    setNewMaxAmount(cl.maxAmount);
    setEditDialogOpen(true);
  };

  const handleAssign = (clientId: string) => {
    const client = (clientsData?.data || []).find((c) => c.id === clientId) || null;
    setCreateForClient(client);
    reset({ clientId, maxAmount: 100000, interestRate: 5, maxInstallments: 12 });
    setCreateOpen(true);
  };

  const handleOpenCreate = () => {
    setCreateForClient(null);
    reset({ clientId: "", maxAmount: 100000, interestRate: 5, maxInstallments: 12 });
    setCreateOpen(true);
  };

  const handleSaveCreditLine = () => {
    if (!editingCreditLine) return;
    updateCreditLineMutation.mutate({ creditLineId: editingCreditLine.id, maxAmount: newMaxAmount });
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">{t("creditLines.title")}</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleOpenCreate}>
          {t("creditLines.newLine")}
        </Button>
      </Box>

      <TextField
        fullWidth
        size="small"
        placeholder={t("admin.searchPlaceholder")}
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        sx={{ mb: 3 }}
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon color="action" />
            </InputAdornment>
          ),
        }}
      />

      <Grid container spacing={2}>
        {filteredClients.map((client) => (
          <Grid item xs={12} md={6} lg={4} key={client.id}>
            <ClientCreditCard
              client={client}
              onApprove={(clId) => setConfirmApproveId(clId)}
              onReject={(clId) => { setRejectDialogId(clId); setRejectReason(""); }}
              onEdit={handleEdit}
              onAssign={handleAssign}
            />
          </Grid>
        ))}
      </Grid>

      {/* Create Credit Line Dialog */}
      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("creditLines.create")}</DialogTitle>
        <DialogContent>
          <form id="create-credit-form" onSubmit={handleSubmit((data) => createMutation.mutate({
            clientId: data.clientId,
            maxAmount: String(data.maxAmount),
            interestRate: String(data.interestRate),
            maxInstallments: data.maxInstallments,
          }))}>
            <Box mt={1} display="flex" flexDirection="column" gap={2}>
              {createForClient ? (
                <TextField
                  fullWidth
                  label={t("admin.client")}
                  value={`${createForClient.firstName} ${createForClient.lastName} (DNI: ${createForClient.dni})`}
                  InputProps={{ readOnly: true }}
                  disabled
                />
              ) : (
                <Controller
                  name="clientId"
                  control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("creditLines.clientId")} error={!!errors.clientId} helperText={errors.clientId?.message} />
                  )}
                />
              )}
              <Controller
                name="maxAmount"
                control={control}
                render={({ field }) => (
                  <TextField {...field} onChange={(e) => field.onChange(Number(e.target.value))} fullWidth type="number" label={t("creditLines.maxAmount")} error={!!errors.maxAmount} helperText={errors.maxAmount?.message} />
                )}
              />
              <Controller
                name="interestRate"
                control={control}
                render={({ field }) => (
                  <TextField {...field} onChange={(e) => field.onChange(Number(e.target.value))} fullWidth type="number" label={t("loans.interestRate") + " (%)"} error={!!errors.interestRate} helperText={errors.interestRate?.message} />
                )}
              />
              <Controller
                name="maxInstallments"
                control={control}
                render={({ field }) => (
                  <TextField {...field} onChange={(e) => field.onChange(Number(e.target.value))} fullWidth type="number" label={t("loans.installments")} error={!!errors.maxInstallments} helperText={errors.maxInstallments?.message} />
                )}
              />
            </Box>
          </form>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setCreateOpen(false)}>{t("common.cancel")}</Button>
          <Button type="submit" form="create-credit-form" variant="contained" disabled={createMutation.isPending}>
            {createMutation.isPending ? t("common.creating") : t("common.create")}
          </Button>
        </DialogActions>
      </Dialog>

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

      {/* Approve Confirm Dialog */}
      <ConfirmDialog
        open={!!confirmApproveId}
        title={t("creditLines.approveLine")}
        message={t("creditLines.approveConfirm")}
        confirmLabel={t("creditLines.approve")}
        confirmColor="success"
        onConfirm={handleApproveConfirm}
        onCancel={() => setConfirmApproveId(null)}
        loading={approveMutation.isPending}
      />

      {/* Reject Dialog with reason */}
      <Dialog open={!!rejectDialogId} onClose={() => setRejectDialogId(null)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("creditLines.rejectLine")}</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" mb={2}>
            {t("creditLines.rejectConfirm")}
          </Typography>
          <TextField
            fullWidth
            multiline
            rows={3}
            label={t("creditLines.rejectionReason")}
            value={rejectReason}
            onChange={(e) => setRejectReason(e.target.value)}
          />
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={() => setRejectDialogId(null)} disabled={rejectMutation.isPending}>
            {t("common.cancel")}
          </Button>
          <Button
            onClick={handleRejectConfirm}
            variant="contained"
            color="error"
            disabled={rejectMutation.isPending || !rejectReason.trim()}
          >
            {rejectMutation.isPending ? t("common.processing") : t("creditLines.reject")}
          </Button>
        </DialogActions>
      </Dialog>

    </Box>
  );
};

export default CreditApproval;
