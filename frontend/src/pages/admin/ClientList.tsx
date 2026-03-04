import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Box, Typography, TextField, InputAdornment } from "@mui/material";
import { Search as SearchIcon } from "@mui/icons-material";
import { format } from "date-fns";
import { adminGetClients, adminSearchClients } from "../../api/endpoints";
import DataTable, { Column } from "../../components/DataTable";
import StatusBadge from "../../components/StatusBadge";
import type { Client } from "../../api/types";

const ClientList: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(20);
  const [search, setSearch] = useState("");
  const [searchQuery, setSearchQuery] = useState("");

  const { data: clientsData, isLoading } = useQuery({
    queryKey: ["admin-clients", page, pageSize],
    queryFn: () => adminGetClients(page + 1, pageSize),
    enabled: !searchQuery,
  });

  const { data: searchResults, isLoading: searchLoading } = useQuery({
    queryKey: ["admin-clients-search", searchQuery],
    queryFn: () => adminSearchClients(searchQuery),
    enabled: !!searchQuery,
  });

  const handleSearch = () => {
    setSearchQuery(search);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSearch();
  };

  const columns: Column<Client>[] = [
    {
      id: "lastName",
      label: t("common.name"),
      minWidth: 150,
      render: (row) => (
        <Typography
          variant="body2"
          sx={{ cursor: "pointer", color: "primary.main", fontWeight: 500 }}
          onClick={() => navigate(`/admin/clients/${row.id}`)}
        >
          {row.lastName}, {row.firstName}
        </Typography>
      ),
    },
    { id: "dni", label: "DNI", minWidth: 100 },
    { id: "email", label: "Email", minWidth: 180 },
    { id: "phone", label: t("registration.phone"), minWidth: 120 },
    {
      id: "city",
      label: t("registration.city"),
      render: (row) => `${row.city}, ${row.province}`,
    },
    {
      id: "isBlocked",
      label: t("common.status"),
      render: (row) => <StatusBadge status={row.isBlocked ? "blocked" : "active"} />,
    },
    {
      id: "createdAt",
      label: t("admin.registrationDate"),
      render: (row) => format(new Date(row.createdAt), "dd/MM/yyyy"),
    },
  ];

  const rows = searchQuery ? (searchResults || []) : (clientsData?.data || []);
  const loading = searchQuery ? searchLoading : isLoading;

  return (
    <Box>
      <Typography variant="h4" gutterBottom>{t("nav.clients")}</Typography>

      <Box mb={3}>
        <TextField
          fullWidth
          placeholder={t("admin.searchPlaceholder")}
          value={search}
          onChange={(e) => {
            setSearch(e.target.value);
            if (!e.target.value) setSearchQuery("");
          }}
          onKeyUp={handleKeyPress}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
          sx={{ maxWidth: 500 }}
        />
      </Box>

      <DataTable
        columns={columns}
        rows={rows}
        total={searchQuery ? rows.length : (clientsData?.total || 0)}
        page={page}
        pageSize={pageSize}
        onPageChange={searchQuery ? undefined : setPage}
        onPageSizeChange={searchQuery ? undefined : (size) => { setPageSize(size); setPage(0); }}
        loading={loading}
        keyExtractor={(row) => row.id}
        emptyMessage={t("admin.noClientsFound")}
      />
    </Box>
  );
};

export default ClientList;
