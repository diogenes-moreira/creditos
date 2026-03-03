import React from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  Paper,
  Typography,
  Box,
  CircularProgress,
} from "@mui/material";

export interface Column<T> {
  id: string;
  label: string;
  minWidth?: number;
  align?: "left" | "center" | "right";
  render?: (row: T) => React.ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  rows: T[];
  total?: number;
  page?: number;
  pageSize?: number;
  onPageChange?: (page: number) => void;
  onPageSizeChange?: (size: number) => void;
  loading?: boolean;
  emptyMessage?: string;
  keyExtractor?: (row: T) => string;
}

function DataTable<T>({
  columns,
  rows,
  total,
  page = 0,
  pageSize = 10,
  onPageChange,
  onPageSizeChange,
  loading = false,
  emptyMessage = "No hay datos para mostrar",
  keyExtractor,
}: DataTableProps<T>) {
  if (loading) {
    return (
      <Box display="flex" justifyContent="center" py={4}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Paper sx={{ width: "100%", overflow: "hidden" }}>
      <TableContainer sx={{ maxHeight: 600 }}>
        <Table stickyHeader size="small">
          <TableHead>
            <TableRow>
              {columns.map((col) => (
                <TableCell
                  key={col.id}
                  align={col.align || "left"}
                  style={{ minWidth: col.minWidth }}
                  sx={{ fontWeight: 600, bgcolor: "grey.50" }}
                >
                  {col.label}
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columns.length} align="center" sx={{ py: 4 }}>
                  <Typography color="text.secondary">{emptyMessage}</Typography>
                </TableCell>
              </TableRow>
            ) : (
              rows.map((row, idx) => (
                <TableRow
                  hover
                  key={keyExtractor ? keyExtractor(row) : idx}
                  sx={{ "&:last-child td": { border: 0 } }}
                >
                  {columns.map((col) => (
                    <TableCell key={col.id} align={col.align || "left"}>
                      {col.render
                        ? col.render(row)
                        : (row as Record<string, unknown>)[col.id] as React.ReactNode}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>
      {onPageChange && total !== undefined && (
        <TablePagination
          rowsPerPageOptions={[5, 10, 20, 50]}
          component="div"
          count={total}
          rowsPerPage={pageSize}
          page={page}
          onPageChange={(_, newPage) => onPageChange(newPage)}
          onRowsPerPageChange={(e) => onPageSizeChange?.(parseInt(e.target.value, 10))}
          labelRowsPerPage="Filas por pagina"
        />
      )}
    </Paper>
  );
}

export default DataTable;
