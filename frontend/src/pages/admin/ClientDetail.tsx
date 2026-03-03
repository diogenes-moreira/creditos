import React from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import {
  Box,
  Typography,
  Card,
  CardContent,
  Grid,
  Button,
  Divider,
  CircularProgress,
  Alert,
} from "@mui/material";
import { ArrowBack as BackIcon } from "@mui/icons-material";
import { format } from "date-fns";
import { adminGetClient } from "../../api/endpoints";
import StatusBadge from "../../components/StatusBadge";

const ClientDetail: React.FC = () => {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: client, isLoading } = useQuery({
    queryKey: ["admin-client", id],
    queryFn: () => adminGetClient(id!),
    enabled: !!id,
  });

  if (isLoading) {
    return <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>;
  }

  if (!client) {
    return <Alert severity="error">{t("admin.clientNotFound")}</Alert>;
  }

  return (
    <Box>
      <Button startIcon={<BackIcon />} onClick={() => navigate("/admin/clients")} sx={{ mb: 2 }}>
        {t("admin.backToClients")}
      </Button>

      <Typography variant="h4" gutterBottom>
        {client.firstName} {client.lastName}
      </Typography>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
            <Typography variant="h6">{t("admin.personalInfo")}</Typography>
            <StatusBadge status={client.status} size="medium" />
          </Box>
          <Divider sx={{ mb: 2 }} />
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("auth.email")}</Typography>
              <Typography>{client.email}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">DNI</Typography>
              <Typography>{client.dni}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("registration.phone")}</Typography>
              <Typography>{client.phone}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("registration.address")}</Typography>
              <Typography>{client.address}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("registration.city")}</Typography>
              <Typography>{client.city}, {client.province}</Typography>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <Typography variant="body2" color="text.secondary">{t("admin.registrationDate")}</Typography>
              <Typography>{format(new Date(client.createdAt), "dd/MM/yyyy")}</Typography>
            </Grid>
          </Grid>
        </CardContent>
      </Card>
    </Box>
  );
};

export default ClientDetail;
