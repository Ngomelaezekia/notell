import { Link, useLocation } from "react-router-dom";
import { mainNavigation } from "../../config/navigation.config";

export default function MobileNavbar() {
  const { pathname } = useLocation();

  const mobileItems = mainNavigation.filter(({ showIn }) =>
    showIn.includes("mobile")
  );

  return (
    <nav className="pointer-events-none fixed inset-x-0 bottom-0 z-50 px-3 pb-3 lg:hidden">
      <div className="pointer-events-auto mx-auto flex h-16 max-w-sm items-center justify-around rounded-[22px] border border-neutral-800/90 bg-neutral-950/95 px-2 shadow-[0_-12px_40px_rgba(0,0,0,0.35)] backdrop-blur-2xl">
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
                  className={`relative flex h-[52px] w-[52px] items-center justify-center rounded-[18px] border transition-all duration-200 active:scale-95 ${
                    isActive
                      ? "-translate-y-2 border-neutral-500 bg-neutral-100 shadow-xl shadow-black/40"
                      : "-translate-y-1 border-neutral-700 bg-neutral-100 shadow-lg shadow-black/30 group-hover:-translate-y-2 group-hover:border-white"
                  }`}
                >
                  <span className="absolute inset-1 rounded-[15px] bg-neutral-200 transition group-hover:bg-white" />
                  <Icon
                    size={25}
                    strokeWidth={2.5}
                    className="relative z-10 text-neutral-950"
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
              className="group flex h-full w-20 flex-col items-center justify-center gap-0.5 active:scale-95"
            >
              <span
                className={`flex h-9 w-12 items-center justify-center rounded-xl transition-all duration-200 ${
                  isActive
                    ? "bg-neutral-800 text-neutral-100"
                    : "text-neutral-500 group-hover:bg-neutral-900 group-hover:text-neutral-200"
                }`}
              >
                <Icon size={20} strokeWidth={isActive ? 2.5 : 2} />
              </span>
              <span
                className={`text-[10px] font-medium transition-colors ${
                  isActive ? "text-neutral-100" : "text-neutral-500"
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
