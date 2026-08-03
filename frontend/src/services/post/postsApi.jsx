import API from "../../utils/api"

export const postsAPI = {
  // Matches: protected.GET("/posts/feed")
  getFeed: async (page = 1, limit = 10) => {
    const response = await API.get(`/posts/feed?page=${page}&limit=${limit}`);
    return response.data;
  },

  // Matches: api.GET("/posts/:id")
  getById: async (id) => {
    const response = await API.get(`/posts/${id}`);
    return response.data;
  },

  // Matches: protected.POST("/posts") with multipart/form-data
  create: async (postData) => {
       console.log("payload : ", postData);

    const response = await API.post("/posts", {
       contentType: postData.contentType,
       contentUrl: postData.contentUrl,
       caption: postData.caption,
  });

  return response.data;
},
  // Matches: protected.DELETE("/posts/:id")
  delete: async (id) => {
    const response = await API.delete(`/posts/${id}`);
    return response.data;
  },

  // Matches: protected.POST("/posts/:id/like")
  toggleLike: async (id) => {
    const response = await API.post(`/posts/${id}/like`);
    return response.data;
  },

  // Matches: protected.POST("/posts/:id/comments")
  addComment: async (id, commentText) => {
    const response = await API.post(`/posts/${id}/comments`, { content: commentText });
    return response.data;
  },
};