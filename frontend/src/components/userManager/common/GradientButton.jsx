import { motion } from "framer-motion";
import { Loader2 } from "lucide-react";
import { buttonAnimation } from "../animation";

const variants = {
  primary:
    "bg-gradient-to-r from-indigo-600 via-blue-600 to-cyan-500 text-white shadow-lg hover:shadow-indigo-500/30",

  secondary:
    "bg-white/80 backdrop-blur-xl border border-slate-200 text-slate-700 hover:bg-white",

  danger:
    "bg-gradient-to-r from-red-500 to-rose-600 text-white shadow-lg hover:shadow-red-500/30",

  success:
    "bg-gradient-to-r from-emerald-500 to-green-600 text-white shadow-lg hover:shadow-green-500/30",

  ghost:
    "bg-transparent text-slate-700 hover:bg-slate-100",
};

const sizes = {
  sm: "h-10 px-4 text-sm",

  md: "h-12 px-6 text-sm",

  lg: "h-14 px-8 text-base",
};

export default function GradientButton({
  children,
  icon: Icon,
  type = "button",
  variant = "primary",
  size = "md",
  loading = false,
  disabled = false,
  fullWidth = false,
  rounded = "2xl",
  className = "",
  onClick,
}) {
  const isDisabled = disabled || loading;

  return (
    <motion.button
      {...buttonAnimation}
      type={type}
      disabled={isDisabled}
      onClick={onClick}
      className={`
        relative
        overflow-hidden
        inline-flex
        items-center
        justify-center
        gap-2
        font-semibold
        transition-all
        duration-300
        select-none

        ${
          rounded === "full"
                 ? "rounded-full"
           : rounded === "xl"
           ? "rounded-xl"
    : rounded === "3xl"
    ? "rounded-3xl"
    : "rounded-2xl"
}

        ${variants[variant]}

        ${sizes[size]}

        ${fullWidth ? "w-full" : ""}

        ${
          isDisabled
            ? "opacity-60 cursor-not-allowed"
            : "cursor-pointer"
        }

        ${className}
      `}
    >
      {/* Shine */}

      <span
        className="
        absolute
        inset-0
        opacity-0
        hover:opacity-100
        transition-opacity
        duration-500
        bg-gradient-to-r
        from-transparent
        via-white/20
        to-transparent
      "
      />

      {/* Loading */}

      {loading ? (
        <>
          <Loader2
            size={18}
            className="animate-spin"
          />

          <span>Loading...</span>
        </>
      ) : (
        <>
          {Icon && <Icon size={18} />}

          <span>{children}</span>
        </>
      )}
    </motion.button>
  );
}