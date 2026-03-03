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
  Button,
  Divider,
  Alert,
  Snackbar,
  FormControlLabel,
  Checkbox,
  InputAdornment,
  IconButton,
} from "@mui/material";
import { Visibility, VisibilityOff } from "@mui/icons-material";
import { vendorRegisterClient } from "../../api/endpoints";
import type { Client } from "../../api/types";

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(6),
  firstName: z.string().min(1),
  lastName: z.string().min(1),
  dni: z.string().min(7).max(8),
  cuit: z.string().min(11).max(13),
  dateOfBirth: z.string().min(1),
  phone: z.string().min(8),
  address: z.string().min(1),
  city: z.string().min(1),
  province: z.string().min(1),
  isPEP: z.boolean(),
});

type FormData = z.infer<typeof schema>;

const ClientRegister: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [showPassword, setShowPassword] = useState(false);
  const [registeredClient, setRegisteredClient] = useState<Client | null>(null);
  const [snackbar, setSnackbar] = useState({
    open: false,
    message: "",
    severity: "success" as "success" | "error",
  });

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: "",
      password: "",
      firstName: "",
      lastName: "",
      dni: "",
      cuit: "",
      dateOfBirth: "",
      phone: "",
      address: "",
      city: "",
      province: "",
      isPEP: false,
    },
  });

  const mutation = useMutation({
    mutationFn: vendorRegisterClient,
    onSuccess: (client) => {
      setRegisteredClient(client);
      setSnackbar({
        open: true,
        message: t("vendor.clientRegisterSuccess"),
        severity: "success",
      });
    },
    onError: () => {
      setSnackbar({
        open: true,
        message: t("vendor.clientRegisterError"),
        severity: "error",
      });
    },
  });

  const onSubmit = (data: FormData) => {
    mutation.mutate(data);
  };

  if (registeredClient) {
    return (
      <Box>
        <Typography variant="h4" fontWeight={700} gutterBottom>
          {t("vendor.registerClient")}
        </Typography>

        <Alert severity="success" sx={{ mb: 3 }}>
          {t("vendor.clientRegisteredSuccessfully", {
            name: `${registeredClient.firstName} ${registeredClient.lastName}`,
          })}
        </Alert>

        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              {t("vendor.registeredClientDetails")}
            </Typography>
            <Divider sx={{ mb: 2 }} />
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Typography variant="body2" color="text.secondary">
                  {t("vendor.fullName")}
                </Typography>
                <Typography fontWeight={600}>
                  {registeredClient.firstName} {registeredClient.lastName}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="body2" color="text.secondary">
                  DNI
                </Typography>
                <Typography fontWeight={600}>
                  {registeredClient.dni}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="body2" color="text.secondary">
                  Email
                </Typography>
                <Typography fontWeight={600}>
                  {registeredClient.email}
                </Typography>
              </Grid>
            </Grid>

            <Box display="flex" gap={2} mt={3}>
              <Button
                variant="contained"
                onClick={() => navigate("/vendor/purchases/new")}
              >
                {t("vendor.newPurchase")}
              </Button>
              <Button
                variant="outlined"
                onClick={() => {
                  setRegisteredClient(null);
                }}
              >
                {t("vendor.registerAnother")}
              </Button>
            </Box>
          </CardContent>
        </Card>
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" fontWeight={700} gutterBottom>
        {t("vendor.registerClient")}
      </Typography>
      <Typography variant="body1" color="text.secondary" mb={3}>
        {t("vendor.registerClientDescription")}
      </Typography>

      <Card>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)}>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="firstName"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("registration.firstName")}
                      error={!!errors.firstName}
                      helperText={
                        errors.firstName
                          ? t("vendor.fieldRequired")
                          : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="lastName"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("registration.lastName")}
                      error={!!errors.lastName}
                      helperText={
                        errors.lastName
                          ? t("vendor.fieldRequired")
                          : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="dni"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("registration.dni")}
                      error={!!errors.dni}
                      helperText={
                        errors.dni ? t("vendor.dniInvalid") : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="cuit"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("vendor.cuit")}
                      error={!!errors.cuit}
                      helperText={
                        errors.cuit ? t("vendor.cuitInvalid") : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="dateOfBirth"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      type="date"
                      label={t("vendor.dateOfBirth")}
                      InputLabelProps={{ shrink: true }}
                      error={!!errors.dateOfBirth}
                      helperText={
                        errors.dateOfBirth
                          ? t("vendor.fieldRequired")
                          : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="phone"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("registration.phone")}
                      error={!!errors.phone}
                      helperText={
                        errors.phone ? t("vendor.phoneInvalid") : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12}>
                <Controller
                  name="email"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("auth.email")}
                      type="email"
                      error={!!errors.email}
                      helperText={
                        errors.email ? t("vendor.emailInvalid") : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="password"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("auth.password")}
                      type={showPassword ? "text" : "password"}
                      error={!!errors.password}
                      helperText={
                        errors.password
                          ? t("vendor.passwordMinLength")
                          : undefined
                      }
                      InputProps={{
                        endAdornment: (
                          <InputAdornment position="end">
                            <IconButton
                              onClick={() => setShowPassword(!showPassword)}
                              edge="end"
                            >
                              {showPassword ? (
                                <VisibilityOff />
                              ) : (
                                <Visibility />
                              )}
                            </IconButton>
                          </InputAdornment>
                        ),
                      }}
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12}>
                <Controller
                  name="address"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("registration.address")}
                      error={!!errors.address}
                      helperText={
                        errors.address ? t("vendor.fieldRequired") : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="city"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("registration.city")}
                      error={!!errors.city}
                      helperText={
                        errors.city ? t("vendor.fieldRequired") : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="province"
                  control={control}
                  render={({ field }) => (
                    <TextField
                      {...field}
                      fullWidth
                      label={t("registration.province")}
                      error={!!errors.province}
                      helperText={
                        errors.province
                          ? t("vendor.fieldRequired")
                          : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12}>
                <Controller
                  name="isPEP"
                  control={control}
                  render={({ field }) => (
                    <FormControlLabel
                      control={
                        <Checkbox
                          checked={field.value}
                          onChange={(e) => field.onChange(e.target.checked)}
                        />
                      }
                      label={t("vendor.isPEP")}
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12}>
                <Divider sx={{ mb: 2 }} />
                <Box display="flex" gap={2}>
                  <Button
                    type="submit"
                    variant="contained"
                    size="large"
                    disabled={mutation.isPending}
                  >
                    {mutation.isPending
                      ? t("common.saving")
                      : t("vendor.registerClientButton")}
                  </Button>
                  <Button
                    variant="outlined"
                    onClick={() => navigate("/vendor/purchases/new")}
                  >
                    {t("common.back")}
                  </Button>
                </Box>
              </Grid>
            </Grid>
          </form>
        </CardContent>
      </Card>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={4000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: "bottom", horizontal: "center" }}
      >
        <Alert
          severity={snackbar.severity}
          onClose={() => setSnackbar({ ...snackbar, open: false })}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default ClientRegister;
