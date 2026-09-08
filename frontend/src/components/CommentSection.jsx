import { useCallback, useEffect, useRef, useState } from "react";
import { MessageSquare, Send, Loader2, Reply, X } from "lucide-react";
import { postsAPI } from "../services/post/postsApi";
import { getApiErrorMessage, getFileUrl } from "../utils/api";

const formatDate = (value) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const diff = Math.max(0, Date.now() - date.getTime());
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "now";
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d`;
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
};

const CommentSkeleton = () => (
  <div className="flex gap-2.5" aria-hidden="true">
    <div className="h-8 w-8 shrink-0 animate-pulse rounded-full bg-neutral-900" />
    <div className="flex-1 space-y-2">
      <div className="h-12 w-[78%] animate-pulse rounded-2xl bg-neutral-900" />
      <div className="h-2.5 w-12 animate-pulse rounded-full bg-neutral-900" />
    </div>
  </div>
);

const CommentItem = ({ comment, onReply }) => {
  const user = comment?.user ?? {};
  const avatar = getFileUrl(user.profilePicture);

  return (
    <div className="group flex gap-2.5">
      <div className="h-8 w-8 shrink-0 overflow-hidden rounded-full border border-neutral-800 bg-neutral-800 text-center text-xs font-semibold leading-8 text-neutral-300">
        {avatar ? (
          <img src={avatar} alt={`${user.username ?? "User"} avatar`} className="h-full w-full object-cover" />
        ) : (
          (user.username ?? "U").charAt(0).toUpperCase()
        )}
      </div>
      <div className="min-w-0 flex-1">
        <div className="rounded-2xl bg-neutral-900 px-3 py-2.5 transition-colors group-hover:bg-neutral-900/90">
          <div className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
            <span className="text-xs font-semibold text-neutral-100">{user.username ?? "User"}</span>
            <span className="text-[11px] text-neutral-600">{formatDate(comment.createdAt)}</span>
          </div>
          <p className="mt-0.5 whitespace-pre-wrap break-words text-sm leading-5 text-neutral-300">{comment.content}</p>
        </div>
        {!comment?.parentId && (
          <button type="button" onClick={() => onReply(comment)} className="ml-2 mt-1 inline-flex min-h-7 items-center gap-1 rounded-full px-2 text-[11px] font-semibold text-neutral-500 transition hover:bg-neutral-900 hover:text-neutral-200 active:scale-95">
            <Reply size={12} /> Reply
          </button>
        )}
      </div>
    </div>
  );
};

export default function CommentSection({ postId, onCommentCountChange }) {
  const [comments, setComments] = useState([]);
  const [content, setContent] = useState("");
  const [replyTo, setReplyTo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const inputRef = useRef(null);

  const loadComments = useCallback(async () => {
    if (!postId) return;
    setLoading(true);
    setError(null);
    try {
      const response = await postsAPI.getComments(postId);
      const nextComments = response.data ?? [];
      setComments(nextComments);
      const nextCount = nextComments.reduce((total, comment) => total + 1 + (comment.replies?.length ?? 0), 0);
      onCommentCountChange?.(nextCount);
    } catch (err) {
      setError(getApiErrorMessage(err, "Failed to load comments."));
    } finally {
      setLoading(false);
    }
  }, [postId, onCommentCountChange]);

  useEffect(() => {
    void loadComments();
  }, [loadComments]);

  useEffect(() => {
    if (replyTo) {
      requestAnimationFrame(() => inputRef.current?.focus());
    }
  }, [replyTo]);

  const submitComment = async (event) => {
    event.preventDefault();
    const text = content.trim();
    if (!text || submitting) return;

    setSubmitting(true);
    setError(null);
    try {
      await postsAPI.addComment(postId, text, replyTo?.commentId ?? null);
      setContent("");
      setReplyTo(null);
      await loadComments();
    } catch (err) {
      setError(getApiErrorMessage(err, "Failed to add comment."));
    } finally {
      setSubmitting(false);
    }
  };

  const startReply = (comment) => setReplyTo(comment);
  const cancelReply = () => {
    setReplyTo(null);
    inputRef.current?.focus();
  };

  const replyCount = comments.reduce((total, comment) => total + (comment.replies?.length ?? 0), 0);
  const totalCount = comments.length + replyCount;

  return (
    <section className="mt-3 border-t border-neutral-900 pt-3" aria-label="Comments">
      <div className="mb-3 flex items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-sm font-semibold text-neutral-200">
          <MessageSquare size={16} className="text-neutral-500" />
          <span>{totalCount ? `${totalCount} ${totalCount === 1 ? "comment" : "comments"}` : "Comments"}</span>
        </div>
        {replyTo && (
          <button type="button" onClick={cancelReply} className="inline-flex min-h-7 items-center gap-1 rounded-full px-2 text-xs text-neutral-500 transition hover:bg-neutral-900 hover:text-neutral-200">
            <X size={13} /> Cancel
          </button>
        )}
      </div>

      {replyTo && (
        <div className="mb-2.5 flex items-center justify-between rounded-xl border border-neutral-800 bg-neutral-900/70 px-3 py-2 text-xs text-neutral-500">
          <span>Replying to <span className="font-semibold text-neutral-200">@{replyTo.user?.username ?? "user"}</span></span>
          <button type="button" onClick={cancelReply} aria-label="Cancel reply" className="text-neutral-600 transition hover:text-neutral-200">
            <X size={14} />
          </button>
        </div>
      )}

      <form onSubmit={submitComment} className="flex items-center gap-2">
        <input
          ref={inputRef}
          value={content}
          onChange={(event) => setContent(event.target.value)}
          maxLength={2000}
          placeholder={replyTo ? "Write a reply..." : "Add a comment..."}
          aria-label={replyTo ? "Write a reply" : "Add a comment"}
          className="min-w-0 flex-1 rounded-full border border-neutral-800 bg-neutral-900 px-4 py-2.5 text-sm text-neutral-100 outline-none placeholder:text-neutral-600 transition-[border-color,background-color,box-shadow] focus:border-neutral-600 focus:bg-neutral-900/90 focus:ring-2 focus:ring-neutral-800/70"
        />
        <button type="submit" disabled={!content.trim() || submitting} aria-label="Send comment" className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-neutral-100 text-neutral-950 transition-all hover:bg-white active:scale-90 disabled:cursor-not-allowed disabled:opacity-40">
          {submitting ? <Loader2 size={16} className="animate-spin" /> : <Send size={16} />}
        </button>
      </form>

      {error && (
        <div className="mt-2 flex items-center justify-between gap-3 rounded-xl border border-red-950/70 bg-red-950/20 px-3 py-2 text-xs text-red-400">
          <span>{error}</span>
          <button type="button" onClick={loadComments} className="shrink-0 font-semibold text-neutral-300 hover:text-white">Retry</button>
        </div>
      )}

      <div className="mt-3 max-h-[28rem] space-y-3 overflow-y-auto pr-1 no-scrollbar">
        {loading ? (
          <div className="space-y-3 py-2"><CommentSkeleton /><CommentSkeleton /></div>
        ) : comments.length === 0 ? (
          <div className="rounded-xl border border-dashed border-neutral-800 px-4 py-6 text-center">
            <MessageSquare size={20} className="mx-auto mb-2 text-neutral-700" />
            <p className="text-xs text-neutral-500">No comments yet.</p>
            <p className="mt-1 text-[11px] text-neutral-600">Be the first to start the conversation.</p>
          </div>
        ) : (
          comments.map((comment) => (
            <div key={comment.commentId}>
              <CommentItem comment={comment} onReply={startReply} />
              {comment.replies?.length > 0 && (
                <div className="ml-10 mt-2 space-y-2 border-l border-neutral-800 pl-3">
                  {comment.replies.map((reply) => (
                    <CommentItem key={reply.commentId} comment={reply} onReply={startReply} />
                  ))}
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </section>
  );
}
