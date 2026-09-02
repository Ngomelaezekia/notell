import { Link, useLocation } from "react-router-dom";
import { AnimatePresence, motion } from "framer-motion";
import { X, Home, UserCircle, LogOut } from "lucide-react";
import { useAuth } from "../../context/AuthContext";

const navigation = [
  { name: "Posts", path: "/", icon: Home },
  { name: "Account", path: "/profile", icon: UserCircle },
];

const drawerAnimation = {
  initial: { x: "-100%" },
  animate: { x: 0 },
  exit: { x: "-100%" },
  transition: { type: "spring", stiffness: 260, damping: 25 },
};

const overlayAnimation = { initial: { opacity: 0 }, animate: { opacity: 1 }, exit: { opacity: 0 } };

export default function MobileDrawer({ open, onClose }) {
  const location = useLocation();
  const { logout } = useAuth();

  const handleLogout = async () => {
    await logout();
    onClose();
  };

  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div {...overlayAnimation} onClick={onClose} className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm lg:hidden" />
          <motion.aside {...drawerAnimation} className="fixed inset-y-0 left-0 z-50 flex w-72 flex-col rounded-r-3xl bg-white p-6 shadow-2xl lg:hidden">
            <header className="mb-8 flex items-center justify-between">
              <h2 className="bg-gradient-to-r from-indigo-600 to-cyan-500 bg-clip-text text-xl font-black text-transparent">Notell</h2>
              <button type="button" onClick={onClose} aria-label="Close menu" className="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-100 transition hover:bg-slate-200">
                <X size={20} />
              </button>
            </header>

            <nav className="flex-1 space-y-2">
              {navigation.map(({ name, path, icon: Icon }) => {
                const isActive = location.pathname === path;
                return (
                  <Link key={path} to={path} onClick={onClose} className={`flex items-center gap-4 rounded-2xl px-4 py-3 font-medium transition-all duration-200 ${isActive ? "bg-indigo-600 text-white shadow-md" : "text-slate-600 hover:bg-slate-100 hover:text-slate-900"}`}>
                    <Icon size={20} />
                    <span>{name}</span>
                  </Link>
                );
              })}
            </nav>

            <button type="button" onClick={handleLogout} className="mt-8 flex items-center justify-center gap-3 rounded-2xl bg-red-50 py-3 font-semibold text-red-600 transition hover:bg-red-100">
              <LogOut size={20} />
              Logout
            </button>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
}
