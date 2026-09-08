import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  ArrowLeft,
  Check,
  Film,
  Image as ImageIcon,
  Loader2,
  Play,
  RotateCcw,
  UploadCloud,
  Video,
  X,
} from "lucide-react";
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

  useEffect(() => {
    return () => {
      if (previewURL?.startsWith("blob:")) {
        URL.revokeObjectURL(previewURL);
      }
    };
  }, [previewURL]);

  const validateFile = (file) => {
    if (!file?.type?.startsWith("image/") && !file?.type?.startsWith("video/")) {
      throw new Error("Choose an image or video file.");
    }

    if (file.size > MAX_FILE_SIZE) {
      throw new Error("File size must be below 50MB.");
    }
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

      if (!uploadResponse?.url) {
        throw new Error("Media upload failed.");
      }

      const mediaURL = uploadResponse.url.startsWith("http")
        ? uploadResponse.url
        : `${import.meta.env.VITE_SERVER_URL ?? "http://localhost:8080"}${uploadResponse.url}`;

      await createPost({
        contentType,
        contentUrl: mediaURL,
        caption: caption.trim(),
      });

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
    <section className="mx-auto w-full max-w-xl pb-6 md:max-w-2xl">
      <div className="overflow-hidden rounded-[28px] border border-slate-200/80 bg-white shadow-xl shadow-slate-900/5">
        <header className="flex h-16 items-center justify-between border-b border-slate-100 px-4 sm:px-6">
          <button
            type="button"
            onClick={() => navigate(-1)}
            aria-label="Go back"
            className="flex h-10 w-10 items-center justify-center rounded-full text-slate-700 transition hover:bg-slate-100"
          >
            <ArrowLeft size={20} />
          </button>
          <h1 className="text-base font-bold text-slate-900">Create post</h1>
          <button
            type="submit"
            form="create-post-form"
            disabled={isSubmitting || !selectedFile}
            className="flex h-10 items-center gap-1.5 rounded-full bg-indigo-600 px-4 text-sm font-bold text-white transition hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {isSubmitting ? <Loader2 size={16} className="animate-spin" /> : <Check size={16} />}
            {uploading ? "Uploading" : loading ? "Posting" : "Share"}
          </button>
        </header>

        {displayError && (
          <div className="mx-4 mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 sm:mx-6">
            {displayError}
          </div>
        )}

        <form id="create-post-form" onSubmit={handleSubmit} className="space-y-5 p-4 sm:p-6">
          {!previewURL ? (
            <div
              onDragOver={(event) => {
                event.preventDefault();
                setDragActive(true);
              }}
              onDragLeave={() => setDragActive(false)}
              onDrop={handleDrop}
              className={`relative flex min-h-[360px] flex-col items-center justify-center overflow-hidden rounded-[24px] border-2 border-dashed px-6 text-center transition sm:min-h-[430px] ${
                dragActive
                  ? "border-indigo-500 bg-indigo-50"
                  : "border-slate-200 bg-slate-50/80 hover:border-indigo-300 hover:bg-indigo-50/40"
              }`}
            >
              <div className="mb-5 flex h-20 w-20 items-center justify-center rounded-[24px] bg-white shadow-sm ring-1 ring-slate-200">
                <ImageIcon size={34} className="text-slate-500" />
              </div>
              <h2 className="text-lg font-bold text-slate-900">Create a new post</h2>
              <p className="mt-2 max-w-xs text-sm leading-6 text-slate-500">
                Share a photo or video with your community.
              </p>

              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="mt-6 rounded-full bg-indigo-600 px-6 py-3 text-sm font-bold text-white shadow-lg shadow-indigo-500/20 transition hover:bg-indigo-700"
              >
                Choose from device
              </button>
              <p className="mt-3 text-xs text-slate-400">JPG, PNG, MP4, MOV · Up to 50MB</p>

              <input
                ref={fileInputRef}
                type="file"
                accept="image/*,video/*"
                onChange={handleFileSelect}
                className="hidden"
              />
            </div>
          ) : (
            <div className="overflow-hidden rounded-[24px] bg-black">
              <div className="relative flex min-h-[360px] max-h-[600px] items-center justify-center sm:min-h-[500px]">
                {contentType === "video" ? (
                  <video src={previewURL} controls playsInline className="max-h-[600px] w-full object-contain" />
                ) : (
                  <img src={previewURL} alt="Selected post preview" className="max-h-[600px] w-full object-contain" />
                )}

                <button
                  type="button"
                  onClick={clearMedia}
                  aria-label="Remove selected media"
                  className="absolute right-3 top-3 flex h-10 w-10 items-center justify-center rounded-full bg-black/65 text-white backdrop-blur transition hover:bg-black/80"
                >
                  <X size={19} />
                </button>

                <div className="absolute bottom-3 left-3 flex items-center gap-2 rounded-full bg-black/65 px-3 py-2 text-xs font-semibold text-white backdrop-blur">
                  {contentType === "video" ? <Video size={15} /> : <ImageIcon size={15} />}
                  {contentType === "video" ? "Video" : "Photo"}
                </div>
              </div>

              <div className="flex items-center justify-between border-t border-white/10 bg-black/90 px-4 py-3 text-white">
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  className="flex items-center gap-2 text-sm font-semibold text-white/85 hover:text-white"
                >
                  <RotateCcw size={16} /> Change media
                </button>
                <span className="text-xs text-white/50">Ready to share</span>
              </div>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*,video/*"
                onChange={handleFileSelect}
                className="hidden"
              />
            </div>
          )}

          <div className="rounded-[22px] border border-slate-200 bg-slate-50/70 p-4">
            <div className="flex items-start gap-3">
              <div className="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-indigo-100 text-indigo-600">
                <Play size={16} fill="currentColor" />
              </div>
              <div>
                <p className="text-sm font-bold text-slate-900">Add a caption</p>
                <p className="mt-0.5 text-xs text-slate-500">Tell people what this post is about.</p>
              </div>
            </div>

            <textarea
              rows={4}
              value={caption}
              onChange={(event) => setCaption(event.target.value)}
              maxLength={2200}
              placeholder="Write a caption..."
              className="mt-4 w-full resize-none rounded-2xl border border-slate-200 bg-white p-4 text-sm leading-6 text-slate-900 outline-none placeholder:text-slate-400 focus:border-indigo-400 focus:ring-4 focus:ring-indigo-100"
            />
            <div className="mt-2 text-right text-xs text-slate-400">{caption.length}/2200</div>
          </div>

          <div className="flex items-center gap-3 rounded-2xl bg-slate-50 px-4 py-3 text-xs text-slate-500">
            <Film size={17} className="shrink-0 text-slate-400" />
            <span>Your photo or video will be uploaded securely before the post is shared.</span>
          </div>

          <button
            type="button"
            onClick={() => navigate("/")}
            disabled={isSubmitting}
            className="w-full rounded-2xl py-3 text-sm font-semibold text-slate-500 transition hover:bg-slate-100 hover:text-slate-700 disabled:opacity-50"
          >
            Cancel
          </button>
        </form>
      </div>

      <div className="mt-4 flex items-center justify-center gap-2 text-xs text-slate-400">
        <UploadCloud size={14} /> Images and videos only · 50MB maximum
      </div>
    </section>
  );
};
