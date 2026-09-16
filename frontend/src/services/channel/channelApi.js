import axios from "axios";

const baseURL = import.meta.env.VITE_CHANNEL_SERVICE_URL || "http://localhost:8082";
const channelClient = axios.create({ baseURL, withCredentials: true, headers: { "Content-Type": "application/json" } });

export const channelAPI = {
  list: async (params = {}) => (await channelClient.get("/v1/channels", { params })).data,
  get: async (id) => (await channelClient.get(`/v1/channels/${id}`)).data,
  programs: async (id) => (await channelClient.get(`/v1/channels/${id}/programs`)).data,
  content: async (id) => (await channelClient.get(`/v1/channels/${id}/content`)).data,
  schedule: async (id) => (await channelClient.get(`/v1/channels/${id}/schedule`)).data,
  create: async (payload) => (await channelClient.post("/v1/channels", payload)).data,
  update: async (id, payload) => (await channelClient.patch(`/v1/channels/${id}`, payload)).data,
  remove: async (id) => (await channelClient.delete(`/v1/channels/${id}`)).data,
  follow: async (id) => (await channelClient.post(`/v1/channels/${id}/follow`)).data,
  unfollow: async (id) => (await channelClient.delete(`/v1/channels/${id}/follow`)).data,
  subscribe: async (id) => (await channelClient.post(`/v1/channels/${id}/subscribe`)).data,
  unsubscribe: async (id) => (await channelClient.delete(`/v1/channels/${id}/subscribe`)).data,
  studioPrograms: async (id) => (await channelClient.get(`/v1/channels/${id}/studio/programs`)).data,
  createProgram: async (id, payload) => (await channelClient.post(`/v1/channels/${id}/studio/programs`, payload)).data,
};

export default channelAPI;
