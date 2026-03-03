import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Box, Typography, Card, CardContent, Chip } from "@mui/material";
import { AccountBalance as AccountIcon } from "@mui/icons-material";
import { format } from "date-fns";
import { getVendorAccount, getVendorMovements } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import type { VendorMovement } from "../../api/types";

const VendorBalance: React.FC = () => {
  const { t } = useTranslation();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);

  const { data: account, isLoading: accountLoading } = useQuery({
    queryKey: ["vendor-account"],
    queryFn: getVendorAccount,
  });

  const { data: movementsData, isLoading: movementsLoading } = useQuery({
    queryKey: ["vendor-movements", page, pageSize],
    queryFn: () => getVendorMovements(page + 1, pageSize),
  });

  const columns: Column<VendorMovement>[] = [
    {
      id: "createdAt",
      label: t("common.date"),
      minWidth: 130,
      render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy HH:mm"),
    },
    {
      id: "type",
      label: t("common.type"),
      render: (row) => {
        const numAmount = parseFloat(row.amount);
        return (
          <Chip
            label={row.type}
            size="small"
            color={numAmount > 0 ? "success" : "error"}
            variant="outlined"
          />
        );
      },
    },
    {
      id: "amount",
      label: t("common.amount"),
      align: "right",
      render: (row) => {
        const numAmount = parseFloat(row.amount);
        return (
          <MoneyDisplay
            amount={numAmount}
            color={numAmount > 0 ? "success.main" : "error.main"}
            fontWeight={500}
          />
        );
      },
    },
    {
      id: "balanceAfter",
      label: t("vendor.balanceAfter"),
      align: "right",
      render: (row) => (
        <MoneyDisplay amount={parseFloat(row.balanceAfter)} fontWeight={500} />
      ),
    },
    {
      id: "description",
      label: t("common.description"),
      minWidth: 200,
    },
  ];

  return (
    <Box>
      <Typography variant="h4" fontWeight={700} gutterBottom>
        {t("vendor.balance")}
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
                {t("vendor.currentBalance")}
              </Typography>
              {accountLoading ? (
                <Typography variant="h4">{t("common.loading")}</Typography>
              ) : (
                <MoneyDisplay
                  amount={parseFloat(account?.balance || "0")}
                  variant="h4"
                  fontWeight={700}
                  color="primary.main"
                />
              )}
            </Box>
          </Box>
        </CardContent>
      </Card>

      <Typography variant="h6" gutterBottom>
        {t("vendor.movements")}
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
        emptyMessage={t("vendor.noMovements")}
      />
    </Box>
  );
};

export default VendorBalance;
