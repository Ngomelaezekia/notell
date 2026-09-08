import { Bell, CheckCheck, Loader2, RefreshCw } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useNotifications } from "../hooks/useNotifications";
import { getFileUrl } from "../utils/api";

const notificationCopy = (notification) => {
  const username = notification?.actor?.username || "Someone";
  switch (notification?.type) {
    case "like":
      return `${username} liked your post.`;
    case "comment":
      return `${username} commented on your post.`;
    case "reply":
      return `${username} replied to your comment.`;
    case "follow":
      return `${username} started following you.`;
    default:
      return `${username} interacted with you.`;
  }
};

export default function NotificationsPage() {
  const navigate = useNavigate();
  const {
    notifications,
    unreadCount,
    loading,
    error,
    refetch,
    markRead,
    markAllRead,
  } = useNotifications(true);

  const openNotification = async (notification) => {
    if (!notification?.read) {
      try {
        await markRead(notification.notificationId);
      } catch {
        return;
      }
    }

    if (notification?.postId) navigate(`/posts/${notification.postId}`);
    else if (notification?.actorId) navigate(`/users/${notification.actorId}`);
  };

  return (
    <div className="min-h-screen w-full bg-slate-950 pb-24 text-white lg:pb-8">
      <div className="mx-auto w-full max-w-3xl">
        <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-white/10 bg-slate-950/90 px-4 backdrop-blur-xl sm:px-6">
          <div className="flex min-w-0 items-center gap-3">
            <button type="button" onClick={() => navigate(-1)} aria-label="Back" className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-white/75 transition hover:bg-white/10 hover:text-white">←</button>
            <div className="flex min-w-0 items-center gap-2">
              <Bell size={18} className="text-orange-300" />
              <h1 className="truncate text-sm font-bold">Notifications</h1>
              {unreadCount > 0 && <span className="rounded-full bg-orange-400/15 px-2 py-0.5 text-[10px] font-bold text-orange-300">{unreadCount > 99 ? "99+" : unreadCount} new</span>}
            </div>
          </div>
          <div className="flex items-center gap-1">
            <button type="button" onClick={() => void refetch()} aria-label="Refresh notifications" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-white/55 transition hover:bg-white/10 hover:text-white" disabled={loading}>
              <RefreshCw size={16} className={loading ? "animate-spin" : ""} />
            </button>
            {unreadCount > 0 && <button type="button" onClick={() => void markAllRead()} className="hidden items-center gap-1.5 rounded-full px-3 py-2 text-[11px] font-bold text-white/70 transition hover:bg-white/10 hover:text-white sm:flex"><CheckCheck size={15} /> Mark all read</button>}
          </div>
        </header>

        <main className="px-3 py-4 sm:px-6 sm:py-6">
          {unreadCount > 0 && <button type="button" onClick={() => void markAllRead()} className="mb-3 flex w-full items-center justify-center gap-2 rounded-2xl border border-white/10 bg-white/[0.04] py-3 text-xs font-bold text-white/70 transition hover:bg-white/[0.07] sm:hidden"><CheckCheck size={15} /> Mark all as read</button>}

          <section className="overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025] shadow-2xl shadow-black/20 sm:rounded-3xl">
            {loading && notifications.length === 0 ? (
              <div className="flex min-h-64 items-center justify-center gap-2 text-sm text-white/45"><Loader2 size={19} className="animate-spin" /> Loading notifications...</div>
            ) : error && notifications.length === 0 ? (
              <div className="flex min-h-64 flex-col items-center justify-center px-6 text-center"><p className="text-sm font-semibold text-red-300">{error}</p><button type="button" onClick={() => void refetch()} className="mt-4 rounded-xl bg-white/10 px-4 py-2 text-xs font-bold text-white hover:bg-white/15">Try again</button></div>
            ) : notifications.length === 0 ? (
              <div className="flex min-h-72 flex-col items-center justify-center px-6 text-center"><div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-orange-400/10 text-orange-300"><Bell size={25} /></div><h2 className="mt-4 text-base font-bold">You’re all caught up</h2><p className="mt-1 max-w-xs text-xs leading-5 text-white/40">Likes, comments, replies, and new followers will appear here.</p></div>
            ) : (
              notifications.map((notification) => {
                const actor = notification.actor || {};
                const avatar = getFileUrl(actor.profilePicture);
                return (
                  <button key={notification.notificationId} type="button" onClick={() => void openNotification(notification)} className={`flex w-full items-start gap-3 border-b border-white/[0.07] px-4 py-4 text-left transition last:border-b-0 hover:bg-white/[0.045] sm:px-5 ${notification.read ? "" : "bg-orange-400/[0.045]"}`}>
                    <div className="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full bg-slate-800 text-xs font-bold text-white/70 ring-1 ring-white/10">
                      {avatar ? <img src={avatar} alt="" className="h-full w-full object-cover" /> : (actor.username || "?").charAt(0).toUpperCase()}
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="text-sm leading-5 text-white/85">{notificationCopy(notification)}</p>
                      <p className="mt-1 text-[11px] text-white/35">{notification.createdAt ? new Date(notification.createdAt).toLocaleString() : ""}</p>
                    </div>
                    {!notification.read && <span className="mt-2 h-2 w-2 shrink-0 rounded-full bg-orange-400 shadow-[0_0_12px_rgba(251,146,60,0.5)]" />}
                  </button>
                );
              })
            )}
          </section>
        </main>
      </div>
    </div>
  );
}
