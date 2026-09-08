import { MoreHorizontal, Trash2, Loader2, Heart, MessageSquare } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { usePostActions } from "../hooks/usePosts";
import { useAuth } from "../context/AuthContext";
import { getFileUrl } from "../utils/api";
import CommentSection from "./CommentSection";

const formatRelativeTime = (value) => {
  if (!value) return "";

  const date = new Date(value);
  const timestamp = date.getTime();
  if (Number.isNaN(timestamp)) return "";

  const diff = Date.now() - timestamp;
  if (diff < 0) return "just now";

  const seconds = Math.floor(diff / 1000);
  if (seconds < 10) return "just now";
  if (seconds < 60) return `${seconds}s`;

  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m`;

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;

  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d`;

  const weeks = Math.floor(days / 7);
  if (weeks < 5) return `${weeks}w`;

  const months = Math.floor(days / 30);
  if (months < 12) return `${months}mo`;

  const years = Math.floor(days / 365);
  return `${years}y`;
};

const CAPTION_PREVIEW_LENGTH = 220;

export const PostCard = ({ post, onPostDeleted }) => {
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const { deletePost, toggleLike, loading } = usePostActions();
  const [showMenu, setShowMenu] = useState(false);
  const [showComments, setShowComments] = useState(false);
  const [showFullCaption, setShowFullCaption] = useState(false);
  const [liked, setLiked] = useState(post?.liked ?? false);
  const [likeCount, setLikeCount] = useState(post?.likeCount ?? 0);
  const [commentCount, setCommentCount] = useState(post?.commentCount ?? 0);
  const [likeLoading, setLikeLoading] = useState(false);

  const postId = post?.postId;
  const author = post?.user ?? {};
  const avatar = getFileUrl(author?.profilePicture);
  const mediaUrl = getFileUrl(post?.contentUrl);
  const username = author?.username || "Anonymous";
  const authorId = author?.id ?? post?.userId;
  const isOwner = Boolean(currentUser?.id && currentUser.id === post?.userId);
  const caption = post?.caption ?? "";
  const hasLongCaption = caption.length > CAPTION_PREVIEW_LENGTH;
  const visibleCaption = showFullCaption || !hasLongCaption
    ? caption
    : `${caption.slice(0, CAPTION_PREVIEW_LENGTH).trimEnd()}…`;

  useEffect(() => {
    setLiked(Boolean(post?.liked));
    setLikeCount(Number(post?.likeCount ?? 0));
    setCommentCount(Number(post?.commentCount ?? 0));
    setShowFullCaption(false);
  }, [post?.postId, post?.liked, post?.likeCount, post?.commentCount]);

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

    const previousLiked = liked;
    setLikeLoading(true);
    setLiked(!previousLiked);
    setLikeCount((current) => Math.max(0, current + (previousLiked ? -1 : 1)));

    try {
      const response = await toggleLike(postId);
      const nextLiked = Boolean(response?.liked);
      setLiked(nextLiked);

      if (typeof response?.likeCount === "number") {
        setLikeCount(response.likeCount);
      }
    } catch (error) {
      setLiked(previousLiked);
      setLikeCount((current) => Math.max(0, current + (previousLiked ? 1 : -1)));
      console.error(error);
    } finally {
      setLikeLoading(false);
    }
  };

  const handleMediaDoubleClick = () => {
    if (!liked && !likeLoading) handleLike();
  };

  return (
    <article className="border-b border-neutral-800 py-4 first:pt-3 last:border-b-0 sm:py-5 sm:first:pt-4">
      <header className="flex items-center justify-between gap-3 px-1 sm:px-0">
        <button
          type="button"
          onClick={openAuthorProfile}
          disabled={!authorId}
          className="group flex min-w-0 flex-1 items-center gap-2.5 rounded-xl text-left transition disabled:cursor-default sm:gap-3"
          aria-label={`View ${username}'s profile`}
        >
          <div className="relative flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-full border border-neutral-700 bg-neutral-800 text-sm font-semibold text-neutral-300 transition group-hover:border-neutral-500 group-hover:ring-2 group-hover:ring-neutral-800 sm:h-10 sm:w-10">
            {avatar ? (
              <img src={avatar} alt={`${username} avatar`} className="h-full w-full object-cover" />
            ) : (
              username.charAt(0).toUpperCase()
            )}
          </div>
          <div className="min-w-0 leading-tight">
            <div className="flex min-w-0 items-center gap-1.5">
              <h3 className="truncate text-sm font-semibold text-neutral-100 group-hover:underline">
                {username}
              </h3>
            </div>
            <div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-neutral-500 sm:text-xs">
              <span>{formatRelativeTime(post?.createdAt)}</span>
              <span aria-hidden="true" className="text-neutral-700">·</span>
              <span>Notell</span>
            </div>
          </div>
        </button>

        {isOwner && (
          <div className="relative shrink-0">
            <button
              type="button"
              onClick={() => setShowMenu((previous) => !previous)}
              disabled={loading}
              aria-label="Post options"
              aria-expanded={showMenu}
              className={`flex h-9 w-9 items-center justify-center rounded-full text-neutral-500 transition active:scale-95 ${
                showMenu ? "bg-neutral-900 text-neutral-200" : "hover:bg-neutral-900 hover:text-neutral-200"
              }`}
            >
              {loading ? <Loader2 size={18} className="animate-spin" /> : <MoreHorizontal size={19} />}
            </button>
            {showMenu && (
              <div className="absolute right-0 z-20 mt-1 w-40 overflow-hidden rounded-xl border border-neutral-800 bg-neutral-950 p-1 shadow-2xl shadow-black/40">
                <button
                  type="button"
                  onClick={handleDelete}
                  disabled={loading}
                  className="flex w-full items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium text-red-400 transition hover:bg-red-950/40 active:bg-red-950/60"
                >
                  <Trash2 size={15} /> Delete post
                </button>
              </div>
            )}
          </div>
        )}
      </header>

      {mediaUrl && (
        <div
          className="group/media mt-2.5 overflow-hidden rounded-xl border border-neutral-800 bg-neutral-950 sm:mt-3 sm:rounded-2xl"
          onDoubleClick={handleMediaDoubleClick}
        >
          {post?.contentType === "video" ? (
            <video
              src={mediaUrl}
              controls
              playsInline
              className="block max-h-[min(72dvh,620px)] w-full object-contain"
            />
          ) : (
            <img
              src={mediaUrl}
              alt={caption || "Post"}
              className="mx-auto block max-h-[min(72dvh,620px)] w-full object-contain"
              loading="lazy"
              draggable="false"
            />
          )}
        </div>
      )}

      <section className="px-1 pt-2.5 sm:px-0 sm:pt-3">
        <div className="flex items-center justify-between border-b border-neutral-900 pb-2.5 sm:pb-3">
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={handleLike}
              disabled={likeLoading}
              aria-pressed={liked}
              aria-label={liked ? "Unlike post" : "Like post"}
              className={`group flex min-h-9 items-center gap-2 rounded-full px-2.5 text-sm font-medium transition active:scale-95 disabled:cursor-wait disabled:opacity-70 sm:px-3 ${
                liked
                  ? "text-red-500 hover:bg-red-500/10"
                  : "text-neutral-400 hover:bg-neutral-900 hover:text-neutral-100"
              }`}
            >
              <Heart
                size={19}
                strokeWidth={liked ? 2.5 : 2}
                fill={liked ? "currentColor" : "none"}
                className="transition-transform duration-200 group-hover:scale-110"
              />
              <span className="tabular-nums">{likeCount}</span>
            </button>

            <button
              type="button"
              onClick={() => setShowComments((previous) => !previous)}
              className={`group flex min-h-9 items-center gap-2 rounded-full px-2.5 text-sm font-medium transition active:scale-95 sm:px-3 ${
                showComments
                  ? "bg-neutral-900 text-neutral-100"
                  : "text-neutral-400 hover:bg-neutral-900 hover:text-neutral-100"
              }`}
              aria-expanded={showComments}
              aria-label={showComments ? "Hide comments" : "Show comments"}
            >
              <MessageSquare size={19} strokeWidth={2} className="transition-transform duration-200 group-hover:scale-105" />
              <span>Comment</span>
              <span className="tabular-nums">{commentCount}</span>
            </button>
          </div>

          {likeCount > 0 && (
            <span className="pr-1 text-[11px] text-neutral-600 sm:text-xs">
              {likeCount === 1 ? "1 like" : `${likeCount} likes`}
            </span>
          )}
        </div>

        {caption && (
          <div className="mt-2.5 text-sm leading-6 text-neutral-200 sm:mt-3">
            <p className="whitespace-pre-wrap break-words">
              <button
                type="button"
                onClick={openAuthorProfile}
                disabled={!authorId}
                className="mr-2 font-semibold text-neutral-100 hover:underline disabled:cursor-default"
              >
                {username}
              </button>
              {visibleCaption}
            </p>
            {hasLongCaption && (
              <button
                type="button"
                onClick={() => setShowFullCaption((previous) => !previous)}
                className="mt-1 text-xs font-semibold text-neutral-500 transition hover:text-neutral-200"
              >
                {showFullCaption ? "Show less" : "Read more"}
              </button>
            )}
          </div>
        )}

        {showComments && (
          <CommentSection
            postId={postId}
            onCommentCountChange={setCommentCount}
          />
        )}
      </section>
    </article>
  );
};
