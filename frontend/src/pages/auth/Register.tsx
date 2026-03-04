import React, { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Link,
  Alert,
  Grid,
  InputAdornment,
  IconButton,
} from "@mui/material";
import {
  Visibility,
  VisibilityOff,
  AdminPanelSettings as LogoIcon,
} from "@mui/icons-material";
import { useTranslation } from "react-i18next";
import { useAuth } from "../../auth/AuthContext";

const schema = z.object({
  email: z.string().email("Email invalido"),
  password: z.string().min(6, "Minimo 6 caracteres"),
  confirmPassword: z.string(),
  firstName: z.string().min(1, "Nombre requerido"),
  lastName: z.string().min(1, "Apellido requerido"),
  dni: z.string().min(7, "DNI invalido").max(8, "DNI invalido"),
  cuit: z.string().min(11, "CUIT invalido").max(11, "CUIT invalido"),
  dateOfBirth: z.string().min(1, "Fecha de nacimiento requerida"),
  phone: z.string().min(8, "Telefono invalido"),
  address: z.string().min(1, "Direccion requerida"),
  city: z.string().min(1, "Ciudad requerida"),
  province: z.string().min(1, "Provincia requerida"),
  isPEP: z.boolean(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Las contrasenas no coinciden",
  path: ["confirmPassword"],
});

type FormData = z.infer<typeof schema>;

const Register: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { register: authRegister } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const { control, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: "", password: "", confirmPassword: "",
      firstName: "", lastName: "", dni: "", cuit: "",
      dateOfBirth: "", phone: "", address: "", city: "", province: "",
      isPEP: false,
    },
  });

  const onSubmit = async (data: FormData) => {
    setError("");
    setLoading(true);
    try {
      const { confirmPassword, ...registerData } = data;
      await authRegister(registerData);
      navigate("/dashboard");
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      setError(axiosErr.response?.data?.message || t("auth.registerError"));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "linear-gradient(135deg, #1565C0 0%, #0D47A1 50%, #1B5E20 100%)",
        p: 2,
      }}
    >
      <Card sx={{ maxWidth: 640, width: "100%", p: 1 }}>
        <CardContent>
          <Box textAlign="center" mb={3}>
            <LogoIcon sx={{ fontSize: 48, color: "primary.main", mb: 1 }} />
            <Typography variant="h5" fontWeight={700} color="primary.main">
              {t("common.appName")}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {t("registration.title")}
            </Typography>
          </Box>

          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

          <form onSubmit={handleSubmit(onSubmit)}>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Controller name="firstName" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.firstName")} error={!!errors.firstName} helperText={errors.firstName?.message} />
                  )} />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller name="lastName" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.lastName")} error={!!errors.lastName} helperText={errors.lastName?.message} />
                  )} />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller name="dni" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.dni")} error={!!errors.dni} helperText={errors.dni?.message} />
                  )} />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller name="phone" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.phone")} error={!!errors.phone} helperText={errors.phone?.message} />
                  )} />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller name="cuit" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label="CUIT" error={!!errors.cuit} helperText={errors.cuit?.message} />
                  )} />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller name="dateOfBirth" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth type="date" label={t("registration.dateOfBirth")} InputLabelProps={{ shrink: true }} error={!!errors.dateOfBirth} helperText={errors.dateOfBirth?.message} />
                  )} />
              </Grid>
              <Grid item xs={12}>
                <Controller name="email" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("auth.email")} type="email" error={!!errors.email} helperText={errors.email?.message} />
                  )} />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller name="password" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("auth.password")}
                      type={showPassword ? "text" : "password"}
                      error={!!errors.password} helperText={errors.password?.message}
                      InputProps={{
                        endAdornment: (
                          <InputAdornment position="end">
                            <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                              {showPassword ? <VisibilityOff /> : <Visibility />}
                            </IconButton>
                          </InputAdornment>
                        ),
                      }}
                    />
                  )} />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller name="confirmPassword" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.confirmPassword")}
                      type={showPassword ? "text" : "password"}
                      error={!!errors.confirmPassword} helperText={errors.confirmPassword?.message} />
                  )} />
              </Grid>
              <Grid item xs={12}>
                <Controller name="address" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.address")} error={!!errors.address} helperText={errors.address?.message} />
                  )} />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller name="city" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.city")} error={!!errors.city} helperText={errors.city?.message} />
                  )} />
              </Grid>
              <Grid item xs={12} sm={6}>
                <Controller name="province" control={control}
                  render={({ field }) => (
                    <TextField {...field} fullWidth label={t("registration.province")} error={!!errors.province} helperText={errors.province?.message} />
                  )} />
              </Grid>
            </Grid>

            <Button
              type="submit"
              fullWidth
              variant="contained"
              size="large"
              disabled={loading}
              sx={{ mt: 3, mb: 2, py: 1.5 }}
            >
              {loading ? t("auth.registering") : t("registration.submit")}
            </Button>
          </form>

          <Box textAlign="center">
            <Typography variant="body2" color="text.secondary">
              {t("auth.hasAccount")}{" "}
              <Link component={RouterLink} to="/login" underline="hover">
                {t("auth.loginHere")}
              </Link>
            </Typography>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};

export default Register;
