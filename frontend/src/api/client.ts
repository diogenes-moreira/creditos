import axios from "axios";

const base = import.meta.env.BASE_URL.replace(/\/$/, "");

const apiClient = axios.create({
  baseURL: `${base}/api/v1`,
  headers: {
    "Content-Type": "application/json",
  },
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("token");
      localStorage.removeItem("user");
      window.location.href = `${base}/login`;
    }
    return Promise.reject(error);
  }
);

export default apiClient;
