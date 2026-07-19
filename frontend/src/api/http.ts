import axios, { InternalAxiosRequestConfig } from "axios";
import {
  getStoredAccessToken,
  getStoredRefreshToken,
  removeStoredAccessToken,
  removeStoredRefreshToken,
  setStoredAccessToken,
  setStoredRefreshToken,
} from "@/lib/browserStorage";

const apiBaseURL = (import.meta.env.VITE_API_BASE_URL || "").trim();

export const http = axios.create({
  baseURL: apiBaseURL || undefined,
});

http.interceptors.request.use((config) => {
  const token = getStoredAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// --- 401 Auto-Refresh Interceptor ---

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (err: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach(({ resolve, reject }) => {
    if (token) resolve(token);
    else reject(error);
  });
  failedQueue = [];
}

http.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    // Only retry on 401, skip if already retried or if it's the refresh endpoint itself
    if (
      error.response?.status !== 401 ||
      originalRequest._retry ||
      originalRequest.url?.includes("/auth/refresh")
    ) {
      return Promise.reject(error);
    }

    if (isRefreshing) {
      // Queue this request while refresh is in progress
      return new Promise((resolve, reject) => {
        failedQueue.push({
          resolve: (token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            resolve(http(originalRequest));
          },
          reject,
        });
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      const refreshToken = getStoredRefreshToken();
      if (!refreshToken) {
        throw new Error("No refresh token");
      }

      const res = await http.post("/auth/refresh", { refreshToken });
      const { accessToken, refreshToken: newRefreshToken } = res.data;

      setStoredAccessToken(accessToken);
      if (newRefreshToken) {
        setStoredRefreshToken(newRefreshToken);
      }

      originalRequest.headers.Authorization = `Bearer ${accessToken}`;
      processQueue(null, accessToken);
      return http(originalRequest);
    } catch (refreshError) {
      processQueue(refreshError, null);
      // Clear tokens and redirect to login only for admin routes
      removeStoredAccessToken();
      removeStoredRefreshToken();
      if (originalRequest.url?.includes("/admin") || window.location.pathname.startsWith("/admin")) {
        window.location.href = "/admin/login";
      }
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  }
);
