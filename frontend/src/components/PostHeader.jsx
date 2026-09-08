import React from "react";
import { useNavigate } from "react-router-dom";
import { Bell, PlusIcon, Search } from "lucide-react";
import { useNotifications } from "../hooks/useNotifications";

export const Headerposts = () => {
  const navigate = useNavigate();
  const { unreadCount } = useNotifications();
  const badge = unreadCount > 99 ? "99+" : unreadCount;

  return (
    <header className="flex w-full shrink-0 items-center justify-between border-b border-neutral-800/80 bg-neutral-950/95 px-3 py-2 sm:px-4 sm:py-2.5">
      <button
        type="button"
        onClick={() => navigate("/")}
        className="cursor-pointer rounded-lg px-1.5 py-1 text-left transition hover:bg-neutral-900 active:scale-[0.98]"
        aria-label="Go to home feed"
      >
        <span className="block text-[15px] font-bold tracking-tight text-white">Notell</span>
      </button>

      <nav className="flex items-center gap-0.5" aria-label="Feed actions">
        <button
          type="button"
          onClick={() => navigate("/search")}
          className="cursor-pointer rounded-full p-2 text-neutral-300 transition hover:bg-neutral-900 hover:text-white active:scale-95"
          title="Search"
          aria-label="Search"
        >
          <Search size={19} strokeWidth={2} />
        </button>

        <button
          type="button"
          onClick={() => navigate("/notifications")}
          className="relative cursor-pointer rounded-full p-2 text-neutral-300 transition hover:bg-neutral-900 hover:text-white active:scale-95"
          title="Notifications"
          aria-label={unreadCount > 0 ? `Notifications, ${unreadCount} unread` : "Notifications"}
        >
          <Bell size={19} strokeWidth={2} />
          {unreadCount > 0 && (
            <span className="absolute right-0.5 top-0.5 flex min-w-4 items-center justify-center rounded-full border-2 border-neutral-950 bg-red-500 px-1 text-[9px] font-bold leading-3 text-white">
              {badge}
            </span>
          )}
        </button>

        <button
          type="button"
          onClick={() => navigate("/create-post")}
          className="cursor-pointer rounded-full p-2 text-neutral-300 transition hover:bg-neutral-900 hover:text-white active:scale-95"
          title="Create Post"
          aria-label="Create Post"
        >
          <PlusIcon size={20} strokeWidth={2.2} />
        </button>
      </nav>
    </header>
  );
};
