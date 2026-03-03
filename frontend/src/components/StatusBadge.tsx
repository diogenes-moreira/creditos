import React from "react";
import { Chip } from "@mui/material";
import { useTranslation } from "react-i18next";

interface StatusBadgeProps {
  status: string;
  size?: "small" | "medium";
}

const statusColorMap: Record<string, "default" | "primary" | "secondary" | "error" | "info" | "success" | "warning"> = {
  pending: "warning",
  approved: "success",
  active: "info",
  disbursed: "primary",
  overdue: "error",
  paid: "success",
  cancelled: "default",
  rejected: "error",
  completed: "success",
  defaulted: "error",
  quoted: "default",
  partial: "warning",
  blocked: "error",
};

const StatusBadge: React.FC<StatusBadgeProps> = ({ status, size = "small" }) => {
  const { t } = useTranslation();
  const color = statusColorMap[status.toLowerCase()] || "default";
  const label = t(`status.${status.toLowerCase()}`, status);

  return (
    <Chip
      label={label}
      color={color}
      size={size}
      variant="filled"
      sx={{ fontWeight: 500 }}
    />
  );
};

export default StatusBadge;
