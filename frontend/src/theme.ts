import { createTheme } from "@mui/material/styles";

const theme = createTheme({
  palette: {
    primary: {
      main: "#1565C0",
      light: "#1E88E5",
      dark: "#0D47A1",
      contrastText: "#FFFFFF",
    },
    secondary: {
      main: "#2E7D32",
      light: "#43A047",
      dark: "#1B5E20",
      contrastText: "#FFFFFF",
    },
    background: {
      default: "#F5F5F5",
      paper: "#FFFFFF",
    },
    error: {
      main: "#D32F2F",
    },
    warning: {
      main: "#ED6C02",
    },
    success: {
      main: "#2E7D32",
    },
  },
  typography: {
    fontFamily: "Roboto, Arial, sans-serif",
    h4: {
      fontWeight: 600,
    },
    h5: {
      fontWeight: 600,
    },
    h6: {
      fontWeight: 600,
    },
  },
  shape: {
    borderRadius: 8,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: "none",
          fontWeight: 500,
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          boxShadow: "0 2px 8px rgba(0,0,0,0.1)",
        },
      },
    },
  },
});

export default theme;
