import { useEffect, useState } from "react";
import { Bell, Menu, Settings, Sparkles } from "lucide-react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { mainNavigation } from "../config/navigation.config";
import { paymentAPI } from "../services/payment/paymentApi";
import MobileDrawer from "../components/navigation/MobileDrawer";
import MobileNavbar from "../components/navigation/MobileNavigation";
import UserMenu from "../components/navigation/UserMenu";
import NotificationsPanel from "../components/NotificationsPanel";
import { useNotifications } from "../hooks/useNotifications";

const isSelected = (pathname, path) => path === "/" ? pathname === "/" : pathname === path || pathname.startsWith(`${path}/`);

export default function AppLayout() {
  const { user } = useAuth();
  const { pathname } = useLocation();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [creatorAccess, setCreatorAccess] = useState(false);
  const notifications = useNotifications(Boolean(user?.id));

  useEffect(() => {
    let cancelled = false;
    paymentAPI.entitlement("platform", "channel")
      .then((result) => { if (!cancelled) setCreatorAccess(Boolean(result?.active || result?.allowed)); })
      .catch(() => { if (!cancelled) setCreatorAccess(false); });
    return () => { cancelled = true; };
  }, [user?.id]);

  const desktopItems = mainNavigation.filter(({ showIn }) => showIn.includes("desktop"));

  return (
    <div className="min-h-screen bg-neutral-950 text-neutral-100">
      <div className="flex min-h-screen">
        <aside className="sticky top-0 hidden h-screen w-64 shrink-0 flex-col border-r border-neutral-800 bg-neutral-950/95 px-4 py-5 lg:flex">
          <Link to="/" className="px-3 text-xl font-black tracking-tight">Notell</Link>
          <nav className="mt-6 space-y-1">
            {desktopItems.map(({ name, path, icon: Icon, isCreate }) => {
              const active = isSelected(pathname, path);
              return (
                <NavLink key={path} to={path} className={`flex items-center gap-3 rounded-2xl px-3 py-3 text-sm font-semibold transition ${active ? "bg-neutral-100 text-neutral-950" : "text-neutral-400 hover:bg-neutral-900 hover:text-white"}`}>
                  <Icon size={18}/><span>{name}</span>{isCreate && <span className="ml-auto rounded-full bg-neutral-800 px-2 py-0.5 text-[9px] text-neutral-400">NEW</span>}
                </NavLink>
              );
            })}
          </nav>
          <div className="mt-6 rounded-2xl border border-neutral-800 bg-neutral-900/60 p-3">
            <div className="flex items-center gap-2 text-xs font-black"><Sparkles size={15}/> Creator tools</div>
            <p className="mt-1 text-[11px] leading-5 text-neutral-500">{creatorAccess ? "Your channel and live tools are unlocked." : "A creator plan unlocks channel management and live streaming."}</p>
            <div className="mt-3 flex gap-2">
              <Link to="/channels/plans" className="rounded-lg bg-neutral-100 px-3 py-2 text-[10px] font-black text-neutral-950">Plans</Link>
              <Link to="/channels/manage" className="rounded-lg border border-neutral-700 px-3 py-2 text-[10px] font-bold text-neutral-300">Manage</Link>
            </div>
          </div>
          <div className="mt-auto flex items-center gap-2 border-t border-neutral-800 pt-4">
            <div className="relative">
              <button type="button" onClick={() => setNotificationsOpen((value) => !value)} className="relative flex h-10 w-10 items-center justify-center rounded-xl text-neutral-400 hover:bg-neutral-900 hover:text-white" aria-label="Notifications" aria-expanded={notificationsOpen}>
                <Bell size={18}/>{notifications.unreadCount > 0 && <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-red-500"/>}
              </button>
              {notificationsOpen && <NotificationsPanel notifications={notifications.notifications} unreadCount={notifications.unreadCount} loading={notifications.loading} error={notifications.error} onMarkRead={notifications.markRead} onMarkAllRead={notifications.markAllRead}/>}
            </div>
            <Link to="/settings" className="flex h-10 w-10 items-center justify-center rounded-xl text-neutral-400 hover:bg-neutral-900 hover:text-white" aria-label="Settings"><Settings size={18}/></Link>
            <UserMenu user={user || {}} />
          </div>
        </aside>
        <div className="min-w-0 flex-1">
          <header className="sticky top-0 z-40 flex h-14 items-center justify-between border-b border-neutral-800/80 bg-neutral-950/90 px-3 backdrop-blur-xl lg:hidden">
            <button type="button" onClick={() => setDrawerOpen(true)} className="flex h-10 w-10 items-center justify-center rounded-xl text-neutral-400 hover:bg-neutral-900 hover:text-white" aria-label="Open navigation"><Menu size={20}/></button>
            <Link to="/" className="text-base font-black">Notell</Link>
            <div className="relative">
              <button type="button" onClick={() => setNotificationsOpen((value) => !value)} className="relative flex h-10 w-10 items-center justify-center rounded-xl text-neutral-400 hover:bg-neutral-900 hover:text-white" aria-label="Notifications" aria-expanded={notificationsOpen}>
                <Bell size={18}/>{notifications.unreadCount > 0 && <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-red-500"/>}
              </button>
              {notificationsOpen && <NotificationsPanel notifications={notifications.notifications} unreadCount={notifications.unreadCount} loading={notifications.loading} error={notifications.error} onMarkRead={notifications.markRead} onMarkAllRead={notifications.markAllRead}/>}
            </div>
          </header>
          <main className="flex min-h-[calc(100vh-3.5rem)] w-full min-w-0 pb-24 md:pb-6"><Outlet /></main>
        </div>
        <MobileNavbar />
        <MobileDrawer open={drawerOpen} onClose={() => setDrawerOpen(false)} />
      </div>
    </div>
  );
}
