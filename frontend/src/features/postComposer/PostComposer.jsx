import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  ArrowLeft,
  Check,
  ChevronLeft,
  Contrast,
  Download,
  Image as ImageIcon,
  Loader2,
  Plus,
  RotateCcw,
  RotateCw,
  SlidersHorizontal,
  Sparkles,
  SunMedium,
  Video,
  X,
} from "lucide-react";
import { usePostActions } from "../../hooks/usePosts";
import { uploadAPI } from "../../services/post/UploadApi";
import { DEFAULT_EDITS, FILTERS, exportEditedImage, getMediaStyle } from "./mediaEditor";

const MAX_FILE_SIZE = 50 * 1024 * 1024;
const EDITOR_TABS = ["Adjust", "Filters", "Transform"];
const ADJUSTMENTS = [
  { key: "brightness", label: "Brightness", icon: SunMedium, min: 70, max: 140 },
  { key: "contrast", label: "Contrast", icon: Contrast, min: 70, max: 140 },
  { key: "saturation", label: "Saturation", icon: SlidersHorizontal, min: 0, max: 160 },
];

const fileKind = (file) => (file?.type?.startsWith("video/") ? "video" : "image");
const editsChanged = (edits) => edits.filter !== "original" || edits.brightness !== 100 || edits.contrast !== 100 || edits.saturation !== 100 || edits.rotation !== 0 || edits.flipX;

export const PostComposer = () => {
  const navigate = useNavigate();
  const inputRef = useRef(null);
  const { createPost, loading, error } = usePostActions();
  const [step, setStep] = useState("select");
  const [file, setFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState("");
  const [kind, setKind] = useState(null);
  const [caption, setCaption] = useState("");
  const [accept, setAccept] = useState("image/jpeg,image/png,image/webp,video/mp4,video/quicktime");
  const [dragActive, setDragActive] = useState(false);
  const [localError, setLocalError] = useState("");
  const [uploading, setUploading] = useState(false);
  const [editorTab, setEditorTab] = useState("Adjust");
  const [edits, setEdits] = useState(DEFAULT_EDITS);
  const [editing, setEditing] = useState(false);

  useEffect(() => () => {
    if (previewUrl?.startsWith("blob:")) URL.revokeObjectURL(previewUrl);
  }, [previewUrl]);

  const chooseFile = useCallback((nextFile) => {
    if (!nextFile?.type?.startsWith("image/") && !nextFile?.type?.startsWith("video/")) {
      setLocalError("Choose an image or video file.");
      return;
    }
    if (nextFile.size > MAX_FILE_SIZE) {
      setLocalError("File size must be below 50MB.");
      return;
    }
    const nextUrl = URL.createObjectURL(nextFile);
    setPreviewUrl((current) => {
      if (current?.startsWith("blob:")) URL.revokeObjectURL(current);
      return nextUrl;
    });
    setFile(nextFile);
    setKind(fileKind(nextFile));
    setEdits(DEFAULT_EDITS);
    setStep("preview");
    setLocalError("");
  }, []);

  const handleInput = (event) => {
    const nextFile = event.target.files?.[0];
    if (nextFile) chooseFile(nextFile);
    event.target.value = "";
  };

  const reset = () => {
    setPreviewUrl((current) => {
      if (current?.startsWith("blob:")) URL.revokeObjectURL(current);
      return "";
    });
    setFile(null);
    setKind(null);
    setEdits(DEFAULT_EDITS);
    setCaption("");
    setAccept("image/jpeg,image/png,image/webp,video/mp4,video/quicktime");
    setLocalError("");
    setStep("select");
  };

  const updateEdit = (key, value) => setEdits((current) => ({ ...current, [key]: value }));

  const autoFix = () => setEdits((current) => ({
    ...current,
    filter: current.filter === "original" ? "vivid" : current.filter,
    brightness: 106,
    contrast: 108,
    saturation: 112,
  }));

  const applyEdits = async () => {
    if (!file) return;
    setLocalError("");
    if (kind === "image" && editsChanged(edits)) {
      setEditing(true);
      try {
        const editedFile = await exportEditedImage(file, edits);
        const editedUrl = URL.createObjectURL(editedFile);
        setPreviewUrl((current) => {
          if (current?.startsWith("blob:")) URL.revokeObjectURL(current);
          return editedUrl;
        });
        setFile(editedFile);
        setEdits(DEFAULT_EDITS);
      } catch (editorError) {
        setLocalError(editorError.message || "Could not apply image edits.");
        return;
      } finally {
        setEditing(false);
      }
    }
    setStep("preview");
  };

  const handleShare = async () => {
    if (!file || uploading || loading) return;
    setLocalError("");
    try {
      setUploading(true);
      const response = await uploadAPI.uploadMedia(file);
      if (!response?.url) throw new Error("Media upload failed.");
      const mediaUrl = response.url.startsWith("http")
        ? response.url
        : `${import.meta.env.VITE_SERVER_URL ?? "http://localhost:8080"}${response.url}`;
      await createPost({ contentType: kind, contentUrl: mediaUrl, caption: caption.trim() });
      navigate("/");
    } catch (shareError) {
      setLocalError(shareError.message || "Failed to publish post.");
    } finally {
      setUploading(false);
    }
  };

  const displayError = localError || error;
  const mediaStyle = getMediaStyle(edits);

  if (step === "edit") {
    return (
      <main className="min-h-[calc(100vh-64px)] bg-neutral-950 text-white">
        <header className="sticky top-0 z-20 flex h-16 items-center justify-between border-b border-white/10 bg-neutral-950/95 px-4 backdrop-blur-xl">
          <button type="button" onClick={() => setStep("preview")} className="flex h-10 items-center gap-2 rounded-full px-2 text-sm font-semibold text-white/80 hover:bg-white/10 hover:text-white"><ChevronLeft size={20} /> Back</button>
          <h1 className="text-[16px] font-bold">Edit {kind === "video" ? "video" : "photo"}</h1>
          <button type="button" onClick={applyEdits} disabled={editing} className="flex h-9 items-center gap-1.5 rounded-full bg-white px-4 text-sm font-bold text-neutral-950 disabled:opacity-50">{editing ? <Loader2 size={15} className="animate-spin" /> : <Check size={15} />} Apply</button>
        </header>

        {displayError && <div role="alert" className="mx-auto mt-3 max-w-3xl px-4 text-sm font-medium text-red-300">{displayError}</div>}

        <section className="mx-auto flex min-h-[calc(100vh-128px)] w-full max-w-3xl flex-col px-3 pb-4">
          <div className="relative mt-3 flex min-h-0 flex-1 items-center justify-center overflow-hidden rounded-[28px] bg-black ring-1 ring-white/10">
            {kind === "video" ? (
              <video src={previewUrl} controls playsInline preload="metadata" className="max-h-[64vh] max-w-full object-contain" style={mediaStyle} />
            ) : (
              <img src={previewUrl} alt="Editing preview" className="max-h-[64vh] max-w-full object-contain" style={mediaStyle} />
            )}
            <button type="button" onClick={autoFix} className="absolute left-3 top-3 flex items-center gap-2 rounded-full bg-black/60 px-3 py-2 text-xs font-bold text-white backdrop-blur hover:bg-black/80"><Sparkles size={14} /> Auto Fix</button>
          </div>

          <div className="mt-3 rounded-[24px] border border-white/10 bg-neutral-900 p-3">
            <div className="mb-3 flex items-center gap-1 overflow-x-auto">
              {EDITOR_TABS.map((tab) => <button key={tab} type="button" onClick={() => setEditorTab(tab)} className={`shrink-0 rounded-full px-4 py-2 text-xs font-bold ${editorTab === tab ? "bg-white text-neutral-950" : "text-white/55 hover:bg-white/10 hover:text-white"}`}>{tab}</button>)}
            </div>

            {editorTab === "Adjust" && <div className="grid gap-3 sm:grid-cols-3">{ADJUSTMENTS.map(({ key, label, icon: Icon, min, max }) => <label key={key} className="rounded-2xl bg-black/20 p-3"><span className="flex items-center justify-between text-xs font-semibold text-white/70"><span className="flex items-center gap-2"><Icon size={14} /> {label}</span><span>{edits[key]}</span></span><input type="range" min={min} max={max} value={edits[key]} onChange={(event) => updateEdit(key, Number(event.target.value))} className="mt-3 w-full accent-white" /></label>)}</div>}

            {editorTab === "Filters" && <div className="flex gap-2 overflow-x-auto pb-1">{FILTERS.map((filter) => <button key={filter.id} type="button" onClick={() => updateEdit("filter", filter.id)} className={`group w-[72px] shrink-0 rounded-xl p-1 ${edits.filter === filter.id ? "bg-white" : "bg-black/30"}`}><div className="aspect-square overflow-hidden rounded-lg bg-neutral-800">{kind === "video" ? <video src={previewUrl} muted playsInline className="h-full w-full object-cover" style={{ filter: filter.css }} /> : <img src={previewUrl} alt="" className="h-full w-full object-cover" style={{ filter: filter.css }} />}</div><span className={`mt-1 block truncate text-[10px] font-semibold ${edits.filter === filter.id ? "text-neutral-950" : "text-white/65"}`}>{filter.label}</span></button>)}</div>}

            {editorTab === "Transform" && <div className="flex flex-wrap gap-2">{[
              ["Rotate left", () => updateEdit("rotation", (edits.rotation + 270) % 360), RotateCcw],
              ["Rotate right", () => updateEdit("rotation", (edits.rotation + 90) % 360), RotateCw],
              ["Flip", () => updateEdit("flipX", !edits.flipX), RotateCw],
              ["Reset", () => setEdits(DEFAULT_EDITS), RotateCcw],
            ].map(([label, action, Icon]) => <button key={label} type="button" onClick={action} className="flex items-center gap-2 rounded-2xl border border-white/10 bg-black/20 px-4 py-3 text-xs font-bold text-white/80 hover:bg-white/10 hover:text-white"><Icon size={15} /> {label}</button>)}</div>}
          </div>

          <p className="px-2 pt-2 text-center text-[11px] text-white/35">Photo edits are rendered into the upload. Video edits are preview controls and remain non-destructive.</p>
        </section>
      </main>
    );
  }

  if (step === "preview") {
    return (
      <main className="min-h-[calc(100vh-64px)] bg-slate-50/80 pb-10">
        <section className="mx-auto w-full max-w-[760px] px-3 sm:px-5">
          <header className="sticky top-0 z-20 -mx-3 flex h-16 items-center justify-between border-b border-slate-200/80 bg-white/90 px-4 backdrop-blur-xl sm:-mx-5 sm:px-6">
            <button type="button" onClick={() => { setStep("select"); setAccept("image/jpeg,image/png,image/webp,video/mp4,video/quicktime"); }} disabled={uploading || loading} className="flex h-10 items-center gap-1 rounded-full px-2 text-sm font-semibold text-slate-700 hover:bg-slate-100"><ArrowLeft size={19} /> Change</button>
            <div className="text-center"><h1 className="text-[17px] font-bold text-slate-950">Preview</h1><p className="hidden text-[11px] text-slate-400 sm:block">Looks good? Share it.</p></div>
            <button type="button" onClick={handleShare} disabled={uploading || loading} className="flex h-9 items-center gap-1.5 rounded-full bg-blue-600 px-4 text-[14px] font-bold text-white shadow-sm hover:bg-blue-700 disabled:opacity-45">{uploading || loading ? <Loader2 size={15} className="animate-spin" /> : <Check size={15} />}{uploading ? "Uploading" : loading ? "Posting" : "Share"}</button>
          </header>

          {displayError && <div role="alert" className="mx-1 mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700">{displayError}</div>}

          <div className="mt-4 overflow-hidden rounded-[28px] bg-neutral-950 shadow-xl">
            <div className="relative flex aspect-[4/3] items-center justify-center overflow-hidden bg-black">
              {kind === "video" ? <video src={previewUrl} controls playsInline preload="metadata" className="h-full w-full object-contain" style={mediaStyle} /> : <img src={previewUrl} alt="Post preview" className="h-full w-full object-contain" />}
              <button type="button" onClick={() => setStep("edit")} className="absolute bottom-4 left-4 flex items-center gap-2 rounded-full bg-black/65 px-4 py-2.5 text-sm font-bold text-white backdrop-blur hover:bg-black/85"><SlidersHorizontal size={16} /> Edit</button>
              <button type="button" onClick={reset} className="absolute right-4 top-4 flex h-10 w-10 items-center justify-center rounded-full bg-black/65 text-white backdrop-blur hover:bg-black/85" aria-label="Remove media"><X size={18} /></button>
              <span className="absolute right-4 bottom-4 flex items-center gap-1.5 rounded-full bg-black/60 px-3 py-2 text-[11px] font-bold text-white backdrop-blur">{kind === "video" ? <Video size={13} /> : <ImageIcon size={13} />}{kind === "video" ? "Video" : "Photo"}</span>
            </div>
            <div className="border-t border-white/10 bg-neutral-950 p-4"><label htmlFor="post-caption" className="text-xs font-bold text-white/55">Caption</label><textarea id="post-caption" value={caption} onChange={(event) => setCaption(event.target.value)} maxLength={2200} rows={3} placeholder="Tell your community what this moment is about..." className="mt-2 w-full resize-none bg-transparent text-sm leading-6 text-white outline-none placeholder:text-white/30" /><div className="mt-1 text-right text-[10px] text-white/30">{caption.length}/2200</div></div>
          </div>
          <button type="button" onClick={() => setStep("edit")} className="mt-3 flex w-full items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white py-3 text-sm font-bold text-slate-800 shadow-sm hover:bg-slate-50"><SlidersHorizontal size={17} /> Edit media</button>
        </section>
      </main>
    );
  }

  return (
    <main className="min-h-[calc(100vh-64px)] bg-slate-50/80 pb-10">
      <section className="mx-auto w-full max-w-[760px] px-3 sm:px-5">
        <header className="sticky top-0 z-20 -mx-3 flex h-16 items-center justify-between border-b border-slate-200/80 bg-white/90 px-4 backdrop-blur-xl sm:-mx-5 sm:px-6">
          <button type="button" onClick={() => navigate(-1)} className="flex h-10 w-10 items-center justify-start rounded-full text-slate-900 hover:bg-slate-100" aria-label="Go back"><ArrowLeft size={21} /></button>
          <div className="text-center"><h1 className="text-[17px] font-bold tracking-tight text-slate-950">Create post</h1><p className="hidden text-[11px] text-slate-400 sm:block">Choose media to get started</p></div>
          <span className="w-10" />
        </header>

        {displayError && <div role="alert" className="mx-1 mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700">{displayError}</div>}
        <input ref={inputRef} type="file" accept={accept} onChange={handleInput} className="hidden" />

        <div onDragOver={(event) => { event.preventDefault(); setDragActive(true); }} onDragLeave={() => setDragActive(false)} onDrop={(event) => { event.preventDefault(); setDragActive(false); chooseFile(event.dataTransfer.files?.[0]); }} className={`relative mt-4 flex min-h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-[30px] border px-6 text-center transition ${dragActive ? "border-blue-500 bg-blue-50" : "border-slate-200 bg-white shadow-sm"}`}>
          <div className="absolute left-5 top-5 flex items-center gap-2 rounded-full border border-slate-200 bg-white/90 px-3 py-1.5 text-xs font-semibold text-slate-600 shadow-sm"><Sparkles size={14} className="text-blue-600" /> Media studio</div>
          <div className="flex h-20 w-20 items-center justify-center rounded-[24px] bg-blue-50 text-blue-600 ring-8 ring-blue-50/60"><Plus size={38} strokeWidth={1.8} /></div>
          <h2 className="mt-7 text-[21px] font-bold tracking-tight text-slate-950">Pick a photo or video</h2>
          <p className="mt-2 max-w-sm text-sm leading-5 text-slate-500">Select media first. You will preview it, edit it, then share.</p>
          <button type="button" onClick={() => { setAccept("image/jpeg,image/png,image/webp,video/mp4,video/quicktime"); inputRef.current?.click(); }} className="mt-6 rounded-full bg-slate-950 px-6 py-3 text-sm font-bold text-white shadow-sm hover:bg-slate-800">Browse device</button>
          <div className="mt-5 flex flex-wrap justify-center gap-2 text-[11px] font-medium text-slate-400"><span>JPG</span><span>•</span><span>PNG</span><span>•</span><span>WEBP</span><span>•</span><span>MP4</span><span>•</span><span>MOV</span><span>•</span><span>Up to 50MB</span></div>
        </div>

        <div className="mt-3 grid grid-cols-2 gap-3">
          <button type="button" onClick={() => { setAccept("image/*"); inputRef.current?.click(); }} className="flex h-12 items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white text-sm font-bold text-slate-800 shadow-sm hover:bg-slate-50"><ImageIcon size={18} className="text-blue-600" /> Photo</button>
          <button type="button" onClick={() => { setAccept("video/*"); inputRef.current?.click(); }} className="flex h-12 items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white text-sm font-bold text-slate-800 shadow-sm hover:bg-slate-50"><Video size={18} className="text-blue-600" /> Video</button>
        </div>
        <p className="pb-2 pt-4 text-center text-[11px] leading-5 text-slate-400">Nothing is uploaded until you press Share after the preview/edit steps.</p>
      </section>
    </main>
  );
};
