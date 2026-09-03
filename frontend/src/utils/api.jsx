import axios from "axios";

export const SERVER_URL =
  import.meta.env.VITE_SERVER_URL || "http://localhost:8080";
export const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:8080/api";

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
