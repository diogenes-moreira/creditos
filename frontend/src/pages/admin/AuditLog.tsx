import React, { useState } from "react";
import { useTranslation } from "react-i18next";
import { useQuery } from "@tanstack/react-query";
import { Box, Typography, Chip } from "@mui/material";
import { format } from "date-fns";
import { getAuditLogs } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import type { AuditEntry } from "../../api/types";

const AuditLog: React.FC = () => {
  const { t } = useTranslation();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);

  const { data: auditData, isLoading } = useQuery({
    queryKey: ["admin-audit", page, pageSize],
    queryFn: () => getAuditLogs(page + 1, pageSize),
  });

  const columns: Column<AuditEntry>[] = [
    {
      id: "createdAt",
      label: t("common.date"),
      minWidth: 150,
      render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy HH:mm:ss"),
    },
    {
      id: "userEmail",
      label: t("audit.user"),
      minWidth: 180,
    },
    {
      id: "action",
      label: t("audit.action"),
      minWidth: 130,
      render: (row) => (
        <Chip
          label={row.action}
          size="small"
          variant="outlined"
          color="primary"
        />
      ),
    },
    {
      id: "description",
      label: t("common.description"),
      minWidth: 250,
    },
    {
      id: "ipAddress",
      label: t("audit.ip"),
      minWidth: 120,
    },
    {
      id: "userAgent",
      label: t("audit.userAgent"),
      minWidth: 200,
      render: (row) => (
        <Typography variant="body2" noWrap sx={{ maxWidth: 200 }} title={row.userAgent}>
          {row.userAgent}
        </Typography>
      ),
    },
  ];

  return (
    <Box>
      <Typography variant="h4" gutterBottom>{t("audit.title")}</Typography>
      <Typography variant="body1" color="text.secondary" mb={3}>
        {t("audit.allActions")}
      </Typography>

      <DataTable
        columns={columns}
        rows={auditData?.data || []}
        total={auditData?.total || 0}
        page={page}
        pageSize={pageSize}
        onPageChange={setPage}
        onPageSizeChange={(size) => {
          setPageSize(size);
          setPage(0);
        }}
        loading={isLoading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("audit.noRecords")}
      />
    </Box>
  );
};

export default AuditLog;
