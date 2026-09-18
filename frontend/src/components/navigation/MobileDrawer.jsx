import { AnimatePresence, motion } from "framer-motion";
import { Link, useLocation } from "react-router-dom";
import { X, Settings, Sparkles, LogOut } from "lucide-react";
import { useAuth } from "../../context/AuthContext";
import { mainNavigation } from "../../config/navigation.config";

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

  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div {...overlayAnimation} onClick={onClose} className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm lg:hidden" />
          <motion.aside {...drawerAnimation} className="fixed inset-y-0 left-0 z-50 flex w-80 max-w-[86vw] flex-col bg-neutral-950 p-5 shadow-2xl ring-1 ring-white/10 lg:hidden">
            <header className="mb-6 flex items-center justify-between">
              <div><p className="text-xl font-black">Notell</p><p className="mt-1 text-[10px] uppercase tracking-[0.2em] text-neutral-600">Navigation</p></div>
              <button type="button" onClick={onClose} aria-label="Close menu" className="flex h-10 w-10 items-center justify-center rounded-xl bg-neutral-900 text-neutral-400 hover:text-white"><X size={20} /></button>
            </header>
            <nav className="space-y-1">
              {mainNavigation.filter(({ showIn }) => showIn.includes("drawer")).map(({ name, path, icon: Icon, isCreate }) => {
                const active = path === "/" ? location.pathname === "/" : location.pathname === path || location.pathname.startsWith(`${path}/`);
                return <Link key={path} to={path} onClick={onClose} className={`flex items-center gap-3 rounded-2xl px-3 py-3 text-sm font-semibold ${active ? "bg-neutral-100 text-neutral-950" : "text-neutral-400 hover:bg-neutral-900 hover:text-white"}`}><Icon size={18}/><span>{name}</span>{isCreate && <span className="ml-auto rounded-full bg-neutral-800 px-2 py-0.5 text-[9px] text-neutral-400">CREATE</span>}</Link>;
              })}
            </nav>
            <div className="mt-6 grid gap-2">
              <Link to="/channels/plans" onClick={onClose} className="flex items-center gap-3 rounded-2xl border border-neutral-800 bg-neutral-900/70 px-3 py-3 text-sm font-semibold text-neutral-200"><Sparkles size={18}/> Channel & Live plans</Link>
              <Link to="/settings" onClick={onClose} className="flex items-center gap-3 rounded-2xl border border-neutral-800 bg-neutral-900/70 px-3 py-3 text-sm font-semibold text-neutral-200"><Settings size={18}/> Settings</Link>
            </div>
            <button type="button" onClick={async () => { await logout(); onClose(); }} className="mt-auto flex items-center justify-center gap-3 rounded-2xl border border-red-900/50 bg-red-950/20 py-3 text-sm font-semibold text-red-300"><LogOut size={18}/> Logout</button>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
}
