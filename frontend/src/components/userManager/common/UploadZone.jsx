import {
  useState,
  useRef,
  useId,
  useEffect,
  useCallback,
} from "react";

import { motion, AnimatePresence } from "framer-motion";

import {
  UploadCloud,
  X,
  CheckCircle2,
  AlertCircle,
} from "lucide-react";

const DEFAULT_MAX_SIZE = 5 * 1024 * 1024;

const uploadAnimation = {
  initial: {
    opacity: 0,
    scale: 0.96,
  },
  animate: {
    opacity: 1,
    scale: 1,
  },
};

const errorAnimation = {
  initial: {
    opacity: 0,
    y: -5,
  },
  animate: {
    opacity: 1,
    y: 0,
  },
};

const uploadBase =
  `
relative
flex
h-64
cursor-pointer
flex-col
items-center
justify-center
rounded-3xl
border-2
border-dashed
backdrop-blur-xl
transition-all
`;

export default function UploadZone({
  value,
  onChange,

  accept = "image/*",

  maxSize = DEFAULT_MAX_SIZE,

  label = "Upload Image",

  description = "PNG, JPG or WEBP up to 5MB",

  successMessage = "Image selected",

  avatar = false,

  disabled = false,
}) {
  const inputId = useId();

  const inputRef = useRef(null);

  const [dragActive, setDragActive] = useState(false);

  const [error, setError] = useState("");

  const preview =
    typeof value === "string"
      ? value
      : value?.preview;

  useEffect(() => {
    return () => {
      if (
        typeof value === "object" &&
        value?.preview?.startsWith("blob:")
      ) {
        URL.revokeObjectURL(value.preview);
      }
    };
  }, [value]);

  const validateFile = useCallback(
    (file) => {
      if (
        accept.startsWith("image") &&
        !file.type.startsWith("image/")
      ) {
        return "Only image files are allowed.";
      }

      if (file.size > maxSize) {
        return `Maximum file size is ${Math.round(
          maxSize / 1024 / 1024
        )}MB.`;
      }

      return null;
    },
    [accept, maxSize]
  );

  const processFile = useCallback(
    (file) => {
      const validation = validateFile(file);

      if (validation) {
        setError(validation);
        return;
      }

      setError("");

      onChange({
        file,
        preview: URL.createObjectURL(file),
      });
    },
    [onChange, validateFile]
  );

  const handleInput = ({ target }) => {
    const file = target.files?.[0];

    if (file) processFile(file);
  };

  const handleDrop = (e) => {
    e.preventDefault();

    if (disabled) return;

    setDragActive(false);

    const file = e.dataTransfer.files?.[0];

    if (file) processFile(file);
  };

  const removeFile = () => {
    if (
      typeof value === "object" &&
      value?.preview?.startsWith("blob:")
    ) {
      URL.revokeObjectURL(value.preview);
    }

    inputRef.current.value = "";

    setError("");

    onChange(null);
  };

  return (
    <div className="space-y-3">
      <input
        id={inputId}
        ref={inputRef}
        hidden
        type="file"
        accept={accept}
        disabled={disabled}
        onChange={handleInput}
      />

      <AnimatePresence mode="wait">
        {preview ? (
          <motion.div
            key="preview"
            {...uploadAnimation}
            className="group relative"
          >
            <img
              src={preview}
              alt="Preview"
              className={`object-cover shadow-xl ${
                avatar
                  ? "h-32 w-32 rounded-full border-4 border-white"
                  : "h-56 w-full rounded-3xl"
              }`}
            />

            <button
              type="button"
              onClick={removeFile}
              aria-label="Remove image"
              className="absolute right-3 top-3 rounded-full bg-black/60 p-2 text-white opacity-0 transition group-hover:opacity-100 hover:bg-black"
            >
              <X size={18} />
            </button>

            <div className="mt-3 flex items-center gap-2 text-sm font-medium text-emerald-600">
              <CheckCircle2 size={18} />
              {successMessage}
            </div>
          </motion.div>
        ) : (
          <motion.label
            key="upload"
            htmlFor={inputId}
            animate={{
              scale: dragActive ? 1.02 : 1,
            }}
            role="button"
            tabIndex={disabled ? -1 : 0}
            onKeyDown={(e) => {
              if (
                (e.key === "Enter" || e.key === " ") &&
                !disabled
              ) {
                e.preventDefault();
                inputRef.current?.click();
              }
            }}
            onDragEnter={() =>
              !disabled && setDragActive(true)
            }
            onDragLeave={() =>
              !disabled && setDragActive(false)
            }
            onDragOver={(e) => e.preventDefault()}
            onDrop={handleDrop}
            className={`${uploadBase}
              ${
                dragActive
                  ? "border-indigo-500 bg-indigo-50/80"
                  : "border-slate-300 bg-white/60 hover:bg-white"
              }
              ${
                disabled
                  ? "cursor-not-allowed opacity-60"
                  : ""
              }
            `}
          >
            <div className="rounded-full bg-indigo-100 p-4 text-indigo-600">
              <UploadCloud size={34} />
            </div>

            <h3 className="mt-4 font-semibold text-slate-700">
              {label}
            </h3>

            <p className="mt-1 text-center text-sm text-slate-500">
              {description}
            </p>
          </motion.label>
        )}
      </AnimatePresence>

      {error && (
        <motion.div
          {...errorAnimation}
          className="flex items-center gap-2 text-sm text-red-500"
        >
          <AlertCircle size={18} />
          <span>{error}</span>
        </motion.div>
      )}
    </div>
  );
}