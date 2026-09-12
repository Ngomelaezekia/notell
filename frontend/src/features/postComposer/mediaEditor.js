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

export const getFilter = (id) => FILTERS.find((item) => item.id === id)?.css ?? "none";

export const getMediaStyle = (edits) => ({
  filter: `${getFilter(edits.filter)} brightness(${edits.brightness}%) contrast(${edits.contrast}%) saturate(${edits.saturation}%)`,
  transform: `${edits.flipX ? "scaleX(-1) " : ""}rotate(${edits.rotation}deg)`,
});

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

    const quarterTurn = Math.abs(edits.rotation % 180) === 90;
    const ratio = CROP_RATIOS.find((item) => item.id === edits.crop)?.value ?? null;
    const orientedWidth = quarterTurn ? image.naturalHeight : image.naturalWidth;
    const orientedHeight = quarterTurn ? image.naturalWidth : image.naturalHeight;
    let targetWidth = orientedWidth;
    let targetHeight = orientedHeight;

    if (ratio) {
      if (orientedWidth / orientedHeight > ratio) targetWidth = orientedHeight * ratio;
      else targetHeight = orientedWidth / ratio;
    }

    const maxDimension = 4096;
    const scale = Math.min(1, maxDimension / Math.max(targetWidth, targetHeight));
    targetWidth = Math.max(1, Math.round(targetWidth * scale));
    targetHeight = Math.max(1, Math.round(targetHeight * scale));

    const canvas = document.createElement("canvas");
    canvas.width = targetWidth;
    canvas.height = targetHeight;
    const context = canvas.getContext("2d", { alpha: false });
    if (!context) throw new Error("Image editor is not available in this browser.");

    context.save();
    context.translate(targetWidth / 2, targetHeight / 2);
    context.rotate((edits.rotation * Math.PI) / 180);
    context.scale(edits.flipX ? -1 : 1, 1);
    context.filter = `${getFilter(edits.filter)} brightness(${edits.brightness}%) contrast(${edits.contrast}%) saturate(${edits.saturation}%)`;

    const cropWidth = ratio
      ? (image.naturalWidth / image.naturalHeight > ratio ? image.naturalHeight * ratio : image.naturalWidth)
      : image.naturalWidth;
    const cropHeight = ratio
      ? (image.naturalWidth / image.naturalHeight > ratio ? image.naturalHeight : image.naturalWidth / ratio)
      : image.naturalHeight;
    const sourceX = (image.naturalWidth - cropWidth) / 2;
    const sourceY = (image.naturalHeight - cropHeight) / 2;
    const scaleX = targetWidth / (quarterTurn ? cropHeight : cropWidth);
    const scaleY = targetHeight / (quarterTurn ? cropWidth : cropHeight);
    const drawScale = Math.max(scaleX, scaleY);
    const drawWidth = image.naturalWidth * drawScale;
    const drawHeight = image.naturalHeight * drawScale;

    if (ratio) {
      context.beginPath();
      context.rect(-targetWidth / 2, -targetHeight / 2, targetWidth, targetHeight);
      context.clip();
    }
    context.drawImage(image, -drawWidth / 2 - sourceX * drawScale, -drawHeight / 2 - sourceY * drawScale, drawWidth, drawHeight);
    context.restore();

    const blob = await new Promise((resolve) => canvas.toBlob(resolve, "image/jpeg", 0.94));
    if (!blob) throw new Error("Could not export the edited image.");
    return new File([blob], file.name.replace(/\.[^.]+$/, "") + "-edited.jpg", {
      type: "image/jpeg",
      lastModified: Date.now(),
    });
  } finally {
    URL.revokeObjectURL(sourceUrl);
  }
};
