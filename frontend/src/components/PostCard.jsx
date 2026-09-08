import { MoreHorizontal, Trash2, Loader2, Heart, MessageSquare } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { usePostActions } from "../hooks/usePosts";
import { useAuth } from "../context/AuthContext";
import { getFileUrl } from "../utils/api";
import CommentSection from "./CommentSection";

export const PostCard = ({ post, onPostDeleted }) => {
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const { deletePost, toggleLike, loading } = usePostActions();
  const [showMenu, setShowMenu] = useState(false);
  const [showComments, setShowComments] = useState(false);
  const [liked, setLiked] = useState(post?.liked ?? false);
  const [likeCount, setLikeCount] = useState(post?.likeCount ?? 0);
  const [likeLoading, setLikeLoading] = useState(false);

  const postId = post?.postId;
  const author = post?.user ?? {};
  const avatar = getFileUrl(author?.profilePicture);
  const mediaUrl = getFileUrl(post?.contentUrl);
  const username = author?.username || "Anonymous";
  const authorId = author?.id ?? post?.userId;
  const isOwner = Boolean(currentUser?.id && currentUser.id === post?.userId);

  useEffect(() => {
    setLiked(Boolean(post?.liked));
    setLikeCount(Number(post?.likeCount ?? 0));
  }, [post?.postId, post?.liked, post?.likeCount]);

  const openAuthorProfile = () => {
    if (!authorId) return;
    navigate(`/users/${authorId}`);
  };

  const handleDelete = async () => {
    if (!postId || !window.confirm("Delete this post?")) return;
    try {
      await deletePost(postId);
      onPostDeleted?.(postId);
    } catch (error) {
      console.error(error);
    } finally {
      setShowMenu(false);
    }
  };

  const handleLike = async () => {
    if (!postId || likeLoading) return;
    setLikeLoading(true);
    try {
      const response = await toggleLike(postId);
      const nextLiked = Boolean(response?.liked);
      setLiked(nextLiked);

      if (typeof response?.likeCount === "number") {
        setLikeCount(response.likeCount);
      } else {
        setLikeCount((current) => {
          if (nextLiked === liked) return current;
          return Math.max(0, current + (nextLiked ? 1 : -1));
        });
      }
    } catch (error) {
      console.error(error);
    } finally {
      setLikeLoading(false);
    }
  };

  return (
    <article className="border-b border-neutral-800 py-5 first:pt-5 last:border-b-0">
      <header className="flex items-center justify-between px-1">
        <button
          type="button"
          onClick={openAuthorProfile}
          disabled={!authorId}
          className="group flex min-w-0 items-center gap-3 rounded-xl text-left transition disabled:cursor-default"
          aria-label={`View ${username}'s profile`}
        >
          <div className="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full border border-neutral-700 bg-neutral-800 font-semibold text-neutral-300 transition group-hover:border-neutral-500 group-hover:ring-2 group-hover:ring-neutral-800">
            {avatar ? (
              <img src={avatar} alt={`${username} avatar`} className="h-full w-full object-cover" />
            ) : (
              username.charAt(0).toUpperCase()
            )}
          </div>
          <div className="min-w-0">
            <h3 className="truncate text-sm font-semibold text-neutral-100 group-hover:underline">
              {username}
            </h3>
            <p className="text-xs text-neutral-500">
              {post?.createdAt ? new Date(post.createdAt).toLocaleDateString() : ""}
            </p>
          </div>
        </button>

        {isOwner && (
          <div className="relative">
            <button
              type="button"
              onClick={() => setShowMenu((previous) => !previous)}
              disabled={loading}
              aria-label="Post options"
              className="rounded-full p-2 text-neutral-400 transition hover:bg-neutral-900 hover:text-neutral-200"
            >
              {loading ? <Loader2 size={18} className="animate-spin" /> : <MoreHorizontal size={18} />}
            </button>
            {showMenu && (
              <div className="absolute right-0 z-20 mt-1 w-36 overflow-hidden rounded-xl border border-neutral-800 bg-neutral-950 shadow-xl">
                <button
                  type="button"
                  onClick={handleDelete}
                  disabled={loading}
                  className="flex w-full items-center gap-2 px-4 py-3 text-sm text-red-400 transition hover:bg-red-950/40"
                >
                  <Trash2 size={15} /> Delete
                </button>
              </div>
            )}
          </div>
        )}
      </header>

      {mediaUrl && (
        <div className="mt-3 max-h-[620px] overflow-hidden rounded-xl border border-neutral-800 bg-black">
          {post?.contentType === "video" ? (
            <video src={mediaUrl} controls className="max-h-[620px] w-full object-contain" />
          ) : (
            <img src={mediaUrl} alt={post?.caption || "Post"} className="max-h-[620px] w-full object-contain" />
          )}
        </div>
      )}

      <section className="px-1 pt-3">
        <div className="flex items-center gap-6">
          <button
            type="button"
            onClick={handleLike}
            disabled={likeLoading}
            aria-pressed={liked}
            className={`flex items-center gap-2 text-sm transition disabled:opacity-60 ${
              liked ? "text-red-500" : "text-neutral-400 hover:text-red-400"
            }`}
          >
            {likeLoading ? (
              <Loader2 size={19} className="animate-spin" />
            ) : (
              <Heart size={19} fill={liked ? "currentColor" : "none"} />
            )}
            <span>{likeCount}</span>
          </button>
          <button
            type="button"
            onClick={() => setShowComments((previous) => !previous)}
            className="flex items-center gap-2 text-sm text-neutral-400 transition hover:text-neutral-200"
            aria-expanded={showComments}
          >
            <MessageSquare size={19} />
            <span>Comment</span>
          </button>
        </div>

        {post?.caption && (
          <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-neutral-200">
            <button
              type="button"
              onClick={openAuthorProfile}
              disabled={!authorId}
              className="mr-2 font-semibold text-neutral-100 hover:underline disabled:cursor-default"
            >
              {username}
            </button>
            {post.caption}
          </p>
        )}

        {showComments && <CommentSection postId={postId} />}
      </section>
    </article>
  );
};
