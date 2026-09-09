import API from "../../utils/api";

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
    const response = await API.post("/posts", {
      contentType: postData.contentType,
      contentUrl: postData.contentUrl,
      caption: postData.caption,
    });
    return response.data;
  },

  delete: async (id) => {
    const response = await API.delete(`/posts/${id}`);
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
