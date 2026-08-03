import API from "../../utils/api";

export const userAPI = {
  // Matches: api.GET("/users/:id")
  getProfile: async (userId) => {
    const response = await API.get(`/users/${userId}`);
    return response.data;
  },

  // Matches: protected.PUT("/users/profile")
  updateProfile: async (userData) => {
    const response = await API.put("/users/profile", userData);
    return response.data;
  },

  // Matches: protected.POST("/users/:id/follow")
  followUser: async (userId) => {
    const response = await API.post(`/users/${userId}/follow`);
    return response.data;
  },

  // Matches: protected.DELETE("/users/:id/unfollow")
  unfollowUser: async (userId) => {
    const response = await API.delete(`/users/${userId}/unfollow`);
    return response.data;
  },
};