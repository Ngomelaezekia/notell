import axios from "axios";

const baseURL = import.meta.env.VITE_MESSAGE_SERVICE_URL || "http://localhost:8083";
const messageClient = axios.create({ baseURL, withCredentials: true, headers: { "Content-Type": "application/json" } });

export const messageAPI = {
  listConversations: async () => (await messageClient.get("/v1/conversations")).data,
  createConversation: async (payload) => (await messageClient.post("/v1/conversations", payload)).data,
  history: async (id, limit = 50) => (await messageClient.get(`/v1/conversations/${id}/messages`, { params: { limit } })).data,
  markRead: async (id, messageId) => (await messageClient.post(`/v1/conversations/${id}/read`, { messageId })).data,
  createCall: async (id) => (await messageClient.post(`/v1/conversations/${id}/calls`)).data,
  callToken: async (id) => (await messageClient.post(`/v1/calls/${id}/token`)).data,
  endCall: async (id) => (await messageClient.post(`/v1/calls/${id}/end`)).data,
  socketURL: (id) => {
    const url = new URL(`${baseURL}/v1/ws/${id}`);
    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    return url.toString();
  },
};

export default messageAPI;
