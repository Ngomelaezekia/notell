import axios from "axios";

const baseURL = (import.meta.env.VITE_MUSIC_SERVICE_URL || "").replace(/\/$/, "");

const musicAPI = axios.create({
  baseURL: baseURL || undefined,
  timeout: 10000,
});

export const musicServiceAPI = {
  search: async (query = "", { country = "TZ", limit = 20, offset = 0 } = {}) => {
    const response = await musicAPI.get("/api/music/search", { params: { q: query, country, limit, offset } });
    return response.data;
  },
  trending: async ({ country = "TZ", city = "", days = 7 } = {}) => {
    const response = await musicAPI.get("/api/music/trending", { params: { country, city, days } });
    return response.data;
  },
  segment: async ({ trackId, startSec, endSec, country = "TZ" }) => {
    const response = await musicAPI.get("/api/music/tracks/segment", { params: { trackId, startSec, endSec, country } });
    return response.data;
  },
  track: async (trackId, country = "TZ") => {
    const response = await musicAPI.get(`/api/music/tracks/${encodeURIComponent(trackId)}`, { params: { country } });
    return response.data;
  },
  event: async (payload) => {
    const response = await musicAPI.post("/api/music/events", payload);
    return response.data;
  },
};

export default musicServiceAPI;
