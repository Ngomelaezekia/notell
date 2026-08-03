import api, {setToken, removeToken} from "../../utils/api";

export const authService ={
  login: async (credentials) => {
    const res = await api.post("/auth/login", credentials);
     setToken(res.data.token);
    return res.data;
  },

  register: async (data) => {
    const res = await api.post("/auth/register", data);
     setToken(res.data.token);
    return res.data;
  },

  logout: async () => {
    try {
      await api.post("/auth/logout");
    } catch {
      // Ignore server errors
    } finally {
      removeToken();
    }
  },

  me: () => api.get("/auth/me"),
  refresh: () => api.post("/auth/refresh"),

  google: () => {
    window.location.href = `${import.meta.env.VITE_API_URL}/auth/google`;
  },
  apple: () => {
    window.location.href = `${import.meta.env.VITE_API_URL}/auth/apple`;
  },
};
