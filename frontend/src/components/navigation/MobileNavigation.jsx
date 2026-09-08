import { Link, useLocation } from "react-router-dom";
import { mainNavigation } from "../../config/navigation.config";

export default function MobileNavbar() {
  const { pathname } = useLocation();

  const mobileItems = mainNavigation.filter(({ showIn }) =>
    showIn.includes("mobile")
  );

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 px-3 pb-safe lg:hidden">
      <div className="mx-auto flex h-[72px] max-w-md items-center justify-around rounded-[28px] border border-white/70 bg-white/90 px-2 shadow-[0_-8px_30px_rgba(15,23,42,0.10)] backdrop-blur-2xl">
        {mobileItems.map(({ name, path, icon: Icon, isCreate }) => {
          const isActive = pathname === path;

          if (isCreate) {
            return (
              <Link
                key={path}
                to={path}
                aria-label="Create post"
                aria-current={isActive ? "page" : undefined}
                className="group flex h-full w-20 items-center justify-center"
              >
                <span
                  className={`relative flex h-14 w-14 items-center justify-center rounded-[20px] border bg-white transition-all duration-200 ${
                    isActive
                      ? "-translate-y-1 border-indigo-200 shadow-xl shadow-indigo-500/20"
                      : "border-slate-200 shadow-lg shadow-slate-900/10 group-hover:-translate-y-1 group-hover:shadow-xl"
                  }`}
                >
                  <span
                    className={`absolute inset-1 rounded-[17px] transition ${
                      isActive ? "bg-indigo-50" : "bg-slate-50 group-hover:bg-indigo-50"
                    }`}
                  />
                  <Icon
                    size={27}
                    strokeWidth={2.4}
                    className={`relative z-10 transition-colors ${
                      isActive ? "fill-indigo-600 text-indigo-600" : "text-slate-700 group-hover:text-indigo-600"
                    }`}
                  />
                </span>
              </Link>
            );
          }

          return (
            <Link
              key={path}
              to={path}
              aria-current={isActive ? "page" : undefined}
              className="flex h-full w-20 flex-col items-center justify-center gap-1"
            >
              <span
                className={`flex h-9 w-9 items-center justify-center rounded-2xl transition-all duration-200 ${
                  isActive
                    ? "bg-indigo-600 text-white shadow-lg shadow-indigo-500/30"
                    : "text-slate-500"
                }`}
              >
                <Icon size={20} strokeWidth={isActive ? 2.5 : 2} />
              </span>
              <span
                className={`text-[10px] font-semibold transition-colors ${
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
