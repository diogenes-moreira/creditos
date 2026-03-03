import React from "react";
import { Card, CardContent, Box, Typography } from "@mui/material";

interface KPICardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  color?: string;
  subtitle?: string;
}

const KPICard: React.FC<KPICardProps> = ({ icon, label, value, color = "primary.main", subtitle }) => {
  return (
    <Card sx={{ height: "100%" }}>
      <CardContent>
        <Box display="flex" alignItems="center" gap={2}>
          <Box
            sx={{
              width: 56,
              height: 56,
              borderRadius: 2,
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              bgcolor: `${color}15`,
              color: color,
            }}
          >
            {icon}
          </Box>
          <Box>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              {label}
            </Typography>
            <Typography variant="h5" fontWeight={700} color={color}>
              {value}
            </Typography>
            {subtitle && (
              <Typography variant="caption" color="text.secondary">
                {subtitle}
              </Typography>
            )}
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
};

export default KPICard;
