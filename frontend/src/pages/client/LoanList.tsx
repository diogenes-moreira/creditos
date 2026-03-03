import React from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Box, Typography, Button } from "@mui/material";
import { Add as AddIcon } from "@mui/icons-material";
import { format } from "date-fns";
import { getLoans } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import StatusBadge from "../../components/StatusBadge";
import type { Loan } from "../../api/types";

const LoanList: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const { data: loans, isLoading } = useQuery({
    queryKey: ["loans"],
    queryFn: getLoans,
  });

  const columns: Column<Loan>[] = [
    {
      id: "id",
      label: "ID",
      minWidth: 100,
      render: (row) => (
        <Typography
          variant="body2"
          sx={{ cursor: "pointer", color: "primary.main", fontWeight: 500 }}
          onClick={() => navigate(`/loans/${row.id}`)}
        >
          #{row.id.slice(0, 8)}
        </Typography>
      ),
    },
    {
      id: "amount",
      label: t("common.amount"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.amount} fontWeight={500} />,
    },
    {
      id: "totalAmount",
      label: t("loans.totalPayment"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.totalAmount} fontWeight={500} />,
    },
    {
      id: "installments",
      label: t("loans.installments"),
      align: "center",
    },
    {
      id: "amortizationType",
      label: t("loans.system"),
      render: (row) => (row.amortizationType === "french" ? t("loans.french") : t("loans.german")),
    },
    {
      id: "interestRate",
      label: t("loans.interestRate"),
      align: "center",
      render: (row) => `${row.interestRate}%`,
    },
    {
      id: "status",
      label: t("common.status"),
      render: (row) => <StatusBadge status={row.status} />,
    },
    {
      id: "createdAt",
      label: t("common.date"),
      render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy"),
    },
  ];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">{t("loans.title")}</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => navigate("/loans/apply")}
        >
          {t("loans.requestCredit")}
        </Button>
      </Box>

      <DataTable
        columns={columns}
        rows={loans || []}
        loading={isLoading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("loans.noLoans")}
      />
    </Box>
  );
};

export default LoanList;
