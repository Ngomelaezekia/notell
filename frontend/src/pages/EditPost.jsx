import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Check, Loader2, Scissors } from "lucide-react";
import { postsAPI } from "../services/post/postsApi";
import { uploadAPI } from "../services/post/UploadApi";
import { getFileUrl } from "../utils/api";
import { CROP_RATIOS, DEFAULT_EDITS, FILTERS, exportEditedImage, getMediaStyle, trimVideo } from "../features/postComposer/mediaEditor";

const ACCEPTED = "image/jpeg,image/png,image/webp,video/mp4,video/quicktime,video/webm";
const MAX = 100 * 1024 * 1024;

export const EditPost = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const inputRef = useRef(null);
  const [post, setPost] = useState(null);
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState("");
  const [caption, setCaption] = useState("");
  const [visibility, setVisibility] = useState("public");
  const [edits, setEdits] = useState(DEFAULT_EDITS);
  const [trim, setTrim] = useState({ duration: 0, start: 0, end: 0 });
  const [dirty, setDirty] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const kind = post?.contentType || (file?.type?.startsWith("video/") ? "video" : "image");
  const mediaStyle = useMemo(() => getMediaStyle(edits), [edits]);

  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const response = await postsAPI.getById(id);
        const data = response?.data || response;
        if (!active) return;
        setPost(data);
        setCaption(data?.caption || "");
        setVisibility(data?.visibility || "public");
        const url = getFileUrl(data?.contentUrl);
        setPreview(url);
        const res = await fetch(url, { credentials: "include" });
        if (!res.ok) throw new Error("Could not load the post media.");
        const blob = await res.blob();
        if (!active) return;
        setFile(new File([blob], `post-${id}${data?.contentType === "video" ? ".mp4" : ".jpg"}`, { type: blob.type || (data?.contentType === "video" ? "video/mp4" : "image/jpeg") }));
      } catch (e) {
        if (active) setError(e?.message || "Could not load post.");
      }
    })();
    return () => { active = false; };
  }, [id]);

  useEffect(() => () => {
    if (preview?.startsWith("blob:")) URL.revokeObjectURL(preview);
  }, [preview]);

  const chooseFile = (next) => {
    if (!next || !ACCEPTED.split(",").includes(next.type)) {
      setError("Choose a supported image or video file.");
      return;
    }
    if (next.size > MAX) {
      setError("File size must be below 100MB.");
      return;
    }
    const url = URL.createObjectURL(next);
    setPreview((current) => { if (current?.startsWith("blob:")) URL.revokeObjectURL(current); return url; });
    setFile(next);
    setDirty(true);
    setError("");
    setTrim({ duration: 0, start: 0, end: 0 });
  };

  const applyImage = async () => {
    if (!file || kind !== "image") return;
    setSaving(true); setError("");
    try {
      const edited = await exportEditedImage(file, edits);
      const url = URL.createObjectURL(edited);
      setPreview((current) => { if (current?.startsWith("blob:")) URL.revokeObjectURL(current); return url; });
      setFile(edited); setDirty(true); setEdits(DEFAULT_EDITS);
    } catch (e) { setError(e?.message || "Could not apply image edits."); }
    finally { setSaving(false); }
  };

  const applyTrim = async () => {
    if (!file || kind !== "video" || !trim.duration || trim.end <= trim.start) return;
    setSaving(true); setError("");
    try {
      const edited = await trimVideo(file, trim.start, trim.end);
      const url = URL.createObjectURL(edited);
      setPreview((current) => { if (current?.startsWith("blob:")) URL.revokeObjectURL(current); return url; });
      setFile(edited); setDirty(true); setTrim({ duration: 0, start: 0, end: 0 });
    } catch (e) { setError(e?.message || "Could not trim video."); }
    finally { setSaving(false); }
  };

  const save = async () => {
    if (!post || saving) return;
    setSaving(true); setError("");
    try {
      let contentUrl;
      let contentType = post.contentType;
      if (dirty && file) {
        const uploaded = await uploadAPI.uploadMedia(file);
        if (!uploaded?.url) throw new Error("Replacement media upload failed.");
        contentUrl = uploaded.url.startsWith("http") ? uploaded.url : `${import.meta.env.VITE_SERVER_URL ?? "http://localhost:8080"}${uploaded.url}`;
        contentType = file.type.startsWith("video/") ? "video" : "image";
      }
      await postsAPI.update(id, {
        caption: caption.trim(),
        visibility,
        ...(contentUrl ? { contentUrl, contentType } : {}),
      });
      navigate(-1);
    } catch (e) {
      setError(e?.message || "Could not save post changes.");
    } finally { setSaving(false); }
  };

  return <main className="min-h-screen bg-black text-white">
    <header className="sticky top-0 z-20 flex h-14 items-center justify-between border-b border-white/10 bg-black/95 px-3 backdrop-blur-xl">
      <button type="button" onClick={() => navigate(-1)} className="flex items-center gap-1 rounded-full px-2 py-2 text-sm font-semibold text-white/75 hover:bg-white/10"><ArrowLeft size={20}/> Back</button>
      <h1 className="text-sm font-bold">Edit post</h1>
      <button type="button" onClick={save} disabled={!post || saving} className="flex items-center gap-1.5 rounded-full bg-white px-4 py-2 text-xs font-bold text-black disabled:opacity-50">{saving ? <Loader2 size={14} className="animate-spin"/> : <Check size={14}/>} Save</button>
    </header>
    {error && <div role="alert" className="mx-auto max-w-2xl px-4 pt-3 text-sm text-red-300">{error}</div>}
    <section className="mx-auto max-w-2xl space-y-4 p-4">
      <div className="flex min-h-[360px] items-center justify-center overflow-hidden rounded-3xl bg-neutral-950 ring-1 ring-white/10">
        {preview ? (kind === "video" ? <video src={preview} controls playsInline preload="metadata" onLoadedMetadata={e => setTrim(t => ({...t, duration:e.currentTarget.duration, end:t.end || e.currentTarget.duration}))} className="max-h-[65vh] max-w-full object-contain" style={mediaStyle}/> : <img src={preview} alt="Post preview" className="max-h-[65vh] max-w-full object-contain" style={mediaStyle}/>) : <Loader2 className="animate-spin"/>}
      </div>
      <div className="flex flex-wrap gap-2">
        <button type="button" onClick={() => inputRef.current?.click()} className="rounded-full bg-white/10 px-4 py-2 text-xs font-bold hover:bg-white/15">Replace media</button>
        {kind === "image" && <button type="button" onClick={applyImage} disabled={saving} className="rounded-full bg-white px-4 py-2 text-xs font-bold text-black disabled:opacity-50">Apply image edits</button>}
        {kind === "video" && trim.duration > 0 && <button type="button" onClick={applyTrim} disabled={saving || trim.end <= trim.start} className="flex items-center gap-1.5 rounded-full bg-white px-4 py-2 text-xs font-bold text-black disabled:opacity-50"><Scissors size={13}/> Apply trim</button>}
      </div>
      {kind === "image" && <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
        {["brightness","contrast","saturation"].map(key => <label key={key} className="rounded-2xl bg-white/[.05] p-3 text-xs text-white/70">{key}<input type="range" min={key === "saturation" ? 0 : 70} max="160" value={edits[key]} onChange={e => setEdits(x => ({...x,[key]:Number(e.target.value)}))} className="mt-2 w-full accent-white"/></label>)}
        <select value={edits.filter} onChange={e => setEdits(x => ({...x,filter:e.target.value}))} className="rounded-2xl bg-white/[.05] p-3 text-xs text-white"><option value="original">Original</option>{FILTERS.filter(x=>x.id!=="original").map(x=><option key={x.id} value={x.id}>{x.label}</option>)}</select>
        <select value={edits.crop} onChange={e => setEdits(x => ({...x,crop:e.target.value}))} className="rounded-2xl bg-white/[.05] p-3 text-xs text-white">{CROP_RATIOS.map(x=><option key={x.id} value={x.id}>{x.label}</option>)}</select>
      </div>}
      {kind === "video" && trim.duration > 0 && <div className="grid grid-cols-2 gap-3 rounded-2xl bg-white/[.05] p-3 text-xs"><label>Start {trim.start.toFixed(1)}s<input type="range" min="0" max={Math.max(.1,trim.duration-.1)} step=".1" value={trim.start} onChange={e=>setTrim(x=>({...x,start:Math.min(Number(e.target.value),x.end-.1)}))} className="mt-2 w-full accent-white"/></label><label>End {trim.end.toFixed(1)}s<input type="range" min=".1" max={trim.duration} step=".1" value={trim.end} onChange={e=>setTrim(x=>({...x,end:Math.max(Number(e.target.value),x.start+.1)}))} className="mt-2 w-full accent-white"/></label></div>}
      <textarea value={caption} onChange={e=>setCaption(e.target.value.slice(0,2000))} rows="4" placeholder="Write a caption..." className="w-full rounded-2xl border border-white/10 bg-white/[.04] p-4 text-sm outline-none placeholder:text-white/30"/>
      <select value={visibility} onChange={e=>setVisibility(e.target.value)} className="w-full rounded-2xl border border-white/10 bg-white/[.04] p-4 text-sm">
        <option value="public">Public</option><option value="followers">Followers</option><option value="subscribers">Subscribers</option><option value="selected">Selected users</option><option value="private">Private</option>
      </select>
      <p className="text-[11px] text-white/40">Changing to Selected users requires managing the audience separately. Subscriber visibility requires an active entitlement.</p>
    </section>
    <input ref={inputRef} type="file" accept={ACCEPTED} onChange={e => { chooseFile(e.target.files?.[0]); e.target.value=""; }} className="hidden"/>
  </main>;
};
