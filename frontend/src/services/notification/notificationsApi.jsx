import API from "../../utils/api";

export const notificationsAPI = {
  getAll: async (page = 1, limit = 20) => {
    const response = await API.get("/notifications", {
      params: { page, limit },
    });
    return response.data;
  },

  markRead: async (notificationId) => {
    const response = await API.post(`/notifications/${notificationId}/read`);
    return response.data;
  },

  markAllRead: async () => {
    const response = await API.post("/notifications/read-all");
    return response.data;
  },
};
