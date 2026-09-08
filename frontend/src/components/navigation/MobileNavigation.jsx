import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../../context/AuthContext";
import { mainNavigation } from "../../config/navigation.config";

export default function MobileNavbar() {
  const { pathname } = useLocation();
  const { user } = useAuth();

  const mobileItems = mainNavigation.filter(({ showIn }) =>
    showIn.includes("mobile")
  );

  return (
    <nav
      className="pointer-events-none fixed inset-x-0 bottom-0 z-50 px-3 pb-[calc(0.75rem+env(safe-area-inset-bottom))] lg:hidden"
      aria-label="Mobile navigation"
    >
      <div className="pointer-events-auto mx-auto flex h-[68px] max-w-sm items-center justify-around rounded-[24px] border border-neutral-800/90 bg-neutral-950/95 px-1.5 shadow-[0_-12px_40px_rgba(0,0,0,0.35)] backdrop-blur-2xl supports-[backdrop-filter]:bg-neutral-950/80">
        {mobileItems.map(({ name, path, icon: Icon, isCreate }) => {
          const isProfileRoute = path === "/user";
          const isActive =
            pathname === path ||
            (isProfileRoute && user?.id && pathname === `/users/${user.id}`);

          if (isCreate) {
            return (
              <Link
                key={path}
                to={path}
                aria-label="Create post"
                aria-current={isActive ? "page" : undefined}
                className="group flex h-full w-16 items-center justify-center"
              >
                <span
                  className={`relative flex h-[50px] w-[50px] items-center justify-center rounded-[17px] border transition-all duration-200 ease-out active:scale-90 ${
                    isActive
                      ? "-translate-y-2 border-neutral-400 bg-neutral-100 shadow-xl shadow-black/40"
                      : "-translate-y-1 border-neutral-700 bg-neutral-100 shadow-lg shadow-black/30 group-hover:-translate-y-2 group-hover:border-white"
                  }`}
                >
                  <span className="absolute inset-1 rounded-[14px] bg-neutral-200 transition-colors duration-200 group-hover:bg-white" />
                  <Icon
                    size={24}
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
              className="group flex h-full w-16 flex-col items-center justify-center gap-0.5 active:scale-95"
            >
              <span
                className={`relative flex h-9 w-12 items-center justify-center rounded-xl transition-all duration-200 ease-out ${
                  isActive
                    ? "bg-neutral-800 text-neutral-100 shadow-sm shadow-black/20"
                    : "text-neutral-500 group-hover:bg-neutral-900 group-hover:text-neutral-200"
                }`}
              >
                {isActive && (
                  <span className="absolute -top-1 h-0.5 w-5 rounded-full bg-neutral-100" />
                )}
                <Icon size={20} strokeWidth={isActive ? 2.5 : 2} />
              </span>
              <span
                className={`text-[10px] font-medium transition-colors duration-200 ${
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
