import { useCallback, useEffect, useState } from "react";
import { MessageSquare, Send, Loader2, Reply } from "lucide-react";
import { postsAPI } from "../services/post/postsApi";
import { getApiErrorMessage, getFileUrl } from "../utils/api";

const formatDate = (value) => {
  if (!value) return "";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "" : date.toLocaleDateString();
};

const CommentItem = ({ comment, onReply }) => {
  const user = comment?.user ?? {};
  const avatar = getFileUrl(user.profilePicture);

  return (
    <div className="rounded-xl bg-slate-50 p-3">
      <div className="flex gap-3">
        <div className="h-9 w-9 shrink-0 overflow-hidden rounded-full bg-slate-200 text-center text-sm font-semibold leading-9 text-slate-600">
          {avatar ? (
            <img src={avatar} alt={`${user.username ?? "User"} avatar`} className="h-full w-full object-cover" />
          ) : (
            (user.username ?? "U").charAt(0).toUpperCase()
          )}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-slate-900">{user.username ?? "User"}</span>
            <span className="text-xs text-slate-400">{formatDate(comment.createdAt)}</span>
          </div>
          <p className="mt-1 whitespace-pre-wrap break-words text-sm text-slate-700">{comment.content}</p>
          {!comment?.parentId && (
            <button type="button" onClick={() => onReply(comment)} className="mt-2 inline-flex items-center gap-1 text-xs font-medium text-slate-500 hover:text-indigo-600">
              <Reply size={13} /> Reply
            </button>
          )}
        </div>
      </div>

      {comment.replies?.length > 0 && (
        <div className="ml-12 mt-3 space-y-2 border-l-2 border-slate-200 pl-3">
          {comment.replies.map((reply) => <CommentItem key={reply.commentId} comment={reply} onReply={onReply} />)}
        </div>
      )}
    </div>
  );
};

export default function CommentSection({ postId }) {
  const [comments, setComments] = useState([]);
  const [content, setContent] = useState("");
  const [replyTo, setReplyTo] = useState(null);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const loadComments = useCallback(async () => {
    if (!postId) return;
    setLoading(true);
    setError(null);
    try {
      const response = await postsAPI.getComments(postId);
      setComments(response.data ?? []);
    } catch (err) {
      setError(getApiErrorMessage(err, "Failed to load comments."));
    } finally {
      setLoading(false);
    }
  }, [postId]);

  useEffect(() => {
    void loadComments();
  }, [loadComments]);

  const submitComment = async (event) => {
    event.preventDefault();
    const text = content.trim();
    if (!text || submitting) return;

    setSubmitting(true);
    setError(null);
    try {
      const response = await postsAPI.addComment(postId, text, replyTo?.commentId ?? null);
      const newComment = response.data;
      setContent("");
      setReplyTo(null);

      if (newComment?.parentId) {
        setComments((current) => current.map((comment) =>
          comment.commentId === newComment.parentId
            ? { ...comment, replies: [...(comment.replies ?? []), newComment] }
            : comment
        ));
      } else if (newComment) {
        setComments((current) => [...current, newComment]);
      }
    } catch (err) {
      setError(getApiErrorMessage(err, "Failed to add comment."));
    } finally {
      setSubmitting(false);
    }
  };

  const replyCount = comments.reduce((total, comment) => total + (comment.replies?.length ?? 0), 0);
  const totalCount = comments.length + replyCount;

  return (
    <section className="mt-4 border-t border-slate-100 pt-4" aria-label="Comments">
      <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-800">
        <MessageSquare size={17} />
        Comments {totalCount > 0 ? `(${totalCount})` : ""}
      </div>

      {replyTo && (
        <div className="mb-2 flex items-center justify-between rounded-lg bg-indigo-50 px-3 py-2 text-xs text-indigo-700">
          <span>Replying to @{replyTo.user?.username ?? "user"}</span>
          <button type="button" onClick={() => setReplyTo(null)} className="font-semibold hover:underline">Cancel</button>
        </div>
      )}

      <form onSubmit={submitComment} className="flex gap-2">
        <input
          value={content}
          onChange={(event) => setContent(event.target.value)}
          maxLength={2000}
          placeholder={replyTo ? "Write a reply..." : "Write a comment..."}
          className="min-w-0 flex-1 rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none transition focus:border-indigo-400 focus:ring-2 focus:ring-indigo-100"
        />
        <button type="submit" disabled={!content.trim() || submitting} aria-label="Send comment" className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-indigo-600 text-white transition hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-50">
          {submitting ? <Loader2 size={17} className="animate-spin" /> : <Send size={17} />}
        </button>
      </form>

      {error && <p className="mt-2 text-xs text-red-500">{error}</p>}

      <div className="mt-4 space-y-2">
        {loading ? (
          <div className="flex items-center justify-center py-5 text-xs text-slate-400"><Loader2 size={16} className="mr-2 animate-spin" /> Loading comments...</div>
        ) : comments.length === 0 ? (
          <p className="py-3 text-center text-xs text-slate-400">No comments yet. Be the first.</p>
        ) : (
          comments.map((comment) => <CommentItem key={comment.commentId} comment={comment} onReply={setReplyTo} />)
        )}
      </div>
    </section>
  );
}
