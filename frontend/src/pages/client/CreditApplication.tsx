import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useMutation } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  TextField,
  MenuItem,
  Button,
  Stepper,
  Step,
  StepLabel,
  Divider,
  Alert,
  Snackbar,
} from "@mui/material";
import { simulateLoan, requestLoan } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import type { SimulationResult, Installment } from "../../api/types";

const schema = z.object({
  creditLineId: z.string().min(1, "Selecciona una linea de credito"),
  amount: z.number().min(1000, "El monto minimo es $1.000").max(5000000, "El monto maximo es $5.000.000"),
  installments: z.number().min(1, "Minimo 1 cuota").max(60, "Maximo 60 cuotas"),
  interestRate: z.number().min(0.1, "Tasa requerida"),
  amortizationType: z.enum(["french", "german"]),
});

type FormData = z.infer<typeof schema>;

const CreditApplication: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const steps = [t("loans.configureCredit"), t("loans.simulate"), t("loans.confirmStep")];
  const [activeStep, setActiveStep] = useState(0);
  const [simulation, setSimulation] = useState<SimulationResult | null>(null);
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" as "success" | "error" });

  const { control, handleSubmit, watch, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      creditLineId: "",
      amount: 50000,
      installments: 12,
      interestRate: 5,
      amortizationType: "french",
    },
  });

  const formValues = watch();

  const simulateMutation = useMutation({
    mutationFn: simulateLoan,
    onSuccess: (data) => {
      setSimulation(data);
      setActiveStep(1);
    },
    onError: () => {
      setSnackbar({ open: true, message: t("loans.simulationError"), severity: "error" });
    },
  });

  const requestMutation = useMutation({
    mutationFn: requestLoan,
    onSuccess: () => {
      setSnackbar({ open: true, message: t("loans.requestSuccess"), severity: "success" });
      setTimeout(() => navigate("/loans"), 1500);
    },
    onError: () => {
      setSnackbar({ open: true, message: t("loans.requestError"), severity: "error" });
    },
  });

  const handleSimulate = (data: FormData) => {
    simulateMutation.mutate({
      amount: data.amount,
      installments: data.installments,
      interestRate: data.interestRate,
      amortizationType: data.amortizationType,
    });
  };

  const handleConfirm = () => {
    requestMutation.mutate({
      creditLineId: formValues.creditLineId,
      amount: formValues.amount,
      installments: formValues.installments,
      amortizationType: formValues.amortizationType,
    });
  };

  const installmentColumns: Column<Installment>[] = [
    { id: "number", label: "#", align: "center" },
    {
      id: "dueDate",
      label: t("loans.dueDate"),
      render: (row) => row.dueDate,
    },
    {
      id: "capitalAmount",
      label: t("loans.capital"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.capitalAmount} />,
    },
    {
      id: "interestAmount",
      label: t("loans.interest"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.interestAmount} />,
    },
    {
      id: "totalAmount",
      label: t("loans.installmentTotal"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.totalAmount} fontWeight={600} />,
    },
    {
      id: "remainingAmount",
      label: t("account.balance"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.remainingAmount} />,
    },
  ];

  return (
    <Box>
      <Typography variant="h4" gutterBottom>{t("loans.requestCredit")}</Typography>

      <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
        {steps.map((label) => (
          <Step key={label}>
            <StepLabel>{label}</StepLabel>
          </Step>
        ))}
      </Stepper>

      {activeStep === 0 && (
        <Card>
          <CardContent>
            <form onSubmit={handleSubmit(handleSimulate)}>
              <Grid container spacing={3}>
                <Grid item xs={12} sm={6}>
                  <Controller
                    name="creditLineId"
                    control={control}
                    render={({ field }) => (
                      <TextField
                        {...field}
                        fullWidth
                        label={t("loans.creditLine")}
                        placeholder={t("loans.creditLineId")}
                        error={!!errors.creditLineId}
                        helperText={errors.creditLineId?.message}
                      />
                    )}
                  />
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Controller
                    name="amount"
                    control={control}
                    render={({ field }) => (
                      <TextField
                        {...field}
                        onChange={(e) => field.onChange(Number(e.target.value))}
                        fullWidth
                        type="number"
                        label={t("loans.loanAmount")}
                        error={!!errors.amount}
                        helperText={errors.amount?.message}
                      />
                    )}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <Controller
                    name="installments"
                    control={control}
                    render={({ field }) => (
                      <TextField
                        {...field}
                        onChange={(e) => field.onChange(Number(e.target.value))}
                        fullWidth
                        type="number"
                        label={t("loans.installments")}
                        error={!!errors.installments}
                        helperText={errors.installments?.message}
                      />
                    )}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <Controller
                    name="interestRate"
                    control={control}
                    render={({ field }) => (
                      <TextField
                        {...field}
                        onChange={(e) => field.onChange(Number(e.target.value))}
                        fullWidth
                        type="number"
                        label={t("loans.interestRate") + " (%)"}
                        error={!!errors.interestRate}
                        helperText={errors.interestRate?.message}
                      />
                    )}
                  />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <Controller
                    name="amortizationType"
                    control={control}
                    render={({ field }) => (
                      <TextField
                        {...field}
                        select
                        fullWidth
                        label={t("loans.amortizationSystem")}
                      >
                        <MenuItem value="french">{t("loans.french")}</MenuItem>
                        <MenuItem value="german">{t("loans.german")}</MenuItem>
                      </TextField>
                    )}
                  />
                </Grid>
                <Grid item xs={12}>
                  <Button
                    type="submit"
                    variant="contained"
                    size="large"
                    disabled={simulateMutation.isPending}
                    sx={{ mr: 2 }}
                  >
                    {simulateMutation.isPending ? t("loans.simulating") : t("loans.simulate")}
                  </Button>
                </Grid>
              </Grid>
            </form>
          </CardContent>
        </Card>
      )}

      {activeStep === 1 && simulation && (
        <Box>
          <Card sx={{ mb: 3 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>{t("loans.simulationResult")}</Typography>
              <Grid container spacing={3}>
                <Grid item xs={12} sm={4}>
                  <Typography variant="body2" color="text.secondary">{t("loans.totalPayment")}</Typography>
                  <MoneyDisplay amount={simulation.totalPayment} variant="h5" fontWeight={700} color="primary.main" />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <Typography variant="body2" color="text.secondary">{t("loans.totalInterest")}</Typography>
                  <MoneyDisplay amount={simulation.totalInterest} variant="h5" fontWeight={700} color="warning.main" />
                </Grid>
                <Grid item xs={12} sm={4}>
                  <Typography variant="body2" color="text.secondary">{t("loans.principal")}</Typography>
                  <MoneyDisplay amount={simulation.principal} variant="h5" fontWeight={700} color="secondary.main" />
                </Grid>
              </Grid>
            </CardContent>
          </Card>

          <Typography variant="h6" gutterBottom>{t("loans.installmentSchedule")}</Typography>
          <DataTable
            columns={installmentColumns}
            rows={simulation.installments || []}
            keyExtractor={(row) => String(row.number)}
          />

          <Box display="flex" gap={2} mt={3}>
            <Button variant="outlined" onClick={() => setActiveStep(0)}>
              {t("loans.editBack")}
            </Button>
            <Button variant="contained" onClick={() => setActiveStep(2)}>
              {t("loans.continueRequest")}
            </Button>
          </Box>
        </Box>
      )}

      {activeStep === 2 && (
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>{t("loans.confirmRequest")}</Typography>
            <Divider sx={{ mb: 2 }} />
            <Grid container spacing={2} mb={3}>
              <Grid item xs={6}>
                <Typography variant="body2" color="text.secondary">{t("common.amount")}</Typography>
                <MoneyDisplay amount={formValues.amount} variant="h6" fontWeight={600} />
              </Grid>
              <Grid item xs={6}>
                <Typography variant="body2" color="text.secondary">{t("loans.installments")}</Typography>
                <Typography variant="h6" fontWeight={600}>{formValues.installments}</Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="body2" color="text.secondary">{t("loans.system")}</Typography>
                <Typography>{formValues.amortizationType === "french" ? t("loans.french") : t("loans.german")}</Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="body2" color="text.secondary">{t("loans.totalPayment")}</Typography>
                <MoneyDisplay amount={simulation?.totalPayment || "0"} variant="h6" fontWeight={600} color="primary.main" />
              </Grid>
            </Grid>

            <Alert severity="info" sx={{ mb: 3 }}>
              {t("loans.requestReview")}
            </Alert>

            <Box display="flex" gap={2}>
              <Button variant="outlined" onClick={() => setActiveStep(1)}>
                {t("common.back")}
              </Button>
              <Button
                variant="contained"
                color="success"
                size="large"
                onClick={handleConfirm}
                disabled={requestMutation.isPending}
              >
                {requestMutation.isPending ? t("common.sending") : t("loans.sendRequest")}
              </Button>
            </Box>
          </CardContent>
        </Card>
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

export default CreditApplication;
