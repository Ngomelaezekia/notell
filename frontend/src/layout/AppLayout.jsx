import { Link, Outlet, useNavigate } from "react-router-dom";
import { useState } from "react";
import { Bell, Search } from "lucide-react";
import { useAuth } from "../context/AuthContext";
import { useNotifications } from "../hooks/useNotifications";
import MobileNavbar from "../components/navigation/MobileNavigation";
import MobileDrawer from "../components/navigation/MobileDrawer";
import UserMenu from "../components/navigation/UserMenu";
import NotificationsPanel from "../components/NotificationsPanel";

export default function AppLayout() {
  const { user } = useAuth();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const navigate = useNavigate();
  const {
    notifications,
    unreadCount,
    loading: notificationsLoading,
    error: notificationsError,
    markRead,
    markAllRead,
  } = useNotifications(Boolean(user));

  const submitSearch = (event) => {
    event.preventDefault();
    const value = query.trim();
    if (value.length < 2) return;
    navigate(`/search?q=${encodeURIComponent(value)}`);
  };

  const toggleNotifications = () => {
    setNotificationsOpen((open) => !open);
  };

  return (
    <div className="min-h-screen bg-linear-to-br from-slate-100 via-white to-indigo-100">
      <header className="sticky top-0 z-50 flex min-h-20 items-center justify-between gap-3 border-b border-white/60 bg-white/80 px-4 py-3 shadow-sm backdrop-blur-xl md:px-6">
        <div className="flex shrink-0 items-center gap-3">
          <button
            type="button"
            onClick={() => setDrawerOpen(true)}
            aria-label="Open navigation"
            className="h-10 w-10 rounded-xl bg-slate-100 lg:hidden"
          >
            ☰
          </button>
          <Link to="/" className="bg-gradient-to-r from-indigo-600 to-cyan-500 bg-clip-text text-2xl font-black text-transparent">
            Notell
          </Link>
          <MobileDrawer open={drawerOpen} onClose={() => setDrawerOpen(false)} />
        </div>

        <form onSubmit={submitSearch} className="relative hidden w-full max-w-md md:block">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" size={18} />
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search people..."
            aria-label="Search people"
            maxLength={100}
            className="w-full rounded-xl border border-slate-200 bg-slate-50 py-2.5 pl-10 pr-4 text-sm outline-none transition focus:border-indigo-400 focus:bg-white focus:ring-4 focus:ring-indigo-100"
          />
        </form>

        <div className="flex shrink-0 items-center gap-3">
          <Link
            to="/search"
            aria-label="Search"
            className="flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-white md:hidden"
          >
            <Search size={19} />
          </Link>
          <div className="relative">
            <button
              type="button"
              onClick={toggleNotifications}
              aria-label="Notifications"
              aria-expanded={notificationsOpen}
              className="relative flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-white"
            >
              <Bell size={19} />
              {unreadCount > 0 && (
                <span className="absolute -right-1 -top-1 flex min-h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold text-white">
                  {unreadCount > 99 ? "99+" : unreadCount}
                </span>
              )}
            </button>
            {notificationsOpen && (
              <NotificationsPanel
                notifications={notifications}
                unreadCount={unreadCount}
                loading={notificationsLoading}
                error={notificationsError}
                onMarkRead={markRead}
                onMarkAllRead={markAllRead}
              />
            )}
          </div>
          <UserMenu user={user} />
        </div>
      </header>

      <div className="flex">
        <main className="flex w-full px-4 py-6 pb-24 md:px-6 lg:pb-6">
          <Outlet />
        </main>
        <MobileNavbar />
      </div>
    </div>
  );
}
