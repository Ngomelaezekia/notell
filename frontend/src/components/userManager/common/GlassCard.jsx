import { motion } from "framer-motion";
import { floatingCard } from "../animation";

const baseStyles = `
relative
overflow-hidden
rounded-3xl
border
border-white/30
bg-white/65
backdrop-blur-2xl
shadow-[0_20px_60px_rgba(15,23,42,.08)]
transition-all
duration-300
`;

const paddingStyles = {
  none: "",
  sm: "p-4",
  md: "p-6",
  lg: "p-8",
};

const variantStyles = {
  default: "",

  elevated: `
    shadow-[0_30px_80px_rgba(15,23,42,.12)]
  `,

  flat: `
    shadow-none
    border-slate-200/60
  `,

  subtle: `
    bg-white/40
    backdrop-blur-xl
  `,
};

function GlassGlow() {
  return (
    <div
      className="
        pointer-events-none
        absolute
        inset-0
        bg-gradient-to-br
        from-white/40
        via-transparent
        to-transparent
      "
    />
  );
}

function BorderGlow() {
  return (
    <div
      className="
        pointer-events-none
        absolute
        inset-[1px]
        rounded-[22px]
        border
        border-white/20
      "
    />
  );
}

function AccentBar() {
  return (
    <div
      className="
        absolute
        top-0
        left-0
        h-1
        w-full
        bg-gradient-to-r
        from-indigo-500
        via-blue-500
        to-cyan-400
      "
    />
  );
}

export default function GlassCard({
  children,

  className = "",

  hover = true,

  padding = "md",

  variant = "default",

  showAccent = true,

  showGlow = true,

  showBorderGlow = true,

  header,

  footer,
}) {
  const Component = hover ? motion.div : "div";

  return (
    <Component
      {...(hover ? floatingCard : {})}
      className={`
        ${baseStyles}
        ${paddingStyles[padding]}
        ${variantStyles[variant]}
        ${className}
      `}
    >
      {showGlow && <GlassGlow />}

      {showBorderGlow && <BorderGlow />}

      {showAccent && <AccentBar />}

      <div className="relative z-10">
        {header && (
          <div className="mb-6">
            {header}
          </div>
        )}

        {children}

        {footer && (
          <div className="mt-6">
            {footer}
          </div>
        )}
      </div>
    </Component>
  );
}