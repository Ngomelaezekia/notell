import React from "react";
import { useNavigate } from "react-router-dom";
import { Bell, PlusIcon, Search } from "lucide-react";
import { useNotifications } from "../hooks/useNotifications";

export const Headerposts = () => {
  const navigate = useNavigate();
  const { unreadCount } = useNotifications();
  const badge = unreadCount > 99 ? "99+" : unreadCount;

  return (
    <header className="sticky top-0 z-30 flex w-full shrink-0 items-center justify-between border-b border-neutral-900/80 bg-neutral-950/90 px-2.5 py-2.5 backdrop-blur-2xl sm:px-4 sm:py-3">
      <button type="button" onClick={() => navigate("/")} className="group flex min-w-0 items-center gap-2.5 rounded-xl px-2 py-1.5 text-left transition hover:bg-neutral-900/70 active:scale-[0.98]" aria-label="Go to home feed">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-neutral-800 bg-neutral-900 text-[13px] font-bold tracking-tight text-neutral-100 shadow-sm transition group-hover:border-neutral-700 group-hover:bg-neutral-800">N</span>
        <span className="min-w-0">
          <span className="block truncate text-[15px] font-bold tracking-[-0.02em] text-white sm:text-[16px]">Notell</span>
          <span className="mt-0.5 block text-[9px] font-medium uppercase tracking-[0.18em] text-neutral-600 transition group-hover:text-neutral-500">Your feed</span>
        </span>
      </button>

      <nav className="flex items-center gap-1 rounded-full border border-neutral-900/80 bg-neutral-950/60 p-1" aria-label="Feed actions">
        <button type="button" onClick={() => navigate("/search")} className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-full text-neutral-400 transition hover:bg-neutral-900 hover:text-white active:scale-95 sm:h-10 sm:w-10" title="Search" aria-label="Search">
          <Search size={18} strokeWidth={2} />
        </button>
        <button type="button" onClick={() => navigate("/notifications")} className="relative flex h-9 w-9 cursor-pointer items-center justify-center rounded-full text-neutral-400 transition hover:bg-neutral-900 hover:text-white active:scale-95 sm:h-10 sm:w-10" title="Notifications" aria-label={unreadCount > 0 ? `Notifications, ${unreadCount} unread` : "Notifications"}>
          <Bell size={18} strokeWidth={2} />
          {unreadCount > 0 && <span className="absolute right-0.5 top-0.5 flex min-w-4 items-center justify-center rounded-full border-2 border-neutral-950 bg-red-500 px-1 text-[8px] font-bold leading-3 text-white">{badge}</span>}
        </button>
        <button type="button" onClick={() => navigate("/create-post")} className="ml-0.5 inline-flex h-9 w-9 cursor-pointer items-center justify-center rounded-full bg-neutral-100 text-neutral-950 shadow-sm transition hover:bg-white hover:shadow-md active:scale-95 sm:h-10 sm:w-10" title="Create Post" aria-label="Create Post">
          <PlusIcon size={18} strokeWidth={2.5} />
        </button>
      </nav>
    </header>
  );
};
