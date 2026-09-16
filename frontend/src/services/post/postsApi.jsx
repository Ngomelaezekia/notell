import API from "../../utils/api";

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

export const postsAPI = {
  getFeed: async (page = 1, limit = 20, category = "all") => {
    const response = await API.get("/posts/feed", {
      params: { page, limit, category },
    });
    return response.data;
  },

  getById: async (id) => {
    const response = await API.get(`/posts/${id}`);
    return response.data;
  },

  recordView: async (id) => {
    const response = await API.post(`/posts/${id}/view`);
    return response.data;
  },

  searchPosts: async (query, page = 1, limit = 20) => {
    const response = await API.get("/posts/search", {
      params: { q: query, page, limit },
    });
    return response.data;
  },

  create: async (postData) => {
    const payload = {
      contentType: postData.contentType,
      contentUrl: postData.contentUrl,
      caption: postData.caption,
    };
    const maxAttempts = 8;
    for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
      try {
        const response = await API.post("/posts", payload);
        return response.data;
      } catch (error) {
        const status = error?.response?.status;
        const message = String(error?.response?.data?.message || "").toLowerCase();
        const mediaPreparing = status === 409 && (message.includes("media") || message.includes("processing") || message.includes("prepared"));
        if (!mediaPreparing || attempt === maxAttempts) throw error;
        await sleep(750);
      }
    }
    throw new Error("Failed to create post");
  },

  delete: async (id) => {
    const response = await API.delete(`/posts/${id}`);
    return response.data;
  },

  setMusic: async (postId, musicData) => {
    const response = await API.post(`/posts/${postId}/music`, {
      uploadId: musicData.uploadId,
      startSec: musicData.startSec ?? 0,
      endSec: musicData.endSec ?? 0,
      volume: musicData.volume ?? 1,
    });
    return response.data;
  },

  removeMusic: async (postId) => {
    const response = await API.delete(`/posts/${postId}/music`);
    return response.data;
  },

  toggleLike: async (id) => {
    const response = await API.post(`/posts/${id}/like`);
    return response.data;
  },

  getComments: async (id) => {
    const response = await API.get(`/posts/${id}/comments`);
    return response.data;
  },

  addComment: async (id, content, parentId = null) => {
    const response = await API.post(`/posts/${id}/comments`, {
      content,
      ...(parentId ? { parentId } : {}),
    });
    return response.data;
  },
};
