import React from "react";
import { Home, Compass, PlusSquare, User as UserIcon } from "lucide-react";

const SidebarItem = ({ icon, label, active = false }) => (
  <button
    className={`flex items-center gap-4 p-3 rounded-xl transition-colors hover:bg-neutral-900 w-full justify-center lg:justify-start ${
      active ? "text-indigo-400 font-semibold" : "text-neutral-400 hover:text-neutral-200"
    }`}
  >
    {icon}
    <span className="hidden lg:inline text-sm">{label}</span>
  </button>
);

export const Sidebar = () => {
  return (
    <aside className="hidden sm:flex flex-col items-center justify-between w-20 lg:w-64 border-r border-neutral-800 p-4 py-6">
      <div className="flex flex-col items-center lg:items-start w-full gap-8">
        <h1 className="hidden lg:block text-2xl font-bold tracking-wider text-indigo-500 px-2">
          NOTELL
        </h1>
        <div className="block lg:hidden font-black text-indigo-500 text-xl">
          N
        </div>

        <nav className="flex flex-col gap-3 w-full">
          <SidebarItem icon={<Home size={22} />} label="Home" active />
          <SidebarItem icon={<Compass size={22} />} label="Explore" />
          <SidebarItem icon={<PlusSquare size={22} />} label="Create" />
          <SidebarItem icon={<UserIcon size={22} />} label="Profile" />
        </nav>
      </div>
    </aside>
  );
};