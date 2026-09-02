import API from "../../utils/api";

export const userAPI = {
  getProfile: async (userId) => {
    const response = await API.get(`/users/${userId}`);
    return response.data;
  },

  getRelationship: async (userId) => {
    const response = await API.get(`/users/${userId}/relationship`);
    return response.data;
  },

  updateProfile: async (userData) => {
    const response = await API.put("/users/profile", userData);
    return response.data;
  },

  followUser: async (userId) => {
    const response = await API.post(`/users/${userId}/follow`);
    return response.data;
  },

  unfollowUser: async (userId) => {
    const response = await API.delete(`/users/${userId}/unfollow`);
    return response.data;
  },

  getFollowers: async (userId) => {
    const response = await API.get(`/users/${userId}/followers`);
    return response.data;
  },

  getFollowing: async (userId) => {
    const response = await API.get(`/users/${userId}/following`);
    return response.data;
  },
};
