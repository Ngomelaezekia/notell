export const FILTERS = [
  { id: "original", label: "Original", css: "none" },
  { id: "vivid", label: "Vivid", css: "saturate(1.35) contrast(1.08)" },
  { id: "warm", label: "Warm", css: "sepia(0.16) saturate(1.18) contrast(1.04)" },
  { id: "cool", label: "Cool", css: "saturate(0.9) hue-rotate(10deg) contrast(1.05)" },
  { id: "mono", label: "Mono", css: "grayscale(1) contrast(1.08)" },
  { id: "soft", label: "Soft", css: "brightness(1.06) saturate(0.86) contrast(0.94)" },
];

export const CROP_RATIOS = [
  { id: "original", label: "Original", value: null },
  { id: "square", label: "1:1", value: 1 },
  { id: "portrait", label: "4:5", value: 4 / 5 },
  { id: "landscape", label: "16:9", value: 16 / 9 },
];

export const DEFAULT_EDITS = {
  filter: "original",
  brightness: 100,
  contrast: 100,
  saturation: 100,
  rotation: 0,
  flipX: false,
  crop: "original",
};

const MAX_VIDEO_OUTPUT_SIZE = 100 * 1024 * 1024;

export const getFilter = (id) => FILTERS.find((item) => item.id === id)?.css ?? "none";

const getCropRatio = (id) => CROP_RATIOS.find((item) => item.id === id)?.value ?? null;

export const getMediaStyle = (edits) => {
  const ratio = getCropRatio(edits.crop);
  const rotated = Math.abs(edits.rotation % 180) === 90;
  const displayRatio = ratio ? (rotated ? 1 / ratio : ratio) : undefined;

  return {
    filter: `${getFilter(edits.filter)} brightness(${edits.brightness}%) contrast(${edits.contrast}%) saturate(${edits.saturation}%)`,
    transform: `${edits.flipX ? "scaleX(-1) " : ""}rotate(${edits.rotation}deg)`,
    transformOrigin: "center center",
    transition: "filter 180ms ease, transform 220ms cubic-bezier(0.2, 0.8, 0.2, 1), aspect-ratio 220ms ease",
    willChange: "filter, transform",
    ...(displayRatio ? { aspectRatio: String(displayRatio), objectFit: "cover" } : {}),
  };
};

export const exportEditedImage = async (file, edits) => {
  const sourceUrl = URL.createObjectURL(file);
  try {
    const image = new Image();
    image.decoding = "async";
    image.src = sourceUrl;
    await new Promise((resolve, reject) => {
      image.onload = resolve;
      image.onerror = () => reject(new Error("Could not prepare the image for editing."));
    });

    const rotation = ((edits.rotation % 360) + 360) % 360;
    const quarterTurn = rotation === 90 || rotation === 270;
    const ratio = getCropRatio(edits.crop);

    const sourceRatio = ratio ? (quarterTurn ? 1 / ratio : ratio) : image.naturalWidth / image.naturalHeight;
    let cropWidth = image.naturalWidth;
    let cropHeight = image.naturalHeight;
    if (ratio) {
      if (image.naturalWidth / image.naturalHeight > sourceRatio) {
        cropWidth = image.naturalHeight * sourceRatio;
      } else {
        cropHeight = image.naturalWidth / sourceRatio;
      }
    }

    const sourceX = (image.naturalWidth - cropWidth) / 2;
    const sourceY = (image.naturalHeight - cropHeight) / 2;
    const orientedWidth = quarterTurn ? cropHeight : cropWidth;
    const orientedHeight = quarterTurn ? cropWidth : cropHeight;

    const maxDimension = 4096;
    const scale = Math.min(1, maxDimension / Math.max(orientedWidth, orientedHeight));
    const targetWidth = Math.max(1, Math.round(orientedWidth * scale));
    const targetHeight = Math.max(1, Math.round(orientedHeight * scale));
    const drawWidth = Math.max(1, Math.round(cropWidth * scale));
    const drawHeight = Math.max(1, Math.round(cropHeight * scale));

    const canvas = document.createElement("canvas");
    canvas.width = targetWidth;
    canvas.height = targetHeight;
    const context = canvas.getContext("2d", { alpha: true });
    if (!context) throw new Error("Image editor is not available in this browser.");

    context.save();
    context.translate(targetWidth / 2, targetHeight / 2);
    context.rotate((rotation * Math.PI) / 180);
    context.scale(edits.flipX ? -1 : 1, 1);
    context.filter = `${getFilter(edits.filter)} brightness(${edits.brightness}%) contrast(${edits.contrast}%) saturate(${edits.saturation}%)`;
    context.drawImage(image, sourceX, sourceY, cropWidth, cropHeight, -drawWidth / 2, -drawHeight / 2, drawWidth, drawHeight);
    context.restore();

    const output = file.type === "image/png"
      ? { type: "image/png", quality: undefined, extension: "png" }
      : file.type === "image/webp"
        ? { type: "image/webp", quality: 0.94, extension: "webp" }
        : { type: "image/jpeg", quality: 0.94, extension: "jpg" };
    const blob = await new Promise((resolve) => canvas.toBlob(resolve, output.type, output.quality));
    if (!blob) throw new Error("Could not export the edited image.");

    return new File([blob], file.name.replace(/\.[^.]+$/, "") + `-edited.${output.extension}`, { type: output.type, lastModified: Date.now() });
  } finally {
    URL.revokeObjectURL(sourceUrl);
  }
};

const waitForVideoEvent = (video, eventName) => new Promise((resolve, reject) => {
  const onEvent = () => { cleanup(); resolve(); };
  const onError = () => { cleanup(); reject(new Error("Could not prepare the video for trimming.")); };
  const cleanup = () => {
    video.removeEventListener(eventName, onEvent);
    video.removeEventListener("error", onError);
  };
  video.addEventListener(eventName, onEvent, { once: true });
  video.addEventListener("error", onError, { once: true });
});

const getRecorderMimeType = () => {
  if (typeof MediaRecorder === "undefined") return "";
  return [
    "video/webm;codecs=vp9,opus",
    "video/webm;codecs=vp8,opus",
    "video/webm",
  ].find((type) => MediaRecorder.isTypeSupported(type)) ?? "";
};

const waitForVideoFrameOrTime = async (video, targetTime) => {
  if (typeof video.requestVideoFrameCallback === "function") {
    await new Promise((resolve) => {
      const check = (_now, metadata) => {
        if (metadata.mediaTime >= targetTime - 0.02) {
          resolve();
          return;
        }
        video.requestVideoFrameCallback(check);
      };
      video.requestVideoFrameCallback(check);
    });
    return;
  }

  await new Promise((resolve) => {
    const check = () => {
      if (video.currentTime >= targetTime - 0.03) {
        video.removeEventListener("timeupdate", check);
        resolve();
      }
    };
    video.addEventListener("timeupdate", check);
  });
};

export const trimVideo = async (file, startSec, endSec) => {
  const requestedStart = Number(startSec);
  const requestedEnd = Number(endSec);
  if (!Number.isFinite(requestedStart) || !Number.isFinite(requestedEnd)) {
    throw new Error("Choose a valid video trim range.");
  }
  if (requestedStart < 0 || requestedEnd <= requestedStart) {
    throw new Error("Choose a valid video trim range.");
  }
  if (!HTMLMediaElement.prototype.play) throw new Error("Video trimming is not supported in this browser.");
  if (typeof MediaRecorder === "undefined") throw new Error("Video trimming is not supported in this browser.");

  const sourceUrl = URL.createObjectURL(file);
  const video = document.createElement("video");
  video.preload = "auto";
  video.playsInline = true;
  video.muted = true;
  video.src = sourceUrl;
  video.load();

  let audioContext;
  let sourceNode;
  let audioDestination;
  try {
    await waitForVideoEvent(video, "loadedmetadata");
    const duration = Number(video.duration);
    if (!Number.isFinite(duration) || duration <= 0) throw new Error("Could not read the video duration.");

    const safeStart = Math.min(requestedStart, Math.max(0, duration - 0.1));
    const safeEnd = Math.min(Math.max(safeStart + 0.1, requestedEnd), duration);
    if (safeEnd - safeStart < 0.1) throw new Error("Video clip must be at least 0.1 seconds long.");

    if (typeof video.captureStream !== "function" || typeof MediaRecorder.isTypeSupported !== "function") {
      throw new Error("Video trimming is not supported in this browser.");
    }

    const AudioContextClass = window.AudioContext || window.webkitAudioContext;
    if (!AudioContextClass) throw new Error("Audio processing is not supported in this browser.");
    audioContext = new AudioContextClass();
    await audioContext.resume();
    sourceNode = audioContext.createMediaElementSource(video);
    audioDestination = audioContext.createMediaStreamDestination();
    sourceNode.connect(audioDestination);

    video.currentTime = safeStart;
    await waitForVideoEvent(video, "seeked");

    const captured = video.captureStream();
    const outputStream = new MediaStream();
    captured.getVideoTracks().forEach((track) => outputStream.addTrack(track));
    const outputAudio = audioDestination.stream.getAudioTracks()[0];
    if (outputAudio) outputStream.addTrack(outputAudio);

    const mimeType = getRecorderMimeType();
    if (!mimeType) throw new Error("Video trimming is not supported in this browser.");

    const chunks = [];
    const recorder = new MediaRecorder(outputStream, { mimeType, videoBitsPerSecond: 6_000_000, audioBitsPerSecond: 128_000 });
    let finished = false;
    let stopTimer;

    const blob = await new Promise((resolve, reject) => {
      let videoStarted = false;
      const cleanup = () => {
        video.pause();
        if (stopTimer) clearTimeout(stopTimer);
        video.removeEventListener("timeupdate", onTimeUpdate);
        outputStream.getTracks().forEach((track) => track.stop());
        captured.getTracks().forEach((track) => track.stop());
      };
      const fail = (error) => { cleanup(); reject(error); };
      const finish = () => {
        if (finished) return;
        finished = true;
        recorder.stop();
      };
      const onTimeUpdate = () => {
        if (video.currentTime >= safeEnd - 0.03) finish();
      };

      recorder.addEventListener("dataavailable", (event) => { if (event.data?.size) chunks.push(event.data); });
      recorder.addEventListener("error", () => fail(new Error("Could not render the trimmed video.")), { once: true });
      recorder.addEventListener("stop", () => {
        cleanup();
        if (!chunks.length) { reject(new Error("Could not render the trimmed video.")); return; }
        resolve(new Blob(chunks, { type: mimeType }));
      }, { once: true });
      video.addEventListener("timeupdate", onTimeUpdate);
      stopTimer = window.setTimeout(finish, Math.max(500, (safeEnd - safeStart + 0.5) * 1000));

      recorder.start(250);
      video.play().then(async () => {
        if (videoStarted) return;
        videoStarted = true;
        try {
          await waitForVideoFrameOrTime(video, safeEnd);
          finish();
        } catch (error) {
          fail(error);
        }
      }).catch(() => fail(new Error("The browser blocked video trimming playback. Try again.")));
    });

    if (blob.size > MAX_VIDEO_OUTPUT_SIZE) {
      throw new Error("The trimmed video is still larger than 100MB. Choose a shorter clip.");
    }

    return new File([blob], file.name.replace(/\.[^.]+$/, "") + `-trimmed-${Math.round(safeStart)}-${Math.round(safeEnd)}.webm`, { type: "video/webm", lastModified: Date.now() });
  } finally {
    try { sourceNode?.disconnect(); } catch {}
    try { audioContext?.close(); } catch {}
    URL.revokeObjectURL(sourceUrl);
    video.removeAttribute("src");
    video.load();
  }
};
