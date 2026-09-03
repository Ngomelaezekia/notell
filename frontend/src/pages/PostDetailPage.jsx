import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Loader2 } from "lucide-react";
import { postsAPI } from "../services/post/postsApi";
import { getApiErrorMessage } from "../utils/api";
import { PostCard } from "../components/PostCard";

export default function PostDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [post, setPost] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    let cancelled = false;

    const loadPost = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await postsAPI.getById(id);
        if (!cancelled) setPost(response?.data || null);
      } catch (err) {
        if (!cancelled) {
          setPost(null);
          setError(getApiErrorMessage(err, "Failed to load post."));
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    void loadPost();
    return () => { cancelled = true; };
  }, [id]);

  return (
    <section className="mx-auto w-full max-w-2xl">
      <div className="mb-4">
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="inline-flex items-center gap-2 rounded-xl px-3 py-2 text-sm font-bold text-slate-600 transition hover:bg-white hover:text-slate-900"
        >
          <ArrowLeft size={17} /> Back
        </button>
      </div>

      {loading && (
        <div className="flex items-center justify-center gap-2 rounded-3xl bg-white/80 py-16 text-sm text-slate-500 shadow-xl">
          <Loader2 size={20} className="animate-spin" /> Loading post...
        </div>
      )}

      {!loading && error && (
        <div className="rounded-3xl border border-red-100 bg-red-50 p-6 text-center text-sm text-red-600">
          {error}
        </div>
      )}

      {!loading && !error && post && <PostCard post={post} onPostDeleted={() => navigate("/")} />}
    </section>
  );
}
