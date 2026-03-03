import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Button,
  Snackbar,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
} from "@mui/material";
import {
  Check as ApproveIcon,
  Close as RejectIcon,
  Add as AddIcon,
} from "@mui/icons-material";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  adminGetClients,
  adminCreateCreditLine,
  adminApproveCreditLine,
  adminRejectCreditLine,
} from "../../api/endpoints";
import MoneyDisplay from "../../components/MoneyDisplay";
import StatusBadge from "../../components/StatusBadge";
import ConfirmDialog from "../../components/ConfirmDialog";

const createSchema = z.object({
  clientId: z.string().min(1, "ID de cliente requerido"),
  maxAmount: z.number().min(1000, "Monto minimo: $1.000"),
  interestRate: z.number().min(0.1, "Tasa requerida"),
});

type CreateForm = z.infer<typeof createSchema>;

const CreditApproval: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);
  const [confirmAction, setConfirmAction] = useState<{ type: "approve" | "reject"; id: string } | null>(null);
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" as "success" | "error" });

  const { data: clientsData } = useQuery({
    queryKey: ["admin-clients-for-credit"],
    queryFn: () => adminGetClients(1, 100),
  });

  const { control, handleSubmit, reset, formState: { errors } } = useForm<CreateForm>({
    resolver: zodResolver(createSchema),
    defaultValues: { clientId: "", maxAmount: 100000, interestRate: 5 },
  });

  const createMutation = useMutation({
    mutationFn: adminCreateCreditLine,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-clients"] });
      setCreateOpen(false);
      reset();
      setSnackbar({ open: true, message: t("creditLines.created"), severity: "success" });
    },
    onError: () => setSnackbar({ open: true, message: t("creditLines.createError"), severity: "error" }),
  });

  const approveMutation = useMutation({
    mutationFn: adminApproveCreditLine,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-clients"] });
      setConfirmAction(null);
      setSnackbar({ open: true, message: t("creditLines.approved"), severity: "success" });
    },
    onError: () => setSnackbar({ open: true, message: t("creditLines.approveError"), severity: "error" }),
  });

  const rejectMutation = useMutation({
    mutationFn: adminRejectCreditLine,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-clients"] });
      setConfirmAction(null);
      setSnackbar({ open: true, message: t("creditLines.rejected"), severity: "success" });
    },
    onError: () => setSnackbar({ open: true, message: t("creditLines.rejectError"), severity: "error" }),
  });

  const handleConfirm = () => {
    if (!confirmAction) return;
    if (confirmAction.type === "approve") approveMutation.mutate(confirmAction.id);
    else rejectMutation.mutate(confirmAction.id);
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">{t("creditLines.title")}</Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={() => setCreateOpen(true)}>
          {t("creditLines.newLine")}
        </Button>
      </Box>

      <Typography variant="body1" color="text.secondary" mb={3}>
        {t("creditLines.managementDesc")}
      </Typography>

      <Grid container spacing={2}>
        {(clientsData?.data || []).slice(0, 12).map((client) => (
          <Grid item xs={12} md={6} lg={4} key={client.id}>
            <Card>
              <CardContent>
                <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                  <Typography variant="subtitle1" fontWeight={600}>
                    {client.firstName} {client.lastName}
                  </Typography>
                  <StatusBadge status={client.status} />
                </Box>
                <Typography variant="body2" color="text.secondary" mb={2}>
                  DNI: {client.dni} | {client.email}
                </Typography>
                <Box display="flex" gap={1}>
                  <Button
                    size="small"
                    variant="contained"
                    color="success"
                    startIcon={<ApproveIcon />}
                    onClick={() => setConfirmAction({ type: "approve", id: client.id })}
                  >
                    {t("creditLines.approve")}
                  </Button>
                  <Button
                    size="small"
                    variant="outlined"
                    color="error"
                    startIcon={<RejectIcon />}
                    onClick={() => setConfirmAction({ type: "reject", id: client.id })}
                  >
                    {t("creditLines.reject")}
                  </Button>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{t("creditLines.create")}</DialogTitle>
        <DialogContent>
          <form id="create-credit-form" onSubmit={handleSubmit((data) => createMutation.mutate(data))}>
            <Box mt={1} display="flex" flexDirection="column" gap={2}>
              <Controller
                name="clientId"
                control={control}
                render={({ field }) => (
                  <TextField {...field} fullWidth label={t("creditLines.clientId")} error={!!errors.clientId} helperText={errors.clientId?.message} />
                )}
              />
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

      <ConfirmDialog
        open={!!confirmAction}
        title={confirmAction?.type === "approve" ? t("creditLines.approveLine") : t("creditLines.rejectLine")}
        message={confirmAction?.type === "approve"
          ? t("creditLines.approveConfirm")
          : t("creditLines.rejectConfirm")}
        confirmLabel={confirmAction?.type === "approve" ? t("creditLines.approve") : t("creditLines.reject")}
        confirmColor={confirmAction?.type === "approve" ? "success" : "error"}
        onConfirm={handleConfirm}
        onCancel={() => setConfirmAction(null)}
        loading={approveMutation.isPending || rejectMutation.isPending}
      />

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

export default CreditApproval;
