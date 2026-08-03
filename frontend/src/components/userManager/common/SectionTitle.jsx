import { motion } from "framer-motion";

const titleAnimation = {
  initial: {
    opacity: 0,
    y: 10,
  },
  animate: {
    opacity: 1,
    y: 0,
  },
  transition: {
    duration: 0.35,
  },
};

const iconStyles = `
flex
items-center
justify-center
h-12
w-12
rounded-2xl
bg-gradient-to-br
from-indigo-500
via-blue-500
to-cyan-400
text-white
shadow-lg
shadow-indigo-500/20
`;

const accentStyles = `
mt-3
h-1
w-16
rounded-full
bg-gradient-to-r
from-indigo-500
via-blue-500
to-cyan-400
`;

export default function SectionTitle({
  title,
  description,
  icon: Icon,
  badge,
  actions,
  heading = "h2",
  showAccent = true,
  className = "",
}) {
  const Heading = heading;

  return (
    <motion.header
      {...titleAnimation}
      className={`mb-6 flex items-start gap-4 ${className}`}
    >
      {Icon && (
        <div className={iconStyles}>
          <Icon size={22} />
        </div>
      )}

      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-3">
          <Heading className="text-xl font-bold text-slate-800">
            {title}
          </Heading>

          {badge && (
            <span className="rounded-full bg-indigo-100 px-3 py-1 text-xs font-semibold text-indigo-600">
              {badge}
            </span>
          )}

          {actions && (
            <div className="ml-auto flex items-center gap-2">
              {actions}
            </div>
          )}
        </div>

        {description && (
          <p className="mt-1 max-w-xl text-sm text-slate-500">
            {description}
          </p>
        )}

        {showAccent && (
          <div className={accentStyles} />
        )}
      </div>
    </motion.header>
  );
}