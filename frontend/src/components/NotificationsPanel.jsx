import { Bell, CheckCheck, Loader2 } from "lucide-react";
import { useNavigate } from "react-router-dom";
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

export default function NotificationsPanel({
  notifications,
  unreadCount,
  loading,
  error,
  onMarkRead,
  onMarkAllRead,
}) {
  const navigate = useNavigate();

  const openNotification = async (notification) => {
    if (!notification?.read) {
      try {
        await onMarkRead(notification.notificationId);
      } catch (err) {
        console.error(err);
      }
    }

    if (notification?.postId) {
      navigate(`/posts/${notification.postId}`);
    } else if (notification?.actorId) {
      navigate(`/users/${notification.actorId}`);
    }
  };

  return (
    <div className="absolute right-0 top-12 z-50 w-[min(24rem,calc(100vw-2rem))] overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-xl">
      <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
        <div className="flex items-center gap-2">
          <Bell size={17} />
          <h2 className="text-sm font-semibold text-slate-900">Notifications</h2>
          {unreadCount > 0 && (
            <span className="rounded-full bg-red-500 px-2 py-0.5 text-[10px] font-bold text-white">
              {unreadCount > 99 ? "99+" : unreadCount}
            </span>
          )}
        </div>
        {unreadCount > 0 && (
          <button type="button" onClick={onMarkAllRead} className="flex items-center gap-1 text-xs font-medium text-indigo-600 hover:text-indigo-800">
            <CheckCheck size={14} /> Mark all read
          </button>
        )}
      </div>

      <div className="max-h-[28rem] overflow-y-auto">
        {loading ? (
          <div className="flex items-center justify-center gap-2 px-4 py-10 text-sm text-slate-500">
            <Loader2 size={17} className="animate-spin" /> Loading...
          </div>
        ) : error ? (
          <div className="px-4 py-8 text-center text-sm text-red-600">Failed to load notifications.</div>
        ) : notifications.length === 0 ? (
          <div className="px-4 py-10 text-center text-sm text-slate-500">You’re all caught up.</div>
        ) : (
          notifications.map((notification) => {
            const actor = notification.actor || {};
            const avatar = getFileUrl(actor.profilePicture);
            return (
              <button
                key={notification.notificationId}
                type="button"
                onClick={() => openNotification(notification)}
                className={`flex w-full items-start gap-3 border-b border-slate-100 px-4 py-3 text-left transition hover:bg-slate-50 ${notification.read ? "bg-white" : "bg-indigo-50/60"}`}
              >
                <div className="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-full bg-slate-100 text-xs font-semibold text-slate-700">
                  {avatar ? <img src={avatar} alt="" className="h-full w-full object-cover" /> : (actor.username || "?").charAt(0).toUpperCase()}
                </div>
                <div className="min-w-0 flex-1">
                  <p className="text-sm text-slate-800">{notificationCopy(notification)}</p>
                  <p className="mt-1 text-xs text-slate-400">
                    {notification.createdAt ? new Date(notification.createdAt).toLocaleString() : ""}
                  </p>
                </div>
                {!notification.read && <span className="mt-2 h-2 w-2 shrink-0 rounded-full bg-indigo-500" />}
              </button>
            );
          })
        )}
      </div>
    </div>
  );
}
