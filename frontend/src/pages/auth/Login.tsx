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
  InputAdornment,
  IconButton,
} from "@mui/material";
import {
  Email as EmailIcon,
  Lock as LockIcon,
  Visibility,
  VisibilityOff,
  AdminPanelSettings as LogoIcon,
} from "@mui/icons-material";
import { useAuth } from "../../auth/AuthContext";
import { useTranslation } from "react-i18next";
import LanguageSwitcher from "../../components/LanguageSwitcher";

const schema = z.object({
  email: z.string().email("Email invalido"),
  password: z.string().min(6, "La contrasena debe tener al menos 6 caracteres"),
});

type FormData = z.infer<typeof schema>;

const Login: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const { t } = useTranslation();
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const { control, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", password: "" },
  });

  const onSubmit = async (data: FormData) => {
    setError("");
    setLoading(true);
    try {
      await login(data);
      navigate("/dashboard");
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      setError(axiosErr.response?.data?.message || t("auth.invalidCredentials"));
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
      <Card sx={{ maxWidth: 440, width: "100%", p: 1 }}>
        <CardContent>
          <Box textAlign="center" mb={3}>
            <Box sx={{ display: "flex", justifyContent: "flex-end" }}><LanguageSwitcher /></Box>
            <LogoIcon sx={{ fontSize: 48, color: "primary.main", mb: 1 }} />
            <Typography variant="h5" fontWeight={700} color="primary.main">
              {t("common.appName")}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {t("auth.login")}
            </Typography>
          </Box>

          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}

          <form onSubmit={handleSubmit(onSubmit)}>
            <Controller
              name="email"
              control={control}
              render={({ field }) => (
                <TextField
                  {...field}
                  fullWidth
                  label={t("auth.email")}
                  type="email"
                  margin="normal"
                  error={!!errors.email}
                  helperText={errors.email?.message}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <EmailIcon color="action" />
                      </InputAdornment>
                    ),
                  }}
                />
              )}
            />

            <Controller
              name="password"
              control={control}
              render={({ field }) => (
                <TextField
                  {...field}
                  fullWidth
                  label={t("auth.password")}
                  type={showPassword ? "text" : "password"}
                  margin="normal"
                  error={!!errors.password}
                  helperText={errors.password?.message}
                  InputProps={{
                    startAdornment: (
                      <InputAdornment position="start">
                        <LockIcon color="action" />
                      </InputAdornment>
                    ),
                    endAdornment: (
                      <InputAdornment position="end">
                        <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                          {showPassword ? <VisibilityOff /> : <Visibility />}
                        </IconButton>
                      </InputAdornment>
                    ),
                  }}
                />
              )}
            />

            <Button
              type="submit"
              fullWidth
              variant="contained"
              size="large"
              disabled={loading}
              sx={{ mt: 3, mb: 2, py: 1.5 }}
            >
              {loading ? t("common.loading") : t("auth.login")}
            </Button>
          </form>

          <Box textAlign="center">
            <Typography variant="body2" color="text.secondary">
              {t("auth.noAccount")}{" "}
              <Link component={RouterLink} to="/register" underline="hover">
                {t("auth.registerHere")}
              </Link>
            </Typography>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};

export default Login;
