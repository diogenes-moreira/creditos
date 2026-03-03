import React, { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
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
  Snackbar,
  Alert,
  CircularProgress,
} from "@mui/material";
import { getProfile, updateProfile, updateMercadoPago } from "../../api/endpoints";

const profileSchema = z.object({
  firstName: z.string().min(1, "Nombre requerido"),
  lastName: z.string().min(1, "Apellido requerido"),
  phone: z.string().min(1, "Telefono requerido"),
  address: z.string().min(1, "Direccion requerida"),
  city: z.string().min(1, "Ciudad requerida"),
  province: z.string().min(1, "Provincia requerida"),
});

const mpSchema = z.object({
  alias: z.string().min(1, "Alias requerido"),
  cvu: z.string().min(22, "CVU invalido").max(22, "CVU invalido"),
});

type ProfileForm = z.infer<typeof profileSchema>;
type MPForm = z.infer<typeof mpSchema>;

const Profile: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [snackbar, setSnackbar] = useState({ open: false, message: "", severity: "success" as "success" | "error" });

  const { data: profile, isLoading } = useQuery({
    queryKey: ["profile"],
    queryFn: getProfile,
  });

  const {
    control: profileControl,
    handleSubmit: handleProfileSubmit,
    formState: { errors: profileErrors },
  } = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    values: profile ? {
      firstName: profile.firstName,
      lastName: profile.lastName,
      phone: profile.phone,
      address: profile.address,
      city: profile.city,
      province: profile.province,
    } : undefined,
  });

  const {
    control: mpControl,
    handleSubmit: handleMPSubmit,
    formState: { errors: mpErrors },
  } = useForm<MPForm>({
    resolver: zodResolver(mpSchema),
    values: profile ? {
      alias: profile.mercadoPagoAlias || "",
      cvu: profile.mercadoPagoCvu || "",
    } : undefined,
  });

  const profileMutation = useMutation({
    mutationFn: updateProfile,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] });
      setSnackbar({ open: true, message: t("profile.updateSuccess"), severity: "success" });
    },
    onError: () => setSnackbar({ open: true, message: t("profile.updateError"), severity: "error" }),
  });

  const mpMutation = useMutation({
    mutationFn: updateMercadoPago,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] });
      setSnackbar({ open: true, message: t("profile.mpUpdated"), severity: "success" });
    },
    onError: () => setSnackbar({ open: true, message: t("profile.mpError"), severity: "error" }),
  });

  if (isLoading) {
    return <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>;
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>{t("profile.title")}</Typography>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>{t("profile.personalData")}</Typography>
          <Divider sx={{ mb: 2 }} />
          <form onSubmit={handleProfileSubmit((data) => profileMutation.mutate(data))}>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="firstName"
                  control={profileControl}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.firstName")} error={!!profileErrors.firstName} helperText={profileErrors.firstName?.message} />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="lastName"
                  control={profileControl}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.lastName")} error={!!profileErrors.lastName} helperText={profileErrors.lastName?.message} />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="phone"
                  control={profileControl}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.phone")} error={!!profileErrors.phone} helperText={profileErrors.phone?.message} />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="address"
                  control={profileControl}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.address")} error={!!profileErrors.address} helperText={profileErrors.address?.message} />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="city"
                  control={profileControl}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.city")} error={!!profileErrors.city} helperText={profileErrors.city?.message} />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="province"
                  control={profileControl}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.province")} error={!!profileErrors.province} helperText={profileErrors.province?.message} />
                  )}
                />
              </Grid>
              <Grid item xs={12}>
                <Typography variant="body2" color="text.secondary" mb={1}>
                  Email: {profile?.email} | DNI: {profile?.dni}
                </Typography>
                <Button type="submit" variant="contained" disabled={profileMutation.isPending}>
                  {profileMutation.isPending ? t("common.saving") : t("profile.saveChanges")}
                </Button>
              </Grid>
            </Grid>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>{t("profile.mercadoPago")}</Typography>
          <Divider sx={{ mb: 2 }} />
          <form onSubmit={handleMPSubmit((data) => mpMutation.mutate(data))}>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="alias"
                  control={mpControl}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("profile.alias")} error={!!mpErrors.alias} helperText={mpErrors.alias?.message} />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller
                  name="cvu"
                  control={mpControl}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("profile.cvu")} error={!!mpErrors.cvu} helperText={mpErrors.cvu?.message} />
                  )}
                />
              </Grid>
              <Grid item xs={12}>
                <Button type="submit" variant="contained" disabled={mpMutation.isPending}>
                  {mpMutation.isPending ? t("common.saving") : t("profile.saveMercadoPago")}
                </Button>
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
        <Alert severity={snackbar.severity} onClose={() => setSnackbar({ ...snackbar, open: false })}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default Profile;
