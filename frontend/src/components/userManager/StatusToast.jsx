import { useEffect, useMemo } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  CheckCircle2,
  AlertCircle,
  AlertTriangle,
  Info,
  X,
} from "lucide-react";

import { toastAnimation } from "./animation";

const TOAST_TYPES = {
  success: {
    icon: CheckCircle2,
    iconClass: "text-emerald-600",
    bgClass: "bg-emerald-50",
  },

  error: {
    icon: AlertCircle,
    iconClass: "text-red-600",
    bgClass: "bg-red-50",
  },

  warning: {
    icon: AlertTriangle,
    iconClass: "text-amber-600",
    bgClass: "bg-amber-50",
  },

  info: {
    icon: Info,
    iconClass: "text-blue-600",
    bgClass: "bg-blue-50",
  },
};

const POSITIONS = {
  "bottom-right": "bottom-6 right-6",
  "bottom-left": "bottom-6 left-6",
  "top-right": "top-6 right-6",
  "top-left": "top-6 left-6",
  "top-center": "top-6 left-1/2 -translate-x-1/2",
  "bottom-center": "bottom-6 left-1/2 -translate-x-1/2",
};

export default function StatusToast({
  type = "success",

  title,

  message,

  duration = 3000,

  position = "bottom-right",

  showProgress = true,

  onClose,
}) {
  const config = useMemo(
    () => TOAST_TYPES[type] ?? TOAST_TYPES.info,
    [type]
  );

  const Icon = config.icon;

  useEffect(() => {
    if (!message || duration <= 0) return;

    const timer = setTimeout(() => {
      onClose?.();
    }, duration);

    return () => clearTimeout(timer);
  }, [message, duration, onClose]);

  return (
    <AnimatePresence>
      {message && (
        <motion.div
          variants={toastAnimation}
          initial="hidden"
          animate="visible"
          exit="exit"
          role="alert"
          aria-live="polite"
          className={`
            fixed
            z-50
            max-w-sm
            overflow-hidden
            rounded-2xl
            border
            border-white/60
            bg-white/80
            backdrop-blur-xl
            shadow-2xl
            ${POSITIONS[position]}
          `}
        >
          <div className="flex items-start gap-3 p-5">
            <div
              className={`
                rounded-xl
                p-2
                ${config.bgClass}
                ${config.iconClass}
              `}
            >
              <Icon size={22} />
            </div>

            <div className="min-w-0 flex-1">
              {title && (
                <h4 className="font-semibold text-slate-800">
                  {title}
                </h4>
              )}

              <p className="text-sm text-slate-600">
                {message}
              </p>
            </div>

            <button
              type="button"
              aria-label="Close notification"
              onClick={onClose}
              className="
                rounded-lg
                p-1
                text-slate-400
                transition
                hover:bg-slate-100
                hover:text-slate-700
              "
            >
              <X size={18} />
            </button>
          </div>

          {showProgress && duration > 0 && (
            <motion.div
              className="h-1 bg-indigo-500"
              initial={{ width: "100%" }}
              animate={{ width: "0%" }}
              transition={{
                duration: duration / 1000,
                ease: "linear",
              }}
            />
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
}