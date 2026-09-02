import { useState, useEffect, useCallback } from "react";
import { postsAPI } from "../services/post/postsApi";

export const usePosts = (page = 1, limit = 10) => {
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState(null);

  const fetchPosts = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await postsAPI.getFeed(page, limit);
      setPosts(response.data || []);
      setHasMore(response.pagination?.hasMore ?? false);
    } catch (err) {
      setError(err.response?.data?.message || "Failed to fetch posts");
    } finally {
      setLoading(false);
    }
  }, [page, limit]);

  const loadMore = useCallback(async () => {
    if (loading || loadingMore || !hasMore || posts.length === 0) return;

    setLoadingMore(true);
    setError(null);

    try {
      const nextPage = Math.floor(posts.length / limit) + 1;
      const response = await postsAPI.getFeed(nextPage, limit);
      const incomingPosts = response.data || [];

      setPosts((current) => {
        const existingIds = new Set(current.map((post) => post.postId));
        const uniquePosts = incomingPosts.filter(
          (post) => !existingIds.has(post.postId)
        );
        return [...current, ...uniquePosts];
      });
      setHasMore(response.pagination?.hasMore ?? false);
    } catch (err) {
      setError(err.response?.data?.message || "Failed to load more posts");
    } finally {
      setLoadingMore(false);
    }
  }, [hasMore, limit, loading, loadingMore, posts.length]);

  useEffect(() => {
    fetchPosts();
  }, [fetchPosts]);

  return {
    posts,
    loading,
    loadingMore,
    hasMore,
    error,
    refetch: fetchPosts,
    loadMore,
  };
};

export const usePostActions = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const createPost = async (postData) => {
    setLoading(true);
    setError(null);
    try {
      return await postsAPI.create(postData);
    } catch (err) {
      const message = err.response?.data?.message || "Failed to create post";
      setError(message);
      throw new Error(message);
    } finally {
      setLoading(false);
    }
  };

  const deletePost = async (postId) => {
    setLoading(true);
    setError(null);
    try {
      return await postsAPI.delete(postId);
    } catch (err) {
      const message = err.response?.data?.message || "Failed to delete post";
      setError(message);
      throw new Error(message);
    } finally {
      setLoading(false);
    }
  };

  const toggleLike = async (postId) => {
    try {
      return await postsAPI.toggleLike(postId);
    } catch (err) {
      const message = err.response?.data?.message || "Failed to update like";
      throw new Error(message);
    }
  };

  const addComment = async (postId, content) => {
    try {
      return await postsAPI.addComment(postId, content);
    } catch (err) {
      const message = err.response?.data?.message || "Failed to add comment";
      throw new Error(message);
    }
  };

  return {
    createPost,
    deletePost,
    toggleLike,
    addComment,
    loading,
    error,
  };
};
