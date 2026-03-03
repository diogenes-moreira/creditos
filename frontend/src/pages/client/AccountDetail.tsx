import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Box, Typography, Card, CardContent, Chip } from "@mui/material";
import { AccountBalance as AccountIcon } from "@mui/icons-material";
import { useTranslation } from "react-i18next";
import { format } from "date-fns";
import { getAccount, getMovements } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import type { Movement } from "../../api/types";

const AccountDetail: React.FC = () => {
  const { t } = useTranslation();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);

  const { data: account, isLoading: accountLoading } = useQuery({
    queryKey: ["account"],
    queryFn: getAccount,
  });

  const { data: movementsData, isLoading: movementsLoading } = useQuery({
    queryKey: ["movements", page, pageSize],
    queryFn: () => getMovements(page + 1, pageSize),
  });

  const columns: Column<Movement>[] = [
    {
      id: "createdAt",
      label: t("common.date"),
      minWidth: 120,
      render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy HH:mm"),
    },
    {
      id: "type",
      label: t("common.type"),
      render: (row) => (
        <Chip
          label={row.type}
          size="small"
          color={row.amount > 0 ? "success" : "error"}
          variant="outlined"
        />
      ),
    },
    {
      id: "description",
      label: t("common.description"),
      minWidth: 200,
    },
    {
      id: "amount",
      label: t("common.amount"),
      align: "right",
      render: (row) => (
        <MoneyDisplay
          amount={row.amount}
          color={row.amount > 0 ? "success.main" : "error.main"}
          fontWeight={500}
        />
      ),
    },
    {
      id: "balance",
      label: t("account.balance"),
      align: "right",
      render: (row) => <MoneyDisplay amount={row.balance} fontWeight={500} />,
    },
  ];

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        {t("account.title")}
      </Typography>

      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Box display="flex" alignItems="center" gap={2}>
            <Box
              sx={{
                width: 64,
                height: 64,
                borderRadius: 2,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                bgcolor: "primary.main",
                color: "white",
              }}
            >
              <AccountIcon sx={{ fontSize: 32 }} />
            </Box>
            <Box>
              <Typography variant="body2" color="text.secondary">
                {t("account.availableBalance")}
              </Typography>
              {accountLoading ? (
                <Typography variant="h4">{t("common.loading")}</Typography>
              ) : (
                <MoneyDisplay amount={account?.balance || 0} variant="h4" fontWeight={700} color="primary.main" />
              )}
              <Typography variant="caption" color="text.secondary">
                {t("common.status")}: {account?.status || "N/A"} | {t("account.currency")}: {account?.currency || "ARS"}
              </Typography>
            </Box>
          </Box>
        </CardContent>
      </Card>

      <Typography variant="h6" gutterBottom>
        {t("account.movements")}
      </Typography>
      <DataTable
        columns={columns}
        rows={movementsData?.data || []}
        total={movementsData?.total || 0}
        page={page}
        pageSize={pageSize}
        onPageChange={setPage}
        onPageSizeChange={(size) => {
          setPageSize(size);
          setPage(0);
        }}
        loading={movementsLoading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("account.noMovements")}
      />
    </Box>
  );
};

export default AccountDetail;
