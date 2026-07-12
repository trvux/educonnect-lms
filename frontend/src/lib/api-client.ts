import axios from "axios";

// Trỏ về backend Go (cmd/api), port mặc định 8080 (xem backend/.env PORT).
export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api",
  headers: {
    "Content-Type": "application/json",
  },
});

// Gắn JWT (lưu ở localStorage sau khi login - US1.2) vào mọi request.
apiClient.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = window.localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});
