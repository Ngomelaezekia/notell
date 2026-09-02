import { Link, Outlet } from "react-router-dom";
import { useState } from "react";
import { Bell } from "lucide-react";
import { useAuth } from "../context/AuthContext";
import MobileNavbar from "../components/navigation/MobileNavigation";
import MobileDrawer from "../components/navigation/MobileDrawer";
import UserMenu from "../components/navigation/UserMenu";

export default function AppLayout() {
  const { user } = useAuth();
  const [drawerOpen, setDrawerOpen] = useState(false);

  return (
    <div className="min-h-screen bg-linear-to-br from-slate-100 via-white to-indigo-100">
      <header className="sticky top-0 z-50 flex h-20 items-center justify-between border-b border-white/60 bg-white/80 px-6 shadow-sm backdrop-blur-xl">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => setDrawerOpen(true)}
            aria-label="Open navigation"
            className="h-10 w-10 rounded-xl bg-slate-100 lg:hidden"
          >
            ☰
          </button>
          <Link
            to="/"
            className="bg-gradient-to-r from-indigo-600 to-cyan-500 bg-clip-text text-2xl font-black text-transparent"
          >
            Notell
          </Link>
          <MobileDrawer open={drawerOpen} onClose={() => setDrawerOpen(false)} />
        </div>

        <div className="flex items-center gap-3">
          <button
            type="button"
            aria-label="Notifications"
            className="relative flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-white"
          >
            <Bell size={19} />
            <span className="absolute right-2 top-2 h-2 w-2 rounded-full bg-red-500" />
          </button>
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
