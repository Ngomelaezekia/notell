export const FILTERS = [
  { id: "original", label: "Original", css: "none" },
  { id: "vivid", label: "Vivid", css: "saturate(1.35) contrast(1.08)" },
  { id: "warm", label: "Warm", css: "sepia(0.16) saturate(1.18) contrast(1.04)" },
  { id: "cool", label: "Cool", css: "saturate(0.9) hue-rotate(10deg) contrast(1.05)" },
  { id: "mono", label: "Mono", css: "grayscale(1) contrast(1.08)" },
  { id: "soft", label: "Soft", css: "brightness(1.06) saturate(0.86) contrast(0.94)" },
];

export const DEFAULT_EDITS = {
  filter: "original",
  brightness: 100,
  contrast: 100,
  saturation: 100,
  rotation: 0,
  flipX: false,
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
    image.src = sourceUrl;
    await new Promise((resolve, reject) => {
      image.onload = resolve;
      image.onerror = () => reject(new Error("Could not prepare the image for editing."));
    });

    const quarterTurn = Math.abs(edits.rotation % 180) === 90;
    const canvas = document.createElement("canvas");
    canvas.width = quarterTurn ? image.naturalHeight : image.naturalWidth;
    canvas.height = quarterTurn ? image.naturalWidth : image.naturalHeight;
    const context = canvas.getContext("2d");
    if (!context) throw new Error("Image editor is not available in this browser.");

    context.save();
    context.translate(canvas.width / 2, canvas.height / 2);
    context.rotate((edits.rotation * Math.PI) / 180);
    context.scale(edits.flipX ? -1 : 1, 1);
    context.filter = `${getFilter(edits.filter)} brightness(${edits.brightness}%) contrast(${edits.contrast}%) saturate(${edits.saturation}%)`;
    context.drawImage(image, -image.naturalWidth / 2, -image.naturalHeight / 2);
    context.restore();

    const blob = await new Promise((resolve) => canvas.toBlob(resolve, "image/jpeg", 0.94));
    if (!blob) throw new Error("Could not export the edited image.");
    return new File([blob], file.name.replace(/\.[^.]+$/, "") + "-edited.jpg", { type: "image/jpeg", lastModified: Date.now() });
  } finally {
    URL.revokeObjectURL(sourceUrl);
  }
};
