import { MoreHorizontal, Trash2, Loader2, Heart, MessageSquare, Volume2, VolumeX, Maximize2, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { usePostActions } from "../hooks/usePosts";
import { useAuth } from "../context/AuthContext";
import { postsAPI } from "../services/post/postsApi";
import { getFileUrl } from "../utils/api";
import CommentSection from "./CommentSection";

const LONG_VIDEO_SECONDS = 45;
const CAPTION_PREVIEW_LENGTH = 220;

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
  return `${Math.floor(days / 365)}y`;
};

export const PostCard = ({ post, onPostDeleted, priority = false }) => {
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const { deletePost, toggleLike, loading } = usePostActions();
  const mediaContainerRef = useRef(null);
  const videoRef = useRef(null);
  const viewRecordedRef = useRef(false);
  const [showMenu, setShowMenu] = useState(false);
  const [showComments, setShowComments] = useState(false);
  const [showFullCaption, setShowFullCaption] = useState(false);
  const [liked, setLiked] = useState(post?.liked ?? false);
  const [likeCount, setLikeCount] = useState(post?.likeCount ?? 0);
  const [commentCount, setCommentCount] = useState(post?.commentCount ?? 0);
  const [likeLoading, setLikeLoading] = useState(false);
  const [likeBurst, setLikeBurst] = useState(false);
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const [mediaActive, setMediaActive] = useState(priority);
  const [videoDuration, setVideoDuration] = useState(0);
  const [videoMuted, setVideoMuted] = useState(true);

  const postId = post?.postId;
  const author = post?.user ?? {};
  const avatar = getFileUrl(author?.profilePicture);
  const mediaUrl = getFileUrl(post?.contentUrl);
  const isVideo = post?.contentType === "video";
  const username = author?.username || "Anonymous";
  const authorId = author?.id ?? post?.userId;
  const isOwner = Boolean(currentUser?.id && currentUser.id === post?.userId);
  const caption = post?.caption ?? "";
  const hasLongCaption = caption.length > CAPTION_PREVIEW_LENGTH;
  const visibleCaption = showFullCaption || !hasLongCaption ? caption : `${caption.slice(0, CAPTION_PREVIEW_LENGTH).trimEnd()}…`;
  const isLongVideo = isVideo && videoDuration > LONG_VIDEO_SECONDS;

  useEffect(() => {
    setLiked(Boolean(post?.liked));
    setLikeCount(Number(post?.likeCount ?? 0));
    setCommentCount(Number(post?.commentCount ?? 0));
    setShowFullCaption(false);
    setMediaActive(priority);
    setVideoDuration(0);
    setVideoMuted(true);
    viewRecordedRef.current = false;
  }, [post?.postId, post?.liked, post?.likeCount, post?.commentCount, priority]);

  useEffect(() => {
    const target = mediaContainerRef.current;
    if (!target || !postId) return undefined;
    const observer = new IntersectionObserver((entries) => {
      const entry = entries[0];
      if (!entry) return;
      if (entry.isIntersecting && entry.intersectionRatio >= 0.25) setMediaActive(true);
      if (entry.isIntersecting && entry.intersectionRatio >= 0.6 && !viewRecordedRef.current) {
        viewRecordedRef.current = true;
        void postsAPI.recordView(postId).catch(() => {});
      }
      if (isVideo && videoRef.current && !isLongVideo) {
        if (entry.isIntersecting && entry.intersectionRatio >= 0.5) {
          videoRef.current.muted = true;
          setVideoMuted(true);
          void videoRef.current.play().catch(() => {});
        } else if (!entry.isIntersecting) videoRef.current.pause();
      } else if (isVideo && videoRef.current && !entry.isIntersecting) {
        videoRef.current.pause();
      }
    }, { threshold: [0, 0.25, 0.5, 0.6], rootMargin: "180px 0px" });
    observer.observe(target);
    return () => observer.disconnect();
  }, [isVideo, isLongVideo, postId]);

  useEffect(() => {
    if (!showMenu) return undefined;
    const close = (event) => { if (!event.target.closest("[data-post-menu]")) setShowMenu(false); };
    document.addEventListener("click", close);
    return () => document.removeEventListener("click", close);
  }, [showMenu]);

  useEffect(() => {
    if (!lightboxOpen) return undefined;
    const closeOnEscape = (event) => { if (event.key === "Escape") setLightboxOpen(false); };
    document.addEventListener("keydown", closeOnEscape);
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => { document.removeEventListener("keydown", closeOnEscape); document.body.style.overflow = previousOverflow; };
  }, [lightboxOpen]);

  const openAuthorProfile = () => { if (authorId) navigate(`/users/${authorId}`); };
  const openVideoFeed = () => { if (postId) { videoRef.current?.pause(); navigate(`/video-feed/${postId}`); } };

  const handleDelete = async () => {
    if (!postId || !window.confirm("Delete this post?")) return;
    try { await deletePost(postId); onPostDeleted?.(postId); }
    catch (error) { console.error(error); }
    finally { setShowMenu(false); }
  };

  const handleLike = async () => {
    if (!postId || likeLoading) return;
    const previousLiked = liked;
    setLikeLoading(true);
    setLiked(!previousLiked);
    setLikeCount((current) => Math.max(0, current + (previousLiked ? -1 : 1)));
    setLikeBurst(true);
    window.setTimeout(() => setLikeBurst(false), 420);
    try {
      const response = await toggleLike(postId);
      setLiked(Boolean(response?.liked));
      if (typeof response?.likeCount === "number") setLikeCount(response.likeCount);
    } catch (error) {
      setLiked(previousLiked);
      setLikeCount((current) => Math.max(0, current + (previousLiked ? 1 : -1)));
      console.error(error);
    } finally { setLikeLoading(false); }
  };

  const handleMediaDoubleClick = () => { if (!liked && !likeLoading) handleLike(); };
  const toggleAudio = (event) => {
    event.stopPropagation();
    if (!videoRef.current) return;
    const nextMuted = !videoRef.current.muted;
    videoRef.current.muted = nextMuted;
    setVideoMuted(nextMuted);
    if (!nextMuted) void videoRef.current.play().catch(() => {});
  };

  return (
    <>
      <article className="overflow-hidden rounded-2xl border border-neutral-900 bg-neutral-950/70 shadow-[0_8px_30px_rgba(0,0,0,0.12)] transition-[border-color,box-shadow] duration-200 hover:border-neutral-800 hover:shadow-[0_10px_34px_rgba(0,0,0,0.16)] sm:rounded-[1.35rem]">
        <header className="flex items-center justify-between gap-3 px-3 py-3 sm:px-4 sm:py-3.5">
          <button type="button" onClick={openAuthorProfile} disabled={!authorId} className="group flex min-w-0 flex-1 items-center gap-2.5 rounded-xl text-left transition active:scale-[0.99] disabled:cursor-default sm:gap-3" aria-label={`View ${username}'s profile`}>
            <div className="relative flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-full border border-neutral-700 bg-neutral-800 text-sm font-semibold text-neutral-300 transition duration-200 group-hover:border-neutral-500 group-hover:ring-2 group-hover:ring-neutral-800 sm:h-10 sm:w-10">
              {avatar ? <img src={avatar} alt={`${username} avatar`} className="h-full w-full object-cover" /> : username.charAt(0).toUpperCase()}
            </div>
            <div className="min-w-0 leading-tight"><h3 className="truncate text-sm font-semibold text-neutral-100 group-hover:underline">{username}</h3><div className="mt-0.5 flex items-center gap-1.5 text-[11px] text-neutral-500 sm:text-xs"><span>{formatRelativeTime(post?.createdAt)}</span><span aria-hidden="true" className="text-neutral-700">·</span><span>Notell</span></div></div>
          </button>
          {isOwner && <div className="relative shrink-0" data-post-menu><button type="button" onClick={() => setShowMenu((previous) => !previous)} disabled={loading} aria-label="Post options" aria-expanded={showMenu} className={`flex h-9 w-9 items-center justify-center rounded-full transition active:scale-90 ${showMenu ? "bg-neutral-900 text-neutral-100" : "text-neutral-500 hover:bg-neutral-900 hover:text-neutral-200"}`}>{loading ? <Loader2 size={18} className="animate-spin" /> : <MoreHorizontal size={19} />}</button>{showMenu && <div className="absolute right-0 z-30 mt-1 w-40 origin-top-right overflow-hidden rounded-xl border border-neutral-800 bg-neutral-950 p-1 shadow-2xl shadow-black/50"><button type="button" onClick={handleDelete} disabled={loading} className="flex w-full items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium text-red-400 transition hover:bg-red-950/40 active:bg-red-950/60"><Trash2 size={15} /> Delete post</button></div>}</div>}
        </header>

        {mediaUrl && <div ref={mediaContainerRef} className="group/media relative aspect-square w-full overflow-hidden border-y border-neutral-900 bg-black" onDoubleClick={handleMediaDoubleClick}>
          {isVideo ? (
            <button type="button" onClick={openVideoFeed} className="relative block h-full w-full cursor-pointer text-left focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-neutral-300" aria-label="Open video feed viewer">
              <video ref={videoRef} src={mediaUrl} muted defaultMuted playsInline preload={mediaActive && !isLongVideo ? "auto" : "metadata"} onLoadedMetadata={(event) => setVideoDuration(event.currentTarget.duration || 0)} className="block h-full w-full object-cover" aria-label={caption || "Post video"} />
              {isLongVideo && <span className="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/70 to-transparent px-4 pb-4 pt-12 text-xs font-medium text-white/90">Tap to watch full video</span>}
              <span onClick={toggleAudio} className="absolute bottom-3 left-3 z-10 flex h-8 w-8 items-center justify-center rounded-full bg-black/60 text-white shadow-md backdrop-blur-sm transition hover:bg-black/80 active:scale-90 sm:h-9 sm:w-9" role="button" aria-label={videoMuted ? "Allow sound" : "Mute video"}>{videoMuted ? <VolumeX size={15} /> : <Volume2 size={15} />}</span>
              {likeBurst && <span className="pointer-events-none absolute inset-0 flex items-center justify-center"><Heart size={82} fill="currentColor" strokeWidth={1.5} className="scale-125 text-white opacity-0 drop-shadow-[0_4px_18px_rgba(0,0,0,0.55)]" style={{ animation: "notellLikePop 420ms ease-out forwards" }} /></span>}
            </button>
          ) : (
            <button type="button" onClick={() => setLightboxOpen(true)} className="relative block h-full w-full cursor-zoom-in text-left focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-neutral-300" aria-label="Open image viewer">
              <img src={mediaUrl} alt={caption || "Post"} className="block h-full w-full object-cover transition duration-300 group-hover/media:scale-[1.008]" loading={priority ? "eager" : "lazy"} fetchPriority={priority ? "high" : "auto"} decoding="async" draggable="false" />
              <span className="pointer-events-none absolute right-3 top-3 flex h-8 w-8 items-center justify-center rounded-full bg-black/55 text-white opacity-0 backdrop-blur-sm transition-opacity duration-200 group-hover/media:opacity-100 sm:h-9 sm:w-9"><Maximize2 size={15} /></span>
              {likeBurst && <span className="pointer-events-none absolute inset-0 flex items-center justify-center"><Heart size={82} fill="currentColor" strokeWidth={1.5} className="scale-125 text-white opacity-0 drop-shadow-[0_4px_18px_rgba(0,0,0,0.55)]" style={{ animation: "notellLikePop 420ms ease-out forwards" }} /></span>}
            </button>
          )}
        </div>}

        <section className="px-3 pb-3 sm:px-4 sm:pb-4">
          <div className="flex items-center justify-between pt-2.5 sm:pt-3"><div className="flex items-center gap-1"><button type="button" onClick={handleLike} disabled={likeLoading} aria-pressed={liked} aria-label={liked ? "Unlike post" : "Like post"} className={`group flex min-h-9 items-center gap-2 rounded-full px-2.5 text-sm font-medium transition-all duration-200 active:scale-90 disabled:cursor-wait disabled:opacity-70 sm:px-3 ${liked ? "bg-red-500/10 text-red-500 hover:bg-red-500/15" : "text-neutral-400 hover:bg-neutral-900 hover:text-neutral-100"}`}><Heart size={19} strokeWidth={liked ? 2.5 : 2} fill={liked ? "currentColor" : "none"} className={`transition-transform duration-200 group-hover:scale-110 ${likeBurst ? "scale-125" : ""}`} /><span className="tabular-nums">{likeCount}</span></button><button type="button" onClick={() => setShowComments((previous) => !previous)} className={`group flex min-h-9 items-center gap-2 rounded-full px-2.5 text-sm font-medium transition-all duration-200 active:scale-90 sm:px-3 ${showComments ? "bg-neutral-900 text-neutral-100" : "text-neutral-400 hover:bg-neutral-900 hover:text-neutral-100"}`} aria-expanded={showComments} aria-label={showComments ? "Hide comments" : "Show comments"}><MessageSquare size={19} strokeWidth={2} className="transition-transform duration-200 group-hover:scale-105" /><span>Comment</span><span className="tabular-nums">{commentCount}</span></button></div>{likeCount > 0 && <span className="pr-1 text-[11px] text-neutral-600 sm:text-xs">{likeCount === 1 ? "1 like" : `${likeCount} likes`}</span>}</div>
          {caption && <div className="mt-2.5 text-sm leading-6 text-neutral-200 sm:mt-3"><p className="whitespace-pre-wrap break-words"><button type="button" onClick={openAuthorProfile} disabled={!authorId} className="mr-2 font-semibold text-neutral-100 hover:underline disabled:cursor-default">{username}</button>{visibleCaption}</p>{hasLongCaption && <button type="button" onClick={() => setShowFullCaption((previous) => !previous)} className="mt-1 text-sm font-medium text-neutral-500 hover:text-neutral-300">{showFullCaption ? "Show less" : "more"}</button>}</div>}
          {showComments && <div className="mt-3 border-t border-neutral-900 pt-3"><CommentSection postId={postId} onCommentCountChange={setCommentCount} /></div>}
        </section>
      </article>

      {lightboxOpen && !isVideo && <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/95 p-3 sm:p-6" role="dialog" aria-modal="true" aria-label="Image viewer" onClick={() => setLightboxOpen(false)}><button type="button" onClick={() => setLightboxOpen(false)} className="absolute right-3 top-3 z-10 flex h-10 w-10 items-center justify-center rounded-full bg-white/10 text-white backdrop-blur hover:bg-white/15" aria-label="Close image viewer"><X size={20} /></button><img src={mediaUrl} alt={caption || "Post"} className="max-h-[92dvh] max-w-full object-contain" onClick={(event) => event.stopPropagation()} /></div>}
    </>
  );
};
