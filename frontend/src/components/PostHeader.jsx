import React from "react";
import { useNavigate } from "react-router-dom";
import { Bell, PlusIcon, Search } from "lucide-react";
import { useNotifications } from "../hooks/useNotifications";

export const Headerposts = () => {
  const navigate = useNavigate();
  const { unreadCount } = useNotifications();
  const badge = unreadCount > 99 ? "99+" : unreadCount;

  return (
    <header className="sticky top-0 z-30 flex w-full shrink-0 items-center justify-between border-b border-neutral-900/80 bg-neutral-950/90 px-3 py-2.5 backdrop-blur-xl sm:px-5 sm:py-3">
      <button type="button" onClick={() => navigate("/")} className="group cursor-pointer rounded-xl px-2 py-1.5 text-left transition hover:bg-neutral-900/70 active:scale-[0.98]" aria-label="Go to home feed">
        <span className="block text-[16px] font-bold tracking-[-0.02em] text-white">Notell</span>
        <span className="mt-0.5 block text-[9px] font-medium uppercase tracking-[0.18em] text-neutral-600 transition group-hover:text-neutral-500">Your feed</span>
      </button>

      <nav className="flex items-center gap-1" aria-label="Feed actions">
        <button type="button" onClick={() => navigate("/search")} className="cursor-pointer rounded-full p-2.5 text-neutral-400 transition hover:bg-neutral-900 hover:text-white active:scale-95" title="Search" aria-label="Search">
          <Search size={19} strokeWidth={2} />
        </button>
        <button type="button" onClick={() => navigate("/notifications")} className="relative cursor-pointer rounded-full p-2.5 text-neutral-400 transition hover:bg-neutral-900 hover:text-white active:scale-95" title="Notifications" aria-label={unreadCount > 0 ? `Notifications, ${unreadCount} unread` : "Notifications"}>
          <Bell size={19} strokeWidth={2} />
          {unreadCount > 0 && <span className="absolute right-1 top-1 flex min-w-4 items-center justify-center rounded-full border-2 border-neutral-950 bg-red-500 px-1 text-[9px] font-bold leading-3 text-white">{badge}</span>}
        </button>
        <button type="button" onClick={() => navigate("/create-post")} className="ml-0.5 inline-flex h-9 w-9 cursor-pointer items-center justify-center rounded-full bg-neutral-100 text-neutral-950 shadow-sm transition hover:bg-white active:scale-95" title="Create Post" aria-label="Create Post">
          <PlusIcon size={19} strokeWidth={2.4} />
        </button>
      </nav>
    </header>
  );
};
