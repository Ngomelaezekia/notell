import React from "react";
import { useNavigate } from "react-router-dom";
import { Bell, PlusIcon, Search } from "lucide-react";
import { useNotifications } from "../hooks/useNotifications";

export const Headerposts = ({ title = "Feed" }) => {
  const navigate = useNavigate();
  const { unreadCount } = useNotifications();
  const badge = unreadCount > 99 ? "99+" : unreadCount;

  return (
    <header className="flex w-full shrink-0 items-center justify-between border-b border-neutral-800 bg-neutral-950 px-3 py-2.5 sm:px-4 sm:py-3">
      <h1 className="text-sm font-semibold text-neutral-200">{title}</h1>

      <div className="flex items-center gap-1">
        <button
          type="button"
          onClick={() => navigate("/search")}
          className="cursor-pointer rounded-full p-2 text-neutral-300 transition hover:bg-neutral-900 hover:text-white active:scale-95"
          title="Search"
          aria-label="Search"
        >
          <Search size={19} />
        </button>

        <button
          type="button"
          onClick={() => navigate("/notifications")}
          className="relative cursor-pointer rounded-full p-2 text-neutral-300 transition hover:bg-neutral-900 hover:text-white active:scale-95"
          title="Notifications"
          aria-label={unreadCount > 0 ? `Notifications, ${unreadCount} unread` : "Notifications"}
        >
          <Bell size={19} />
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
          <PlusIcon size={20} />
        </button>
      </div>
    </header>
  );
};
