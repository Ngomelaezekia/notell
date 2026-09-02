import { useState, useRef, useEffect, useCallback } from "react";
import { Link } from "react-router-dom";
import { AnimatePresence, motion } from "framer-motion";
import { User, LogOut, ChevronDown } from "lucide-react";
import { useAuth } from "../../context/AuthContext";
import { getFileUrl } from "../../utils/api";

const menuAnimation = {
  initial: { opacity: 0, y: -10, scale: 0.95 },
  animate: { opacity: 1, y: 0, scale: 1 },
  exit: { opacity: 0, y: -10, scale: 0.95 },
};

const menuItems = [{ label: "Profile", path: "/profile", icon: User }];

export default function UserMenu({ user = {} }) {
  const { logout } = useAuth();
  const [open, setOpen] = useState(false);
  const menuRef = useRef(null);
  const avatar = getFileUrl(user?.profilePicture);
  const name = user?.username || "User";

  const closeMenu = useCallback(() => setOpen(false), []);

  useEffect(() => {
    const handleClickOutside = ({ target }) => {
      if (menuRef.current && !menuRef.current.contains(target)) closeMenu();
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [closeMenu]);

  const handleLogout = async () => {
    await logout();
    closeMenu();
  };

  return (
    <div ref={menuRef} className="relative">
      <button type="button" onClick={() => setOpen((prev) => !prev)} aria-haspopup="menu" aria-expanded={open} className="flex items-center gap-2">
        <div className="flex h-10 w-10 items-center justify-center overflow-hidden rounded-2xl bg-linear-to-br from-indigo-500 to-cyan-400 font-bold text-white">
          {avatar ? <img src={avatar} alt={`${name} avatar`} className="h-full w-full object-cover" /> : <span className="text-lg">{name.charAt(0).toUpperCase()}</span>}
        </div>
        <ChevronDown size={18} className={`hidden text-slate-500 transition-transform duration-200 md:block ${open ? "rotate-180" : ""}`} />
      </button>

      <AnimatePresence>
        {open && (
          <motion.div {...menuAnimation} className="absolute right-0 mt-3 w-56 rounded-3xl border border-white/60 bg-white/90 p-3 shadow-2xl backdrop-blur-xl">
            <nav className="space-y-1">
              {menuItems.map(({ label, path, icon: Icon }) => (
                <Link key={path} to={path} onClick={closeMenu} className="flex items-center gap-3 rounded-2xl px-4 py-3 text-slate-700 transition-colors hover:bg-slate-100">
                  <Icon size={18} />
                  <span>{label}</span>
                </Link>
              ))}
            </nav>
            <hr className="my-2 border-slate-200" />
            <button type="button" onClick={handleLogout} className="flex w-full items-center gap-3 rounded-2xl px-4 py-3 text-red-600 transition-colors hover:bg-red-50">
              <LogOut size={18} />
              <span>Logout</span>
            </button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
