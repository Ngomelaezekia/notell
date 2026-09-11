import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, Check, Image as ImageIcon, Loader2, Plus, RotateCcw, Video, X } from "lucide-react";
import { usePostActions } from "../hooks/usePosts";
import { uploadAPI } from "../services/post/UploadApi";

const MAX_FILE_SIZE = 50 * 1024 * 1024;

export const CreatePost = () => {
  const navigate = useNavigate();
  const fileInputRef = useRef(null);
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
    <main className="min-h-[calc(100vh-64px)] bg-white pb-8">
      <section className="mx-auto w-full max-w-[720px]">
        <header className="sticky top-0 z-20 flex h-14 items-center justify-between border-b border-slate-200 bg-white/95 px-4 backdrop-blur sm:px-5">
          <button type="button" onClick={() => navigate(-1)} aria-label="Go back" className="flex h-10 w-10 items-center justify-start text-slate-900 transition hover:text-slate-500">
            <ArrowLeft size={22} />
          </button>
          <h1 className="text-[17px] font-semibold tracking-tight text-slate-950">New post</h1>
          <button type="submit" form="create-post-form" disabled={isSubmitting || !selectedFile} className="flex h-9 items-center gap-1.5 rounded-lg px-2 text-[15px] font-semibold text-blue-600 transition hover:bg-blue-50 disabled:cursor-not-allowed disabled:opacity-40">
            {isSubmitting ? <Loader2 size={16} className="animate-spin" /> : <Check size={16} />}
            {uploading ? "Uploading" : loading ? "Posting" : "Share"}
          </button>
        </header>

        {displayError && <div className="mx-4 mt-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 sm:mx-5">{displayError}</div>}

        <form id="create-post-form" onSubmit={handleSubmit} className="px-4 pt-4 sm:px-5">
          {!previewURL ? (
            <div onDragOver={(event) => { event.preventDefault(); setDragActive(true); }} onDragLeave={() => setDragActive(false)} onDrop={handleDrop} className={`relative flex aspect-square w-full flex-col items-center justify-center overflow-hidden rounded-2xl border border-slate-200 px-6 text-center transition sm:aspect-[4/3] ${dragActive ? "border-blue-500 bg-blue-50" : "bg-slate-50 hover:bg-slate-100/70"}`}>
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-200/80 text-slate-700"><Plus size={34} strokeWidth={1.8} /></div>
              <h2 className="mt-5 text-[18px] font-semibold text-slate-950">Add photo or video</h2>
              <p className="mt-1 text-sm text-slate-500">Choose something to share with your community</p>
              <button type="button" onClick={() => fileInputRef.current?.click()} className="mt-5 rounded-lg bg-blue-600 px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-blue-700">Choose from device</button>
              <p className="mt-3 text-xs text-slate-400">JPG, PNG, MP4, MOV · Up to 50MB</p>
              <input ref={fileInputRef} type="file" accept="image/*,video/*" onChange={handleFileSelect} className="hidden" />
            </div>
          ) : (
            <div className="overflow-hidden rounded-2xl bg-black">
              <div className="relative flex aspect-square max-h-[620px] items-center justify-center sm:aspect-[4/3]">
                {contentType === "video" ? <video src={previewURL} controls playsInline className="h-full w-full object-contain" /> : <img src={previewURL} alt="Selected post preview" className="h-full w-full object-contain" />}
                <button type="button" onClick={clearMedia} aria-label="Remove selected media" className="absolute right-3 top-3 flex h-9 w-9 items-center justify-center rounded-full bg-black/65 text-white backdrop-blur transition hover:bg-black/80"><X size={18} /></button>
                <div className="absolute bottom-3 left-3 flex items-center gap-1.5 rounded-full bg-black/65 px-3 py-1.5 text-xs font-semibold text-white backdrop-blur">
                  {contentType === "video" ? <Video size={14} /> : <ImageIcon size={14} />}{contentType === "video" ? "Video" : "Photo"}
                </div>
              </div>
              <div className="flex items-center justify-between border-t border-white/10 bg-black px-4 py-3 text-sm text-white">
                <button type="button" onClick={() => fileInputRef.current?.click()} className="flex items-center gap-2 font-medium text-white/90 transition hover:text-white"><RotateCcw size={15} /> Change media</button>
                <span className="text-xs text-white/50">Ready to share</span>
              </div>
              <input ref={fileInputRef} type="file" accept="image/*,video/*" onChange={handleFileSelect} className="hidden" />
            </div>
          )}

          <div className="mt-5 grid grid-cols-2 gap-2">
            <button type="button" onClick={() => { setLocalError(""); fileInputRef.current?.click(); }} className="flex h-11 items-center justify-center gap-2 rounded-xl bg-slate-100 text-sm font-semibold text-slate-800 transition hover:bg-slate-200"><ImageIcon size={17} /> Photo</button>
            <button type="button" onClick={() => { setLocalError(""); fileInputRef.current?.click(); }} className="flex h-11 items-center justify-center gap-2 rounded-xl bg-slate-100 text-sm font-semibold text-slate-800 transition hover:bg-slate-200"><Video size={17} /> Video</button>
          </div>

          <div className="mt-6 border-t border-slate-200 pt-5">
            <label htmlFor="caption" className="block text-sm font-semibold text-slate-950">Caption</label>
            <textarea id="caption" rows={4} value={caption} onChange={(event) => setCaption(event.target.value)} maxLength={2200} placeholder="Write a caption..." className="mt-2 w-full resize-none border-0 bg-transparent p-0 text-[15px] leading-6 text-slate-900 outline-none placeholder:text-slate-400 focus:ring-0" />
            <div className="flex justify-end text-xs text-slate-400">{caption.length}/2200</div>
          </div>

          <div className="mt-4 flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3.5">
            <div><p className="text-sm font-semibold text-slate-900">Post visibility</p><p className="mt-0.5 text-xs text-slate-500">Who can see this post</p></div>
            <span className="text-sm text-slate-500">Public&nbsp; ›</span>
          </div>

          <button type="button" onClick={() => navigate("/")} disabled={isSubmitting} className="mt-4 w-full rounded-xl py-3 text-sm font-semibold text-slate-500 transition hover:bg-slate-100 hover:text-slate-700 disabled:opacity-50">Cancel</button>
          <p className="pb-3 pt-2 text-center text-xs text-slate-400">Your photo or video will be uploaded securely before the post is shared.</p>
        </form>
      </section>
    </main>
  );
};
