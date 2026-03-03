import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Box, Typography } from "@mui/material";
import { format } from "date-fns";
import { getVendorPurchases } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import MoneyDisplay from "../../components/MoneyDisplay";
import type { Purchase } from "../../api/types";

const PurchaseHistory: React.FC = () => {
  const { t } = useTranslation();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);

  const { data: purchasesData, isLoading } = useQuery({
    queryKey: ["vendor-purchases", page, pageSize],
    queryFn: () => getVendorPurchases(page + 1, pageSize),
  });

  const columns: Column<Purchase>[] = [
    {
      id: "createdAt",
      label: t("common.date"),
      minWidth: 130,
      render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy HH:mm"),
    },
    {
      id: "clientName",
      label: t("vendor.client"),
      minWidth: 160,
      render: (row) => row.clientName || row.clientId.slice(0, 8),
    },
    {
      id: "amount",
      label: t("common.amount"),
      align: "right",
      render: (row) => (
        <MoneyDisplay amount={parseFloat(row.amount)} fontWeight={500} />
      ),
    },
    {
      id: "description",
      label: t("common.description"),
      minWidth: 200,
    },
    {
      id: "creditLineId",
      label: t("vendor.creditLine"),
      render: (row) => `#${row.creditLineId.slice(0, 8)}`,
    },
  ];

  return (
    <Box>
      <Typography variant="h4" fontWeight={700} gutterBottom>
        {t("vendor.purchaseHistory")}
      </Typography>
      <Typography variant="body1" color="text.secondary" mb={3}>
        {t("vendor.allPurchases")}
      </Typography>

      <DataTable
        columns={columns}
        rows={purchasesData?.data || []}
        total={purchasesData?.total || 0}
        page={page}
        pageSize={pageSize}
        onPageChange={setPage}
        onPageSizeChange={(size) => {
          setPageSize(size);
          setPage(0);
        }}
        loading={isLoading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("vendor.noPurchases")}
      />
    </Box>
  );
};

export default PurchaseHistory;
