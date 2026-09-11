import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, Check, Image as ImageIcon, Loader2, Plus, RotateCcw, Sparkles, Video, X } from "lucide-react";
import { usePostActions } from "../hooks/usePosts";
import { uploadAPI } from "../services/post/UploadApi";

const MAX_FILE_SIZE = 50 * 1024 * 1024;

export const CreatePost = () => {
  const navigate = useNavigate();
  const fileInputRef = useRef(null);
  const photoInputRef = useRef(null);
  const videoInputRef = useRef(null);
  const { createPost, loading, error } = usePostActions();
  const [selectedFile, setSelectedFile] = useState(null);
  const [previewURL, setPreviewURL] = useState("");
  const [contentType, setContentType] = useState(null);
  const [caption, setCaption] = useState("");
  const [localError, setLocalError] = useState("");
  const [dragActive, setDragActive] = useState(false);
  const [uploading, setUploading] = useState(false);

  useEffect(() => () => {
    if (previewURL?.startsWith("blob:")) URL.revokeObjectURL(previewURL);
  }, [previewURL]);

  const validateFile = (file) => {
    if (!file?.type?.startsWith("image/") && !file?.type?.startsWith("video/")) throw new Error("Choose an image or video file.");
    if (file.size > MAX_FILE_SIZE) throw new Error("File size must be below 50MB.");
  };

  const processFile = useCallback((file) => {
    try {
      validateFile(file);
      setPreviewURL((current) => {
        if (current?.startsWith("blob:")) URL.revokeObjectURL(current);
        return URL.createObjectURL(file);
      });
      setSelectedFile(file);
      setContentType(file.type.startsWith("video/") ? "video" : "image");
      setLocalError("");
    } catch (fileError) {
      setLocalError(fileError.message);
    }
  }, []);

  const handleFileSelect = (event) => {
    const file = event.target.files?.[0];
    if (file) processFile(file);
    event.target.value = "";
  };

  const handleDrop = (event) => {
    event.preventDefault();
    setDragActive(false);
    const file = event.dataTransfer.files?.[0];
    if (file) processFile(file);
  };

  const clearMedia = () => {
    setPreviewURL((current) => {
      if (current?.startsWith("blob:")) URL.revokeObjectURL(current);
      return "";
    });
    setSelectedFile(null);
    setContentType(null);
    setLocalError("");
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setLocalError("");
    if (!selectedFile) {
      setLocalError("Please select an image or video first.");
      return;
    }
    try {
      setUploading(true);
      const uploadResponse = await uploadAPI.uploadMedia(selectedFile);
      if (!uploadResponse?.url) throw new Error("Media upload failed.");
      const mediaURL = uploadResponse.url.startsWith("http")
        ? uploadResponse.url
        : `${import.meta.env.VITE_SERVER_URL ?? "http://localhost:8080"}${uploadResponse.url}`;
      await createPost({ contentType, contentUrl: mediaURL, caption: caption.trim() });
      navigate("/");
    } catch (submitError) {
      setLocalError(submitError.message || "Failed to publish post.");
    } finally {
      setUploading(false);
    }
  };

  const isSubmitting = loading || uploading;
  const displayError = localError || error;

  return (
    <main className="min-h-[calc(100vh-64px)] bg-slate-50/80 pb-10">
      <section className="mx-auto w-full max-w-[760px] px-3 sm:px-5">
        <header className="sticky top-0 z-20 -mx-3 flex h-16 items-center justify-between border-b border-slate-200/80 bg-white/90 px-4 backdrop-blur-xl sm:-mx-5 sm:px-6">
          <button type="button" onClick={() => navigate(-1)} aria-label="Go back" className="flex h-10 w-10 items-center justify-start rounded-full text-slate-900 transition hover:bg-slate-100 hover:text-slate-600">
            <ArrowLeft size={21} />
          </button>
          <div className="text-center">
            <h1 className="text-[17px] font-bold tracking-tight text-slate-950">Create post</h1>
            <p className="hidden text-[11px] text-slate-400 sm:block">Share a moment with your community</p>
          </div>
          <button type="submit" form="create-post-form" disabled={isSubmitting || !selectedFile} className="flex h-9 items-center gap-1.5 rounded-full bg-blue-600 px-3.5 text-[14px] font-bold text-white shadow-sm transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-40">
            {isSubmitting ? <Loader2 size={15} className="animate-spin" /> : <Check size={15} />}
            {uploading ? "Uploading" : loading ? "Posting" : "Share"}
          </button>
        </header>

        {displayError && <div role="alert" className="mx-1 mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700">{displayError}</div>}

        <form id="create-post-form" onSubmit={handleSubmit} className="pt-4">
          {!previewURL ? (
            <div onDragOver={(event) => { event.preventDefault(); setDragActive(true); }} onDragLeave={() => setDragActive(false)} onDrop={handleDrop} className={`relative flex min-h-[430px] w-full flex-col items-center justify-center overflow-hidden rounded-[28px] border px-6 text-center transition sm:min-h-[500px] ${dragActive ? "border-blue-500 bg-blue-50 shadow-lg shadow-blue-100" : "border-slate-200 bg-white shadow-sm hover:border-slate-300"}`}>
              <div className="absolute left-5 top-5 flex items-center gap-2 rounded-full border border-slate-200 bg-white/90 px-3 py-1.5 text-xs font-semibold text-slate-600 shadow-sm backdrop-blur"><Sparkles size={14} className="text-blue-600" /> Create something</div>
              <div className="flex h-20 w-20 items-center justify-center rounded-[24px] bg-blue-50 text-blue-600 ring-8 ring-blue-50/60"><Plus size={38} strokeWidth={1.8} /></div>
              <h2 className="mt-7 text-[21px] font-bold tracking-tight text-slate-950">Add photo or video</h2>
              <p className="mt-2 max-w-sm text-sm leading-5 text-slate-500">Choose media from your device, or drag and drop it here.</p>
              <button type="button" onClick={() => fileInputRef.current?.click()} className="mt-6 rounded-full bg-slate-950 px-6 py-3 text-sm font-bold text-white shadow-sm transition hover:bg-slate-800">Browse device</button>
              <div className="mt-5 flex flex-wrap justify-center gap-2 text-[11px] font-medium text-slate-400"><span>JPG</span><span>•</span><span>PNG</span><span>•</span><span>WEBP</span><span>•</span><span>MP4</span><span>•</span><span>MOV</span><span>•</span><span>Up to 50MB</span></div>
            </div>
          ) : (
            <div className="overflow-hidden rounded-[28px] bg-slate-950 shadow-xl shadow-slate-200/70">
              <div className="relative flex aspect-square max-h-[640px] items-center justify-center sm:aspect-[4/3]">
                {contentType === "video" ? <video src={previewURL} controls playsInline preload="metadata" className="h-full w-full object-contain" /> : <img src={previewURL} alt="Selected post preview" className="h-full w-full object-contain" />}
                <button type="button" onClick={clearMedia} aria-label="Remove selected media" className="absolute right-4 top-4 flex h-10 w-10 items-center justify-center rounded-full bg-black/65 text-white backdrop-blur transition hover:bg-black/85"><X size={18} /></button>
                <div className="absolute bottom-4 left-4 flex items-center gap-1.5 rounded-full bg-black/65 px-3.5 py-2 text-xs font-bold text-white backdrop-blur">{contentType === "video" ? <Video size={14} /> : <ImageIcon size={14} />}{contentType === "video" ? "Video" : "Photo"}</div>
              </div>
              <div className="flex items-center justify-between border-t border-white/10 bg-slate-950 px-4 py-3.5 text-sm text-white">
                <button type="button" onClick={() => fileInputRef.current?.click()} className="flex items-center gap-2 rounded-full px-2 py-1 font-semibold text-white/90 transition hover:bg-white/10 hover:text-white"><RotateCcw size={15} /> Change media</button>
                <span className="text-xs font-medium text-white/45">Ready to share</span>
              </div>
            </div>
          )}

          <input ref={fileInputRef} type="file" accept="image/*,video/*" onChange={handleFileSelect} className="hidden" />
          <input ref={photoInputRef} type="file" accept="image/*" capture="environment" onChange={handleFileSelect} className="hidden" />
          <input ref={videoInputRef} type="file" accept="video/*" capture="environment" onChange={handleFileSelect} className="hidden" />

          <div className="mt-3 grid grid-cols-2 gap-3">
            <button type="button" onClick={() => { setLocalError(""); photoInputRef.current?.click(); }} className="flex h-12 items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white text-sm font-bold text-slate-800 shadow-sm transition hover:border-slate-300 hover:bg-slate-50"><ImageIcon size={18} className="text-blue-600" /> Photo</button>
            <button type="button" onClick={() => { setLocalError(""); videoInputRef.current?.click(); }} className="flex h-12 items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white text-sm font-bold text-slate-800 shadow-sm transition hover:border-slate-300 hover:bg-slate-50"><Video size={18} className="text-blue-600" /> Video</button>
          </div>

          <div className="mt-5 overflow-hidden rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
            <div className="flex items-center justify-between"><label htmlFor="caption" className="text-sm font-bold text-slate-950">Caption</label><span className="text-xs font-medium text-slate-400">{caption.length}/2200</span></div>
            <textarea id="caption" rows={4} value={caption} onChange={(event) => setCaption(event.target.value)} maxLength={2200} placeholder="Tell your community what this moment is about..." className="mt-3 w-full resize-none bg-transparent text-[15px] leading-6 text-slate-900 outline-none placeholder:text-slate-400" />
          </div>

          <div className="mt-3 flex items-center justify-between rounded-2xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
            <div><p className="text-sm font-bold text-slate-900">Post visibility</p><p className="mt-0.5 text-xs text-slate-500">Who can see this post</p></div>
            <span className="rounded-full bg-slate-100 px-3 py-1.5 text-xs font-bold text-slate-600">Public&nbsp;›</span>
          </div>

          <button type="button" onClick={() => navigate("/")} disabled={isSubmitting} className="mt-3 w-full rounded-2xl py-3 text-sm font-bold text-slate-500 transition hover:bg-white hover:text-slate-700 disabled:opacity-50">Cancel</button>
          <p className="pb-2 pt-1 text-center text-[11px] leading-5 text-slate-400">Media is uploaded securely before your post is shared.</p>
        </form>
      </section>
    </main>
  );
};
