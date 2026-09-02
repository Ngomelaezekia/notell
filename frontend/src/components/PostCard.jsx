import { MoreHorizontal, Trash2, Loader2, Heart, MessageSquare } from "lucide-react";
import { useState } from "react";
import { usePostActions } from "../hooks/usePosts";
import { useAuth } from "../context/AuthContext";
import { getFileUrl } from "../utils/api";
import CommentSection from "./CommentSection";

export const PostCard = ({ post, onPostDeleted }) => {
  const { user: currentUser } = useAuth();
  const { deletePost, toggleLike, loading } = usePostActions();
  const [showMenu, setShowMenu] = useState(false);
  const [showComments, setShowComments] = useState(false);
  const [liked, setLiked] = useState(post?.liked ?? false);

  const postId = post?.postId;
  const author = post?.user ?? {};
  const avatar = getFileUrl(author?.profilePicture);
  const mediaUrl = getFileUrl(post?.contentUrl);
  const username = author?.username || "Anonymous";
  const isOwner = Boolean(currentUser?.id && currentUser.id === post?.userId);

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
    if (!postId) return;
    try {
      const response = await toggleLike(postId);
      setLiked(Boolean(response?.liked));
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <article className="mb-6 overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm transition hover:shadow-md">
      <header className="flex items-center justify-between px-5 py-4">
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center overflow-hidden rounded-full border bg-slate-100 font-semibold text-slate-700">
            {avatar ? <img src={avatar} alt={`${username} avatar`} className="h-full w-full object-cover" /> : username.charAt(0).toUpperCase()}
          </div>
          <div>
            <h3 className="text-sm font-semibold text-slate-900">{username}</h3>
            <p className="text-xs text-slate-500">{post?.createdAt ? new Date(post.createdAt).toLocaleDateString() : ""}</p>
          </div>
        </div>

        {isOwner && (
          <div className="relative">
            <button type="button" onClick={() => setShowMenu((previous) => !previous)} disabled={loading} aria-label="Post options" className="rounded-full p-2 hover:bg-slate-100">
              {loading ? <Loader2 size={18} className="animate-spin" /> : <MoreHorizontal size={18} />}
            </button>
            {showMenu && (
              <div className="absolute right-0 z-20 mt-2 w-40 rounded-xl border bg-white shadow-lg">
                <button type="button" onClick={handleDelete} disabled={loading} className="flex w-full items-center gap-2 px-4 py-3 text-sm text-red-600 hover:bg-red-50">
                  <Trash2 size={15} /> Delete
                </button>
              </div>
            )}
          </div>
        )}
      </header>

      {mediaUrl && (
        <div className="flex max-h-[600px] justify-center overflow-hidden bg-black">
          {post?.contentType === "video" ? <video src={mediaUrl} controls className="w-full object-contain" /> : <img src={mediaUrl} alt={post?.caption || "Post"} className="w-full object-cover" />}
        </div>
      )}

      <section className="px-5 py-4">
        <div className="flex gap-5">
          <button type="button" onClick={handleLike} className={`flex items-center gap-2 text-sm transition ${liked ? "text-red-500" : "text-slate-600 hover:text-red-500"}`}>
            <Heart size={19} fill={liked ? "currentColor" : "none"} /> Like
          </button>
          <button type="button" onClick={() => setShowComments((previous) => !previous)} className="flex items-center gap-2 text-sm text-slate-600 hover:text-indigo-600" aria-expanded={showComments}>
            <MessageSquare size={19} /> Comment
          </button>
        </div>

        {post?.caption && <p className="mt-4 text-sm text-slate-800"><span className="mr-2 font-semibold">{username}</span>{post.caption}</p>}
        {showComments && <CommentSection postId={postId} />}
      </section>
    </article>
  );
};
