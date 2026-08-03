import axios from "axios";

// Environment Variable Resolution (Supports Vite and CRA)
const getEnvVar = (viteKey, craKey, fallback) => {
  if (typeof import.meta !== "undefined" && import.meta.env && import.meta.env[viteKey]) {
    return import.meta.env[viteKey];
  }
  if (typeof process !== "undefined" && process.env && process.env[craKey]) {
    return process.env[craKey];
  }
  return fallback;
};

export const SERVER_URL = getEnvVar("VITE_SERVER_URL", "REACT_APP_SERVER_URL", "http://localhost:8080");
export const API_BASE_URL = getEnvVar("VITE_API_URL", "REACT_APP_API_URL", "http://localhost:8080/api");

// Token Storage Helpers (Imported directly by AuthContext)
export const getToken = () => localStorage.getItem("token");
export const setToken = (token) => localStorage.setItem("token", token);
export const removeToken = () => localStorage.removeItem("token");

// Helper to convert relative media paths from Go backend (e.g., "/uploads/123.jpg") into full browser URLs
export const getFileUrl = (path) => {
  if (!path) return "";
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }
  // Remove leading slash if present to avoid double slash
  const cleanPath = path.startsWith("/") ? path : `/${path}`;
  return `${SERVER_URL}${cleanPath}`;
};

// Primary Axios Instance
const API = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

// Interceptor: Attach JWT Bearer Token to outgoing requests automatically
API.interceptors.request.use(
  (config) => {
    const token = getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Interceptor: Handle global response errors (e.g., Token expiration)
API.interceptors.response.use(
  (response) => response,
  (error) => {
    // Automatically purge token if backend returns 401 Unauthorized
    if (error.response && error.response.status === 401) {
      removeToken();
    }
    return Promise.reject(error);
  }
);

export default API;