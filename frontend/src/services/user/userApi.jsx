import API from "../../utils/api";

const getRelationshipData = (response) => response?.data?.data ?? response?.data ?? {};

const publishRelationshipChange = (userId, data) => {
  if (typeof window === "undefined") return;
  window.dispatchEvent(
    new CustomEvent("notell:relationship-changed", {
      detail: { userId: String(userId), ...data },
    })
  );
};

export const userAPI = {
  getProfile: async (userId, page = 1, limit = 36) => {
    const response = await API.get(`/users/${userId}`, {
      params: { page, limit },
    });
    return response.data;
  },

  getRelationship: async (userId) => {
    const response = await API.get(`/users/${userId}/relationship`);
    return response.data;
  },

  searchUsers: async (query, page = 1, limit = 20) => {
    const response = await API.get("/users/search", {
      params: { q: query, page, limit },
    });
    return response.data;
  },

  updateProfile: async (userData) => {
    const response = await API.put("/users/profile", userData);
    return response.data;
  },

  followUser: async (userId) => {
    const response = await API.post(`/users/${userId}/follow`);
    const data = getRelationshipData(response);
    publishRelationshipChange(userId, data);
    return response.data;
  },

  unfollowUser: async (userId) => {
    const response = await API.delete(`/users/${userId}/unfollow`);
    const data = getRelationshipData(response);
    publishRelationshipChange(userId, data);
    return response.data;
  },

  removeFollower: async (userId) => {
    const response = await API.delete(`/users/followers/${userId}`);
    const data = getRelationshipData(response);
    publishRelationshipChange(userId, { following: false, follower: false, ...data });
    return response.data;
  },

  getFollowers: async (userId, page = 1, limit = 20) => {
    const response = await API.get(`/users/${userId}/followers`, {
      params: { page, limit },
    });
    return response.data;
  },

  getFollowing: async (userId, page = 1, limit = 20) => {
    const response = await API.get(`/users/${userId}/following`, {
      params: { page, limit },
    });
    return response.data;
  },
};