import axios from "axios";

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1",
  headers: { "Content-Type": "application/json" },
  timeout: 15000
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("myfinance_token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  response => response,
  error => {
    const status = error.response?.status;
    const url = String(error.config?.url || "");
    const isAuthAttempt = /\/auth\/(login|register)\/?$/.test(url);
    if (status === 401 && !isAuthAttempt) {
      localStorage.removeItem("myfinance_token");
      localStorage.removeItem("myfinance_user");
      if (window.location.pathname !== "/login") window.location.assign("/login");
    }
    return Promise.reject(error);
  }
);

export function getErrorMessage(error: unknown): string {
  if (typeof error === "object" && error !== null && "response" in error) {
    const response = (error as {response?:{data?:{message?:string;error?:string}}}).response;
    return response?.data?.message || response?.data?.error || "Something went wrong. Please try again.";
  }
  return error instanceof Error ? error.message : "Something went wrong.";
}
