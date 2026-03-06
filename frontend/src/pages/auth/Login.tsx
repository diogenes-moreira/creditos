import React, { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Link,
  InputAdornment,
  Tabs,
  Tab,
  Alert,
} from "@mui/material";
import {
  Email as EmailIcon,
  AdminPanelSettings as LogoIcon,
} from "@mui/icons-material";
import { useAuth } from "../../auth/AuthContext";
import { useTranslation } from "react-i18next";
import LanguageSwitcher from "../../components/LanguageSwitcher";
import { useNotification } from "../../contexts/NotificationContext";
import { getErrorMessage } from "../../api/errorUtils";

const Login: React.FC = () => {
  const navigate = useNavigate();
  const { login, requestOTP, verifyOTP } = useAuth();
  const { t } = useTranslation();
  const { showError } = useNotification();

  const [tab, setTab] = useState(0); // 0 = client (OTP), 1 = admin/vendor
  const [email, setEmail] = useState("");
  const [otpCode, setOtpCode] = useState("");
  const [otpSent, setOtpSent] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleAdminLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) return;
    setLoading(true);
    try {
      await login({ email });
      navigate("/dashboard");
    } catch (err: unknown) {
      showError(getErrorMessage(err, t("auth.invalidCredentials")));
    } finally {
      setLoading(false);
    }
  };

  const handleRequestOTP = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) return;
    setLoading(true);
    try {
      await requestOTP(email);
      setOtpSent(true);
    } catch (err: unknown) {
      showError(getErrorMessage(err, t("auth.invalidCredentials")));
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOTP = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !otpCode) return;
    setLoading(true);
    try {
      await verifyOTP(email, otpCode);
      navigate("/dashboard");
    } catch (err: unknown) {
      showError(getErrorMessage(err, t("auth.otpError")));
    } finally {
      setLoading(false);
    }
  };

  const resetOtpFlow = () => {
    setOtpSent(false);
    setOtpCode("");
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
          <Box textAlign="center" mb={2}>
            <Box sx={{ display: "flex", justifyContent: "flex-end" }}><LanguageSwitcher /></Box>
            <LogoIcon sx={{ fontSize: 48, color: "primary.main", mb: 1 }} />
            <Typography variant="h5" fontWeight={700} color="primary.main">
              {t("common.appName")}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {t("auth.login")}
            </Typography>
          </Box>

          <Tabs
            value={tab}
            onChange={(_, v) => { setTab(v); resetOtpFlow(); }}
            variant="fullWidth"
            sx={{ mb: 2 }}
          >
            <Tab label={t("auth.clientLogin")} />
            <Tab label={t("auth.adminLogin")} />
          </Tabs>

          {/* Client OTP Login */}
          {tab === 0 && !otpSent && (
            <form onSubmit={handleRequestOTP}>
              <TextField
                fullWidth
                label={t("auth.email")}
                type="email"
                margin="normal"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <EmailIcon color="action" />
                    </InputAdornment>
                  ),
                }}
              />
              <Button
                type="submit"
                fullWidth
                variant="contained"
                size="large"
                disabled={loading || !email}
                sx={{ mt: 2, mb: 2, py: 1.5 }}
              >
                {loading ? t("auth.sendingOtp") : t("auth.requestOtp")}
              </Button>
            </form>
          )}

          {/* OTP Verification Step */}
          {tab === 0 && otpSent && (
            <form onSubmit={handleVerifyOTP}>
              <Alert severity="info" sx={{ mb: 2 }}>
                {t("auth.otpSent")}
              </Alert>
              <TextField
                fullWidth
                label={t("auth.enterOtp")}
                margin="normal"
                value={otpCode}
                onChange={(e) => {
                  const val = e.target.value.replace(/\D/g, "").slice(0, 6);
                  setOtpCode(val);
                }}
                inputProps={{ maxLength: 6, inputMode: "numeric" }}
              />
              <Button
                type="submit"
                fullWidth
                variant="contained"
                size="large"
                disabled={loading || otpCode.length !== 6}
                sx={{ mt: 2, mb: 1, py: 1.5 }}
              >
                {loading ? t("auth.verifyingOtp") : t("auth.verifyOtp")}
              </Button>
              <Button
                fullWidth
                variant="text"
                size="small"
                onClick={resetOtpFlow}
              >
                {t("common.back")}
              </Button>
            </form>
          )}

          {/* Admin/Vendor Login */}
          {tab === 1 && (
            <form onSubmit={handleAdminLogin}>
              <TextField
                fullWidth
                label={t("auth.email")}
                type="email"
                margin="normal"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <EmailIcon color="action" />
                    </InputAdornment>
                  ),
                }}
              />
              <Button
                type="submit"
                fullWidth
                variant="contained"
                size="large"
                disabled={loading || !email}
                sx={{ mt: 2, mb: 2, py: 1.5 }}
              >
                {loading ? t("common.loading") : t("auth.login")}
              </Button>
            </form>
          )}

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
