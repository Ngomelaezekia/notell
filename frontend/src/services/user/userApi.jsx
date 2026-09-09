import API from "../../utils/api";

const normalizeProfilePosts = (user) => {
  if (!user || !Array.isArray(user.posts) || !Number.isFinite(user.postCount)) {
    return user;
  }

  const visiblePosts = user.posts;
  const totalPosts = Math.max(visiblePosts.length, Number(user.postCount));

  // Keep the existing profile grid limited to the server's preview while
  // allowing the existing posts.length stat to represent the real total.
  user.posts = new Proxy(visiblePosts, {
    get(target, property, receiver) {
      if (property === "length") return totalPosts;
      return Reflect.get(target, property, receiver);
    },
  });

  return user;
};

export const userAPI = {
  getProfile: async (userId) => {
    const response = await API.get(`/users/${userId}`);
    const payload = response.data;
    if (payload?.data?.user) {
      payload.data.user = normalizeProfilePosts(payload.data.user);
    }
    return payload;
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
    return response.data;
  },

  unfollowUser: async (userId) => {
    const response = await API.delete(`/users/${userId}/unfollow`);
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
