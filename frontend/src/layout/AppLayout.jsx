import { Outlet } from "react-router-dom";
import MobileNavbar from "../components/navigation/MobileNavigation";

export default function AppLayout() {
  return (
    <div className="min-h-screen bg-neutral-950 text-neutral-100">
      <div className="flex min-h-screen">
        <main className="flex w-full min-w-0 px-0 pb-24 md:pb-6">
          <Outlet />
        </main>
        <MobileNavbar />
      </div>
    </div>
  );
}
