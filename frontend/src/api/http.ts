import axios from "axios";

const apiBaseURL = (import.meta.env.VITE_API_BASE_URL || "").trim();

export const http = axios.create({
  baseURL: apiBaseURL || undefined,
});

http.interceptors.request.use((config) => {
  const token = localStorage.getItem("accessToken");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
