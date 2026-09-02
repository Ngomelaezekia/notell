import api from "../../utils/api";

export const authService = {
  login: async (credentials) => {
    const response = await api.post("/auth/login", credentials);
    return response.data;
  },

  register: async (data) => {
    const response = await api.post("/auth/register", data);
    return response.data;
  },

  logout: async () => {
    const response = await api.post("/auth/logout");
    return response.data;
  },

  me: () => api.get("/auth/me"),

  google: () => {
    window.location.href = `${import.meta.env.VITE_API_URL}/auth/google`;
  },
};
