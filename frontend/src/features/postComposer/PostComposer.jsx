import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  ArrowLeft,
  Check,
  ChevronLeft,
  Contrast,
  Crop,
  Image as ImageIcon,
  Loader2,
  Pencil,
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
import { CROP_RATIOS, DEFAULT_EDITS, FILTERS, exportEditedImage, getMediaStyle } from "./mediaEditor";

const MAX_FILE_SIZE = 100 * 1024 * 1024;
const MAX_CAPTION_LENGTH = 2000;
const EDITOR_TABS = ["Adjust", "Filters", "Crop", "Transform"];
const ADJUSTMENTS = [
  { key: "brightness", label: "Brightness", icon: SunMedium, min: 70, max: 140 },
  { key: "contrast", label: "Contrast", icon: Contrast, min: 70, max: 140 },
  { key: "saturation", label: "Saturation", icon: SlidersHorizontal, min: 0, max: 160 },
];

const ACCEPTED_MEDIA = "image/jpeg,image/png,image/webp,video/mp4,video/quicktime";
const fileKind = (file) => (file?.type?.startsWith("video/") ? "video" : "image");
const editsChanged = (edits) =>
  edits.filter !== "original" ||
  edits.brightness !== 100 ||
  edits.contrast !== 100 ||
  edits.saturation !== 100 ||
  edits.rotation !== 0 ||
  edits.flipX ||
  edits.crop !== "original";

const pressable = "transition-[transform,background-color,opacity,box-shadow] duration-150 ease-out active:scale-[0.96] disabled:active:scale-100";
const editorButton = `${pressable} disabled:opacity-50`;

export const PostComposer = () => {
  const navigate = useNavigate();
  const inputRef = useRef(null);
  const editSnapshotRef = useRef(DEFAULT_EDITS);
  const { createPost, loading, error } = usePostActions();
  const [step, setStep] = useState("select");
  const [file, setFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState("");
  const [kind, setKind] = useState(null);
  const [caption, setCaption] = useState("");
  const [accept, setAccept] = useState(ACCEPTED_MEDIA);
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
      setLocalError("File size must be below 100MB.");
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
    editSnapshotRef.current = DEFAULT_EDITS;
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
    editSnapshotRef.current = DEFAULT_EDITS;
    setCaption("");
    setAccept(ACCEPTED_MEDIA);
    setLocalError("");
    setStep("select");
  };

  const openEditor = () => {
    if (!file || uploading || loading) return;
    editSnapshotRef.current = edits;
    setEditorTab("Adjust");
    setLocalError("");
    setStep("edit");
  };

  const cancelEditor = () => {
    setEdits(editSnapshotRef.current);
    setLocalError("");
    setStep("preview");
  };

  const updateEdit = (key, value) => setEdits((current) => ({ ...current, [key]: value }));

  const autoFix = () =>
    setEdits((current) => ({
      ...current,
      filter: current.filter === "original" ? "vivid" : current.filter,
      brightness: 106,
      contrast: 108,
      saturation: 112,
    }));

  const resetEdits = () => setEdits(DEFAULT_EDITS);

  const applyEdits = async () => {
    if (!file || editing) return;
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
        editSnapshotRef.current = DEFAULT_EDITS;
      } catch (editorError) {
        setLocalError(editorError.message || "Could not apply image edits.");
        return;
      } finally {
        setEditing(false);
      }
    } else {
      editSnapshotRef.current = edits;
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
  const hasVideoPreviewEdits = kind === "video" && editsChanged(edits);
  const imageHasEdits = kind === "image" && editsChanged(edits);

  if (step === "edit") {
    return (
      <main className="min-h-[calc(100vh-64px)] bg-black text-white">
        <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-white/10 bg-black/95 px-3 backdrop-blur-xl sm:h-16 sm:px-4">
          <button type="button" onClick={cancelEditor} disabled={editing} className={`flex h-10 items-center gap-1 rounded-full px-2 text-sm font-semibold text-white/80 hover:bg-white/10 hover:text-white ${editorButton}`}>
            <ChevronLeft size={20} /> <span>Cancel</span>
          </button>
          <h1 className="text-[15px] font-bold sm:text-[16px]">Edit {kind === "video" ? "video" : "photo"}</h1>
          <button type="button" onClick={applyEdits} disabled={editing} className={`flex h-9 items-center gap-1.5 rounded-full bg-white px-4 text-sm font-bold text-black shadow-sm hover:bg-white/90 ${editorButton}`}>
            {editing ? <Loader2 size={15} className="animate-spin" /> : <Check size={15} />} Done
          </button>
        </header>

        {displayError && <div role="alert" className="mx-auto max-w-3xl px-4 pt-3 text-sm font-medium text-red-300">{displayError}</div>}

        <section className="mx-auto flex min-h-[calc(100vh-56px)] w-full max-w-3xl flex-col px-0 pb-3 sm:min-h-[calc(100vh-64px)] sm:px-3 sm:pb-4">
          <div className="relative flex min-h-0 flex-1 items-center justify-center overflow-hidden bg-black sm:mt-3 sm:rounded-[28px] sm:ring-1 sm:ring-white/10">
            {kind === "video" ? (
              <video src={previewUrl} controls playsInline preload="metadata" className="max-h-[72vh] max-w-full object-contain sm:max-h-[64vh]" style={mediaStyle} />
            ) : (
              <img src={previewUrl} alt="Editing preview" className="max-h-[72vh] max-w-full object-contain sm:max-h-[64vh]" style={mediaStyle} />
            )}
            <button type="button" onClick={autoFix} disabled={editing} aria-label="Auto fix" className={`absolute left-3 top-3 flex items-center gap-2 rounded-full bg-black/65 px-3 py-2 text-xs font-bold text-white backdrop-blur hover:bg-black/85 ${editorButton}`}><Sparkles size={14} /> Auto Fix</button>
          </div>

          <div className="border-t border-white/10 bg-black px-3 pb-[max(12px,env(safe-area-inset-bottom))] pt-2 sm:mt-3 sm:rounded-[24px] sm:border sm:bg-neutral-900 sm:p-3">
            <div className="mb-2 flex gap-1 overflow-x-auto border-b border-white/10 pb-2 sm:mb-3 sm:border-0 sm:pb-0" role="tablist" aria-label="Editor tools">
              {EDITOR_TABS.map((tab) => {
                const videoDisabled = kind === "video" && (tab === "Crop" || tab === "Transform");
                return (
                  <button key={tab} type="button" onClick={() => !videoDisabled && setEditorTab(tab)} disabled={videoDisabled} aria-pressed={editorTab === tab} className={`shrink-0 rounded-full px-4 py-2 text-xs font-bold ${pressable} ${editorTab === tab ? "bg-white text-black" : "text-white/55 hover:bg-white/10 hover:text-white"} ${videoDisabled ? "cursor-not-allowed opacity-30" : ""}`}>
                    {tab}
                  </button>
                );
              })}
            </div>

            {editorTab === "Adjust" && (
              <div className="grid gap-2 sm:grid-cols-3 sm:gap-3">
                {ADJUSTMENTS.map(({ key, label, icon: Icon, min, max }) => (
                  <label key={key} className={`rounded-2xl bg-white/[0.04] p-3 transition-colors ${edits[key] !== 100 ? "ring-1 ring-white/15" : ""}`}>
                    <span className="flex items-center justify-between text-xs font-semibold text-white/70"><span className="flex items-center gap-2"><Icon size={14} /> {label}</span><span className="tabular-nums text-white">{edits[key]}</span></span>
                    <input aria-label={`${label} value`} type="range" min={min} max={max} value={edits[key]} onChange={(event) => updateEdit(key, Number(event.target.value))} className="mt-3 w-full accent-white" disabled={editing} />
                  </label>
                ))}
                {editsChanged(edits) && <button type="button" onClick={resetEdits} disabled={editing} className={`justify-self-start rounded-full px-3 py-1.5 text-[11px] font-bold text-white/60 hover:bg-white/10 hover:text-white ${editorButton}`}>Reset adjustments</button>}
              </div>
            )}

            {editorTab === "Filters" && (
              <div className="flex gap-2 overflow-x-auto pb-1">
                {FILTERS.map((filter) => {
                  const selected = edits.filter === filter.id;
                  return (
                    <button key={filter.id} type="button" onClick={() => updateEdit("filter", filter.id)} disabled={editing} aria-pressed={selected} className={`group w-[72px] shrink-0 rounded-xl p-1 ${pressable} ${selected ? "bg-white shadow-[0_0_0_2px_rgba(255,255,255,0.22)]" : "bg-white/[0.04]"}`}>
                      <div className={`relative aspect-square overflow-hidden rounded-lg bg-neutral-800 transition-transform duration-200 ${selected ? "scale-[0.96]" : "group-hover:scale-[0.98]"}`}>
                        {kind === "video" ? (
                          <video src={previewUrl} muted playsInline preload="none" aria-hidden="true" className="h-full w-full object-cover" style={{ filter: filter.css }} />
                        ) : (
                          <img src={previewUrl} alt="" className="h-full w-full object-cover" style={{ filter: filter.css }} />
                        )}
                        {selected && <span className="absolute right-1.5 top-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-white text-black shadow-sm"><Check size={12} strokeWidth={3} /></span>}
                      </div>
                      <span className={`mt-1 block truncate text-[10px] font-semibold ${selected ? "text-black" : "text-white/65"}`}>{filter.label}</span>
                    </button>
                  );
                })}
              </div>
            )}

            {editorTab === "Crop" && (
              <div className="flex gap-2 overflow-x-auto pb-1">
                {CROP_RATIOS.map((ratio) => {
                  const selected = edits.crop === ratio.id;
                  return (
                    <button key={ratio.id} type="button" onClick={() => updateEdit("crop", ratio.id)} disabled={editing || kind === "video"} aria-pressed={selected} className={`relative flex min-w-[78px] flex-col items-center gap-2 rounded-2xl border px-3 py-3 text-xs font-bold ${pressable} ${selected ? "border-white bg-white text-black shadow-sm" : "border-white/10 bg-white/[0.04] text-white/70 hover:bg-white/10"} ${kind === "video" ? "opacity-45" : ""}`}>
                      <Crop size={17} />
                      {ratio.label}
                      {selected && <Check size={12} className="absolute right-2 top-2" strokeWidth={3} />}
                    </button>
                  );
                })}
              </div>
            )}

            {editorTab === "Transform" && (
              <div className="flex flex-wrap gap-2">
                {[
                  ["Rotate left", () => updateEdit("rotation", (edits.rotation + 270) % 360), RotateCcw],
                  ["Rotate right", () => updateEdit("rotation", (edits.rotation + 90) % 360), RotateCw],
                  ["Flip", () => updateEdit("flipX", !edits.flipX), RotateCw],
                  ["Reset", resetEdits, RotateCcw],
                ].map(([label, action, Icon]) => <button key={label} type="button" onClick={action} disabled={editing || kind === "video"} className={`flex items-center gap-2 rounded-2xl border border-white/10 bg-white/[0.04] px-4 py-3 text-xs font-bold text-white/80 hover:bg-white/10 hover:text-white ${editorButton}`}><Icon size={15} /> {label}</button>)}
              </div>
            )}
          </div>

          {kind === "video" && <p className="px-3 pt-2 text-center text-[10px] text-amber-200/70 sm:text-[11px]">Video edits are preview-only for now. Your original video is uploaded unchanged.</p>}
          {kind === "image" && <p className="px-3 pt-2 text-center text-[10px] text-white/35 sm:text-[11px]">Edits are rendered locally only when you press Done. Cancel keeps the current media unchanged.</p>}
        </section>
      </main>
    );
  }

  if (step === "preview") {
    return (
      <main className="min-h-[calc(100vh-64px)] bg-slate-50/80 pb-10">
        <section className="mx-auto w-full max-w-[760px] px-3 sm:px-5">
          <header className="sticky top-0 z-20 -mx-3 flex h-16 items-center justify-between border-b border-slate-200/80 bg-white/90 px-4 backdrop-blur-xl sm:-mx-5 sm:px-6">
            <button type="button" onClick={() => { setStep("select"); setAccept(ACCEPTED_MEDIA); }} disabled={uploading || loading} className={`flex h-10 items-center gap-1 rounded-full px-2 text-sm font-semibold text-slate-700 hover:bg-slate-100 ${pressable}`}><ArrowLeft size={19} /> Change</button>
            <div className="text-center"><h1 className="text-[17px] font-bold text-slate-950">Preview</h1><p className="hidden text-[11px] text-slate-400 sm:block">Looks good? Share it.</p></div>
            <button type="button" onClick={handleShare} disabled={uploading || loading} className={`flex h-9 items-center gap-1.5 rounded-full bg-blue-600 px-4 text-[14px] font-bold text-white shadow-sm hover:bg-blue-700 ${pressable} disabled:opacity-45`}>{uploading || loading ? <Loader2 size={15} className="animate-spin" /> : <Check size={15} />}{uploading ? "Uploading" : loading ? "Posting" : "Share"}</button>
          </header>

          {displayError && <div role="alert" className="mx-1 mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700">{displayError}</div>}

          <div className="mt-4 overflow-hidden rounded-[28px] bg-neutral-950 shadow-xl">
            <div className="relative flex aspect-[4/3] items-center justify-center overflow-hidden bg-black">
              {kind === "video" ? <video src={previewUrl} controls playsInline preload="metadata" className="h-full w-full object-contain" style={mediaStyle} /> : <img src={previewUrl} alt="Post preview" className="h-full w-full object-contain" />}
              <button type="button" onClick={openEditor} disabled={uploading || loading} className={`absolute bottom-4 left-4 flex h-11 w-11 items-center justify-center rounded-full bg-black/70 text-white shadow-lg ring-1 ring-white/15 backdrop-blur hover:bg-black/85 ${pressable} disabled:opacity-50`} aria-label="Edit media" title="Edit media"><Pencil size={17} /></button>
              <button type="button" onClick={reset} disabled={uploading || loading} className={`absolute right-4 top-4 flex h-10 w-10 items-center justify-center rounded-full bg-black/65 text-white backdrop-blur hover:bg-black/85 ${pressable} disabled:opacity-50`} aria-label="Remove media"><X size={18} /></button>
              <span className="absolute right-4 bottom-4 flex items-center gap-1.5 rounded-full bg-black/60 px-3 py-2 text-[11px] font-bold text-white backdrop-blur">{kind === "video" ? <Video size={13} /> : <ImageIcon size={13} />}{kind === "video" ? "Video" : "Photo"}</span>
            </div>
            <div className="border-t border-white/10 bg-neutral-950 p-4"><label htmlFor="post-caption" className="text-xs font-bold text-white/55">Caption</label><textarea id="post-caption" value={caption} onChange={(event) => setCaption(event.target.value)} maxLength={MAX_CAPTION_LENGTH} rows={3} placeholder="Tell your community what this moment is about..." className="mt-2 w-full resize-none bg-transparent text-sm leading-6 text-white outline-none placeholder:text-white/30" /><div className="mt-1 text-right text-[10px] text-white/30">{caption.length}/{MAX_CAPTION_LENGTH}</div></div>
          </div>
          {(hasVideoPreviewEdits || imageHasEdits) && <p className="mt-2 text-center text-[10px] text-slate-400">{hasVideoPreviewEdits ? "Video adjustments are preview-only and will not alter the uploaded source." : "Edited photo ready to share."}</p>}
        </section>
      </main>
    );
  }

  return (
    <main className="min-h-[calc(100vh-64px)] bg-slate-50/80 pb-10">
      <section className="mx-auto w-full max-w-[760px] px-3 sm:px-5">
        <header className="sticky top-0 z-20 -mx-3 flex h-16 items-center justify-between border-b border-slate-200/80 bg-white/90 px-4 backdrop-blur-xl sm:-mx-5 sm:px-6">
          <button type="button" onClick={() => navigate(-1)} className={`flex h-10 w-10 items-center justify-start rounded-full text-slate-900 hover:bg-slate-100 ${pressable}`} aria-label="Go back"><ArrowLeft size={21} /></button>
          <div className="text-center"><h1 className="text-[17px] font-bold tracking-tight text-slate-950">Create post</h1><p className="hidden text-[11px] text-slate-400 sm:block">Choose media to get started</p></div>
          <span className="w-10" />
        </header>

        {displayError && <div role="alert" className="mx-1 mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700">{displayError}</div>}
        <input ref={inputRef} type="file" accept={accept} onChange={handleInput} className="hidden" />

        <div onDragOver={(event) => { event.preventDefault(); setDragActive(true); }} onDragLeave={() => setDragActive(false)} onDrop={(event) => { event.preventDefault(); setDragActive(false); chooseFile(event.dataTransfer.files?.[0]); }} className={`relative mt-4 flex min-h-[500px] w-full flex-col items-center justify-center overflow-hidden rounded-[30px] border px-6 text-center transition ${dragActive ? "border-blue-500 bg-blue-50" : "border-slate-200 bg-white shadow-sm"}`}>
          <div className="absolute left-5 top-5 flex items-center gap-2 rounded-full border border-slate-200 bg-white/90 px-3 py-1.5 text-xs font-semibold text-slate-600 shadow-sm"><Sparkles size={14} className="text-blue-600" /> Media studio</div>
          <div className="flex h-20 w-20 items-center justify-center rounded-[24px] bg-blue-50 text-blue-600 ring-8 ring-blue-50/60"><Plus size={38} strokeWidth={1.8} /></div>
          <h2 className="mt-7 text-[21px] font-bold tracking-tight text-slate-950">Pick a photo or video</h2>
          <p className="mt-2 max-w-sm text-sm leading-5 text-slate-500">Select media first. You will preview it, then optionally edit it, then share.</p>
          <button type="button" onClick={() => { setAccept(ACCEPTED_MEDIA); inputRef.current?.click(); }} className={`mt-6 rounded-full bg-slate-950 px-6 py-3 text-sm font-bold text-white shadow-sm hover:bg-slate-800 ${pressable}`}>Browse device</button>
          <div className="mt-5 flex flex-wrap justify-center gap-2 text-[11px] font-medium text-slate-400"><span>JPG</span><span>•</span><span>PNG</span><span>•</span><span>WEBP</span><span>•</span><span>MP4</span><span>•</span><span>MOV</span><span>•</span><span>Up to 100MB</span></div>
        </div>

        <div className="mt-3 grid grid-cols-2 gap-3">
          <button type="button" onClick={() => { setAccept("image/*"); inputRef.current?.click(); }} className={`flex h-12 items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white text-sm font-bold text-slate-800 shadow-sm hover:bg-slate-50 ${pressable}`}><ImageIcon size={18} className="text-blue-600" /> Photo</button>
          <button type="button" onClick={() => { setAccept("video/*"); inputRef.current?.click(); }} className={`flex h-12 items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white text-sm font-bold text-slate-800 shadow-sm hover:bg-slate-50 ${pressable}`}><Video size={18} className="text-blue-600" /> Video</button>
        </div>
        <p className="pb-2 pt-4 text-center text-[11px] leading-5 text-slate-400">Nothing is uploaded until you press Share after the preview step.</p>
      </section>
    </main>
  );
};
