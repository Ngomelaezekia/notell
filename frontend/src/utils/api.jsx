import axios from "axios";

const getEnvVar = (viteKey, craKey, fallback) => {
  if (typeof import.meta !== "undefined" && import.meta.env && import.meta.env[viteKey]) {
    return import.meta.env[viteKey];
  }
  if (typeof process !== "undefined" && process.env && process.env[craKey]) {
    return process.env[craKey];
  }
  return fallback;
};

export const SERVER_URL = getEnvVar(
  "VITE_SERVER_URL",
  "REACT_APP_SERVER_URL",
  "http://localhost:8080"
);
export const API_BASE_URL = getEnvVar(
  "VITE_API_URL",
  "REACT_APP_API_URL",
  "http://localhost:8080/api"
);

export const getApiErrorMessage = (error, fallback = "Something went wrong") => {
  return (
    error?.response?.data?.message ||
    error?.response?.data?.error ||
    error?.message ||
    fallback
  );
};

export const getFileUrl = (path) => {
  if (!path) return "";
  if (/^https?:\/\//i.test(path)) return path;
  const cleanPath = path.startsWith("/") ? path : `/${path}`;
  return `${SERVER_URL}${cleanPath}`;
};

const API = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
});

API.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status;

    if (status === 401 && typeof window !== "undefined") {
      const path = window.location.pathname;
      const isPublicRoute = path === "/auth" || path.startsWith("/auth/");

      if (!isPublicRoute) {
        window.dispatchEvent(new CustomEvent("notell:session-expired"));
      }
    }

    return Promise.reject(error);
  }
);

export default API;
