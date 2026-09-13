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

    // Crop in the source orientation first. For a quarter-turn the desired
    // display ratio is inverted relative to the source image.
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
    context.drawImage(
      image,
      sourceX,
      sourceY,
      cropWidth,
      cropHeight,
      -drawWidth / 2,
      -drawHeight / 2,
      drawWidth,
      drawHeight,
    );
    context.restore();

    const output = file.type === "image/png"
      ? { type: "image/png", quality: undefined, extension: "png" }
      : file.type === "image/webp"
        ? { type: "image/webp", quality: 0.94, extension: "webp" }
        : { type: "image/jpeg", quality: 0.94, extension: "jpg" };
    const blob = await new Promise((resolve) => canvas.toBlob(resolve, output.type, output.quality));
    if (!blob) throw new Error("Could not export the edited image.");

    return new File([blob], file.name.replace(/\.[^.]+$/, "") + `-edited.${output.extension}`, {
      type: output.type,
      lastModified: Date.now(),
    });
  } finally {
    URL.revokeObjectURL(sourceUrl);
  }
};
