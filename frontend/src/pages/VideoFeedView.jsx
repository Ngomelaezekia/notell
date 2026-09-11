import { ArrowLeft, Pause, Play, Volume2, VolumeX } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { postsAPI } from "../services/post/postsApi";
import { getFileUrl } from "../utils/api";

const formatTime = (value) => {
  if (!Number.isFinite(value)) return "0:00";
  const seconds = Math.max(0, Math.floor(value));
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}`;
};

export default function VideoFeedView() {
  const { id } = useParams();
  const navigate = useNavigate();
  const videoRef = useRef(null);
  const [post, setPost] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [playing, setPlaying] = useState(true);
  const [muted, setMuted] = useState(true);
  const [duration, setDuration] = useState(0);
  const [currentTime, setCurrentTime] = useState(0);

  useEffect(() => {
    let cancelled = false;
    const load = async () => {
      setLoading(true);
      setError("");
      try {
        const response = await postsAPI.getById(id);
        if (!cancelled) setPost(response?.data || null);
      } catch (err) {
        if (!cancelled) setError(err?.response?.data?.message || "Unable to load this video.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    load();
    return () => { cancelled = true; };
  }, [id]);

  useEffect(() => {
    const video = videoRef.current;
    if (!video || !post) return;
    video.muted = muted;
    if (playing) video.play().catch(() => setPlaying(false));
    else video.pause();
  }, [post, playing, muted]);

  useEffect(() => {
    const onKeyDown = (event) => {
      if (event.key === "Escape" || event.key === "Backspace") navigate(-1);
      if (event.code === "Space") {
        event.preventDefault();
        setPlaying((value) => !value);
      }
      if (event.key.toLowerCase() === "m") setMuted((value) => !value);
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [navigate]);

  const mediaUrl = getFileUrl(post?.contentUrl);
  const author = post?.user || {};

  return (
    <main className="fixed inset-0 z-[90] flex min-h-dvh flex-col bg-black text-white">
      <div className="absolute inset-x-0 top-0 z-20 bg-gradient-to-b from-black/90 via-black/50 to-transparent px-3 pb-5 pt-3 sm:px-5 sm:pt-5">
        <div className="mx-auto flex max-w-5xl items-center gap-2">
          <button type="button" onClick={() => navigate(-1)} className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-white/10 backdrop-blur-md transition hover:bg-white/20 active:scale-95" aria-label="Back">
            <ArrowLeft size={20} />
          </button>
          <button type="button" onClick={() => author?.id && navigate(`/users/${author.id}`)} className="min-w-0 flex-1 text-left">
            <div className="truncate text-sm font-semibold">{author?.username || "Video"}</div>
            <div className="truncate text-xs text-white/60">{post?.caption || ""}</div>
          </button>
          <button type="button" onClick={() => setMuted((value) => !value)} className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-white/10 backdrop-blur-md transition hover:bg-white/20 active:scale-95" aria-label={muted ? "Allow sound" : "Mute video"}>
            {muted ? <VolumeX size={19} /> : <Volume2 size={19} />}
          </button>
          <button type="button" onClick={() => setPlaying((value) => !value)} className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-white/10 backdrop-blur-md transition hover:bg-white/20 active:scale-95" aria-label={playing ? "Pause video" : "Play video"}>
            {playing ? <Pause size={18} /> : <Play size={18} className="ml-0.5" />}
          </button>
        </div>
        <div className="mx-auto mt-3 flex max-w-5xl items-center gap-2 text-[10px] font-medium tabular-nums text-white/70">
          <span>{formatTime(currentTime)}</span>
          <input type="range" min="0" max={duration || 0} step="0.01" value={Math.min(currentTime, duration || 0)} onChange={(event) => { const next = Number(event.target.value); setCurrentTime(next); if (videoRef.current) videoRef.current.currentTime = next; }} className="h-1 min-w-0 flex-1 cursor-pointer accent-white" aria-label="Video progress" />
          <span>{formatTime(duration)}</span>
        </div>
      </div>

      <div className="flex min-h-0 flex-1 items-center justify-center px-2 py-20 sm:px-6">
        {loading && <div className="text-sm text-white/50">Loading video…</div>}
        {!loading && error && <div className="rounded-2xl border border-white/10 bg-white/5 px-5 py-4 text-center text-sm text-white/70">{error}</div>}
        {!loading && !error && mediaUrl && <video ref={videoRef} src={mediaUrl} autoPlay muted={muted} playsInline preload="auto" className="max-h-full max-w-full rounded-xl object-contain shadow-2xl" onLoadedMetadata={(event) => setDuration(event.currentTarget.duration || 0)} onTimeUpdate={(event) => setCurrentTime(event.currentTarget.currentTime)} onPlay={() => setPlaying(true)} onPause={() => setPlaying(false)} onEnded={() => setPlaying(false)} onClick={() => setPlaying((value) => !value)} aria-label={post?.caption || "Video"} />}
      </div>
    </main>
  );
}
