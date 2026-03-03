import React from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Box, Typography } from "@mui/material";
import { format } from "date-fns";
import { getPayments } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import StatusBadge from "../../components/StatusBadge";
import type { Payment } from "../../api/types";

const PaymentHistory: React.FC = () => {
  const { t } = useTranslation();
  const { data: payments, isLoading } = useQuery({
    queryKey: ["payments"],
    queryFn: getPayments,
  });

  const columns: Column<Payment>[] = [
    {
      id: "paidAt",
      label: t("payments.paymentDate"),
      minWidth: 130,
      render: (row) => format(new Date(row.paidAt), "dd/MM/yyyy HH:mm"),
    },
    {
      id: "loanId",
      label: t("payments.loan"),
      render: (row) => `#${row.loanId.slice(0, 8)}`,
    },
    {
      id: "installmentNumber",
      label: t("payments.installmentNumber"),
      align: "center",
    },
    {
      id: "amount",
      label: t("common.amount"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.amount} fontWeight={500} />,
    },
    {
      id: "method",
      label: t("payments.method"),
      render: (row) => {
        const methods: Record<string, string> = {
          transfer: t("payments.transfer"),
          mercadopago: t("payments.mercadoPago"),
          cash: t("payments.cash"),
        };
        return methods[row.method] || row.method;
      },
    },
    {
      id: "status",
      label: t("common.status"),
      render: (row) => <StatusBadge status={row.status} />,
    },
    {
      id: "adjustedAmount",
      label: t("payments.adjustment"),
      align: "right",
      render: (row) =>
        row.adjustedAmount ? (
          <MoneyDisplay amount={row.adjustedAmount} color="warning.main" fontWeight={500} />
        ) : (
          <Typography variant="body2" color="text.secondary">-</Typography>
        ),
    },
  ];

  return (
    <Box>
      <Typography variant="h4" gutterBottom>{t("payments.title")}</Typography>
      <Typography variant="body1" color="text.secondary" mb={3}>
        {t("payments.allPayments")}
      </Typography>

      <DataTable
        columns={columns}
        rows={payments || []}
        loading={isLoading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("payments.noPayments")}
      />
    </Box>
  );
};

export default PaymentHistory;
