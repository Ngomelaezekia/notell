import { motion } from "framer-motion";
import {
  Bell,
  Settings,
  Sparkles,
} from "lucide-react";

import { fadeUp, buttonAnimation } from "./animation";

const iconButton = `
flex
h-11
w-11
items-center
justify-center
rounded-2xl
border
border-slate-200
bg-white
text-slate-600
transition-all
hover:bg-slate-50
hover:border-slate-300
`;

export default function Header({
  user = {
    name: "John Doe",
    role: "Premium User",
    avatar: null,
  },

  title = "Manage Profile",

  description = "Update your account information and preferences",

  showNotification = true,

  showSettings = true,

  notificationCount = 1,

  actions,
}) {
  const initial =
    user.name?.charAt(0)?.toUpperCase() ?? "?";

  return (
    <motion.header
      variants={fadeUp}
      initial="hidden"
      animate="visible"
      className="
        mx-auto
        w-full
        max-w-5xl
        rounded-3xl
        border
        border-white/60
        bg-white/70
        px-6
        py-5
        shadow-xl
        backdrop-blur-xl
      "
    >
      <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
        {/* Left */}
        <div className="flex items-center gap-4">
          {/* Avatar */}
          <div className="relative">
            <div
              className="
                flex
                h-16
                w-16
                items-center
                justify-center
                overflow-hidden
                rounded-2xl
                bg-gradient-to-br
                from-indigo-500
                via-blue-500
                to-cyan-400
                text-xl
                font-bold
                text-white
                shadow-lg
              "
            >
              {user.avatar ? (
                <img
                  src={user.avatar}
                  alt={user.name}
                  className="h-full w-full object-cover"
                />
              ) : (
                initial
              )}
            </div>

            <span
              className="
                absolute
                bottom-1
                right-1
                h-4
                w-4
                rounded-full
                border-2
                border-white
                bg-emerald-500
              "
            />
          </div>

          {/* Title */}
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <h1 className="text-2xl font-bold text-slate-800">
                {title}
              </h1>

              <Sparkles
                size={20}
                className="text-indigo-500"
              />
            </div>

            <p className="mt-1 max-w-xl text-sm text-slate-500">
              {description}
            </p>
          </div>
        </div>

        {/* Right */}
        <div className="flex items-center gap-3">
          {showNotification && (
            <motion.button
              {...buttonAnimation}
              type="button"
              aria-label="Notifications"
              className={`${iconButton} relative`}
            >
              <Bell size={20} />

              {notificationCount > 0 && (
                <span
                  className="
                    absolute
                    right-2
                    top-2
                    flex
                    h-4
                    min-w-4
                    items-center
                    justify-center
                    rounded-full
                    bg-red-500
                    px-1
                    text-[10px]
                    font-bold
                    text-white
                  "
                >
                  {notificationCount > 9
                    ? "9+"
                    : notificationCount}
                </span>
              )}
            </motion.button>
          )}

          {showSettings && (
            <motion.button
              {...buttonAnimation}
              type="button"
              aria-label="Settings"
              className={iconButton}
            >
              <Settings size={20} />
            </motion.button>
          )}

          {actions}
        </div>
      </div>
    </motion.header>
  );
}