import { useState, useEffect, useCallback, useRef } from "react";
import { postsAPI } from "../services/post/postsApi";
import { getApiErrorMessage } from "../utils/api";

const FEED_CACHE_PREFIX = "notell:feed:";
const FEED_CACHE_TTL = 5 * 60 * 1000;

const readFeedCache = (category) => {
  try {
    const raw = sessionStorage.getItem(`${FEED_CACHE_PREFIX}${category}`);
    if (!raw) return null;
    const cached = JSON.parse(raw);
    if (!cached?.timestamp || Date.now() - cached.timestamp > FEED_CACHE_TTL || !Array.isArray(cached.posts)) return null;
    return cached;
  } catch {
    return null;
  }
};

const writeFeedCache = (category, posts, pagination) => {
  try {
    sessionStorage.setItem(`${FEED_CACHE_PREFIX}${category}`, JSON.stringify({ timestamp: Date.now(), posts, pagination }));
  } catch {
    // Storage can be unavailable/full; the network path still works normally.
  }
};

export const usePosts = (page = 1, limit = 20, category = "all") => {
  const cached = readFeedCache(category);
  const hasContentRef = useRef(Boolean(cached?.posts?.length));
  const [posts, setPosts] = useState(cached?.posts || []);
  const [loading, setLoading] = useState(!cached?.posts?.length);
  const [loadingMore, setLoadingMore] = useState(false);
  const [currentPage, setCurrentPage] = useState(cached?.pagination?.page ?? page);
  const [hasMore, setHasMore] = useState(cached?.pagination?.hasMore ?? true);
  const [error, setError] = useState(null);

  const fetchPosts = useCallback(async () => {
    setLoading(!hasContentRef.current);
    setError(null);
    try {
      const response = await postsAPI.getFeed(page, limit, category);
      const nextPosts = Array.isArray(response.data) ? response.data : [];
      const pagination = response.pagination || {};
      setPosts(nextPosts);
      hasContentRef.current = nextPosts.length > 0;
      setCurrentPage(pagination.page ?? page);
      setHasMore(pagination.hasMore ?? false);
      writeFeedCache(category, nextPosts, pagination);
    } catch (err) {
      setError(getApiErrorMessage(err, "Failed to fetch posts"));
      if (!hasContentRef.current) setHasMore(false);
    } finally {
      setLoading(false);
    }
  }, [category, limit, page]);

  const loadMore = useCallback(async () => {
    if (loading || loadingMore || !hasMore) return;
    setLoadingMore(true);
    setError(null);
    const nextPage = currentPage + 1;
    try {
      const response = await postsAPI.getFeed(nextPage, limit, category);
      const incomingPosts = Array.isArray(response.data) ? response.data : [];
      setPosts((current) => {
        const existingIds = new Set(current.map((post) => post.postId));
        const uniquePosts = incomingPosts.filter((post) => !existingIds.has(post.postId));
        const merged = [...current, ...uniquePosts];
        writeFeedCache(category, merged, response.pagination || { page: nextPage, hasMore: false });
        return merged;
      });
      setCurrentPage(response.pagination?.page ?? nextPage);
      setHasMore(response.pagination?.hasMore ?? false);
    } catch (err) {
      setError(getApiErrorMessage(err, "Failed to load more posts"));
    } finally {
      setLoadingMore(false);
    }
  }, [category, currentPage, hasMore, limit, loading, loadingMore]);

  const removePost = useCallback((postId) => {
    setPosts((current) => {
      const next = current.filter((post) => post.postId !== postId);
      writeFeedCache(category, next, { page: currentPage, hasMore });
      return next;
    });
  }, [category, currentPage, hasMore]);

  useEffect(() => {
    fetchPosts();
  }, [fetchPosts]);

  return { posts, loading, loadingMore, hasMore, error, refetch: fetchPosts, loadMore, removePost };
};

export const usePostActions = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const createPost = async (postData) => {
    setLoading(true); setError(null);
    try { return await postsAPI.create(postData); }
    catch (err) { const message = getApiErrorMessage(err, "Failed to create post"); setError(message); throw new Error(message, { cause: err }); }
    finally { setLoading(false); }
  };

  const deletePost = async (postId) => {
    setLoading(true); setError(null);
    try { return await postsAPI.delete(postId); }
    catch (err) { const message = getApiErrorMessage(err, "Failed to delete post"); setError(message); throw new Error(message, { cause: err }); }
    finally { setLoading(false); }
  };

  const toggleLike = async (postId) => {
    try { return await postsAPI.toggleLike(postId); }
    catch (err) { throw new Error(getApiErrorMessage(err, "Failed to update like"), { cause: err }); }
  };

  const addComment = async (postId, content, parentId = null) => {
    try { return await postsAPI.addComment(postId, content, parentId); }
    catch (err) { throw new Error(getApiErrorMessage(err, "Failed to add comment"), { cause: err }); }
  };

  return { createPost, deletePost, toggleLike, addComment, loading, error };
};
