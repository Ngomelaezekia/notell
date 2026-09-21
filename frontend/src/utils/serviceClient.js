import axios from "axios";

export const createServiceClient = (baseURL) => {
  const client = axios.create({
    baseURL: String(baseURL || "").replace(/\/$/, ""),
    withCredentials: true,
    headers: { "Content-Type": "application/json" },
  });

  client.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error?.response?.status === 401 && typeof window !== "undefined") {
        const path = window.location.pathname;
        if (!path.startsWith("/auth")) {
          window.dispatchEvent(new CustomEvent("notell:session-expired"));
        }
      }
      return Promise.reject(error);
    }
  );

  return client;
};

export default createServiceClient;
