import axios from "axios";

const baseURL = import.meta.env.VITE_LIVE_SERVICE_URL || "http://localhost:8083";
const liveClient = axios.create({
  baseURL,
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
});

export const liveAPI = {
  current: async (channelId) => (await liveClient.get(`/v1/channels/${channelId}/live`)).data,
  playbackToken: async (streamId) => (await liveClient.post(`/v1/streams/${streamId}/playback-token`)).data,
  stream: async (streamId) => (await liveClient.get(`/v1/streams/${streamId}`)).data,
  create: async (channelId, payload) => (await liveClient.post(`/v1/channels/${channelId}/streams`, payload)).data,
  start: async (streamId) => (await liveClient.post(`/v1/streams/${streamId}/start`)).data,
  end: async (streamId) => (await liveClient.post(`/v1/streams/${streamId}/end`)).data,
};

export default liveAPI;
