import { Link, Outlet, useLocation } from "react-router-dom";
import { useState } from "react";
import { Home, UserCircle, Compass, Bell, Search } from "lucide-react";
import { useAuth } from "../context/AuthContext";
import MobileNavbar from "../components/navigation/MobileNavigation";
import MobileDrawer from "../components/navigation/MobileDrawer";
import UserMenu from "../components/navigation/UserMenu";
import { useUser } from "../hooks/useUser";

const navigation = [
  {
    name: "Posts",
    path: "/",
    icon: Home,
  },
  {
    name: "Explore",
    path: "/explore",
    icon: Compass,
  },
  {
    name: "Account",
    path: "/profile",
    icon: UserCircle,
  },
];

export default function AppLayout() {
  const { user } = useUser();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { logout } = useAuth();
  const location = useLocation();

  return (
    <div className="min-h-screen bg-linear-to-br from-slate-100 via-white to-indigo-100">
      {/* ====================== TOP HEADER ======================= */}
      <header className="sticky top-0 z-50 flex h-20 max-h-32 items-center justify-between border-b border-white/60 bg-white/80 px-6 shadow-sm backdrop-blur-xl">
        {/* Logo & Mobile Menu Toggle */}
        <div className="flex items-center gap-3">
          <button
            onClick={() => setDrawerOpen(true)}
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

        {/* Right Actions */}
        <div className="flex items-center gap-3">
          <button className="relative flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-white">
            <Bell size={19} />
            <span className="absolute top-2 right-2 h-2 w-2 rounded-full bg-red-500" />
          </button>

          <UserMenu user={user} />
        </div>
      </header>

      <div className="flex">
        {/* ====================== DESKTOP NAVIGATION ======================= */}
          <main className="flex px-4 py-6 pb-24 md:px-6 lg:pb-6">
             <Outlet />
          </main>

        <MobileNavbar />
      </div>
    </div>
  );
}