import { Link, useLocation } from "react-router-dom";
import { mainNavigation } from "../../config/navigation.config";

export default function MobileNavbar() {
  const { pathname } = useLocation();

  const mobileItems = mainNavigation.filter(({ showIn }) =>
    showIn.includes("mobile")
  );

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 px-4 pb-safe lg:hidden">
      <div className="flex h-16 items-center justify-around rounded-3xl border border-white/60 bg-white/80 backdrop-blur-xl shadow-2xl">
        {mobileItems.map(({ name, path, icon: Icon }) => {
          const isActive = pathname === path;

          return (
            <Link
              key={path}
              to={path}
              aria-current={isActive ? "page" : undefined}
              className="flex h-full w-full flex-col items-center justify-center gap-1 transition-colors"
            >
              <div
                className={`flex h-9 w-9 items-center justify-center rounded-2xl transition-all duration-200 ${
                  isActive
                    ? "bg-indigo-600 text-white shadow-lg shadow-indigo-500/30"
                    : "text-slate-500"
                }`}
              >
                <Icon size={20} />
              </div>

              <span
                className={`text-[11px] font-medium transition-colors ${
                  isActive ? "text-indigo-600" : "text-slate-500"
                }`}
              >
                {name}
              </span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}