import React from "react";
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
  CircularProgress,
} from "@mui/material";
import { getVendorProfile, updateVendorProfile } from "../../api/endpoints";
import { useNotification } from "../../contexts/NotificationContext";
import { getErrorMessage } from "../../api/errorUtils";

const profileSchema = z.object({
  phone: z.string().min(1),
  address: z.string().min(1),
  city: z.string().min(1),
  province: z.string().min(1),
});

type ProfileForm = z.infer<typeof profileSchema>;

const VendorProfile: React.FC = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { showSuccess, showError } = useNotification();

  const { data: profile, isLoading } = useQuery({
    queryKey: ["vendor-profile"],
    queryFn: getVendorProfile,
  });

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    values: profile
      ? {
          phone: profile.phone,
          address: profile.address,
          city: profile.city,
          province: profile.province,
        }
      : undefined,
  });

  const mutation = useMutation({
    mutationFn: updateVendorProfile,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vendor-profile"] });
      showSuccess(t("vendor.profileUpdateSuccess"));
    },
    onError: (err) =>
      showError(getErrorMessage(err, t("vendor.profileUpdateError"))),
  });

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={4}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" fontWeight={700} gutterBottom>
        {t("vendor.profile")}
      </Typography>

      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            {t("vendor.businessInfo")}
          </Typography>
          <Divider sx={{ mb: 2 }} />

          <Grid container spacing={2} mb={3}>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label={t("vendor.businessName")}
                value={profile?.businessName || ""}
                InputProps={{ readOnly: true }}
                disabled
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label={t("vendor.cuit")}
                value={profile?.cuit || ""}
                InputProps={{ readOnly: true }}
                disabled
              />
            </Grid>
          </Grid>

          <Typography variant="h6" gutterBottom>
            {t("vendor.contactInfo")}
          </Typography>
          <Divider sx={{ mb: 2 }} />

          <form onSubmit={handleSubmit((data) => mutation.mutate(data))}>
            <Grid container spacing={2}>
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
                        errors.phone ? t("vendor.fieldRequired") : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
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
                        errors.province ? t("vendor.fieldRequired") : undefined
                      }
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12}>
                <Typography variant="body2" color="text.secondary" mb={1}>
                  Email: {profile?.email}
                </Typography>
                <Button
                  type="submit"
                  variant="contained"
                  disabled={mutation.isPending}
                >
                  {mutation.isPending
                    ? t("common.saving")
                    : t("vendor.saveChanges")}
                </Button>
              </Grid>
            </Grid>
          </form>
        </CardContent>
      </Card>

    </Box>
  );
};

export default VendorProfile;
