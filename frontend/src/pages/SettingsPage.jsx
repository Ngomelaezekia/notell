import { useState } from "react";
import {
  Bell,
  ChevronRight,
  CircleHelp,
  Clock3,
  Download,
  FileText,
  HardDrive,
  History,
  Info,
  LockKeyhole,
  LogOut,
  Megaphone,
  Moon,
  PlaySquare,
  ShieldCheck,
  SlidersHorizontal,
  UserRound,
  WifiOff,
  X,
} from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const sections = [
  {
    title: "Account",
    items: [
      { label: "Edit profile", description: "Profile picture, name, bio and profile details", icon: UserRound, to: "/profile" },
      { label: "Manage and privacy", description: "Privacy, followers and account controls", icon: LockKeyhole, to: "/settings/privacy" },
    ],
  },
  {
    title: "Activities",
    items: [
      { label: "Contents", description: "Control what you see and how your content is shown", icon: PlaySquare },
      { label: "Notifications", description: "Choose which notifications you receive", icon: Bell },
      { label: "Ads", description: "Manage ad preferences and personalization", icon: Megaphone },
      { label: "History", description: "View your activity and recently visited content", icon: History },
    ],
  },
  {
    title: "Data",
    items: [
      { label: "Media downloads", description: "Control automatic and manual media downloads", icon: Download },
      { label: "Cache and storage", description: "Manage cached media and local storage", icon: HardDrive },
      { label: "Storage settings", description: "Choose how Notell uses device storage", icon: SlidersHorizontal },
      { label: "Data saver", description: "Reduce mobile data usage when viewing media", icon: WifiOff },
    ],
  },
  {
    title: "About",
    items: [
      { label: "About Notell", description: "Learn about the website and its features", icon: Info },
      { label: "Rules and terms", description: "Community rules, terms of service and policies", icon: FileText },
      { label: "More", description: "Open-source notices, licenses and other information", icon: CircleHelp },
    ],
  },
  {
    title: "Contacts and help",
    items: [
      { label: "Help center", description: "Find answers and troubleshooting guides", icon: CircleHelp },
      { label: "Email", description: "Contact the Notell team", icon: Bell },
    ],
  },
];

const SettingRow = ({ item, onUnavailable }) => {
  const Icon = item.icon;
  const content = (
    <>
      <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-slate-700 transition group-hover:bg-indigo-50 group-hover:text-indigo-600">
        <Icon size={18} />
      </span>
      <span className="min-w-0 flex-1 text-left">
        <span className="block text-sm font-semibold text-slate-900">{item.label}</span>
        <span className="mt-0.5 block text-xs leading-5 text-slate-500">{item.description}</span>
      </span>
      <ChevronRight size={18} className="shrink-0 text-slate-300 transition group-hover:translate-x-0.5 group-hover:text-slate-500" />
    </>
  );

  if (item.to) {
    return <Link to={item.to} className="group flex items-center gap-3 px-4 py-3.5 transition hover:bg-slate-50 sm:px-5">{content}</Link>;
  }

  return <button type="button" onClick={() => onUnavailable(item.label)} className="group flex w-full items-center gap-3 px-4 py-3.5 transition hover:bg-slate-50 sm:px-5">{content}</button>;
};

export default function SettingsPage() {
  const { logout } = useAuth();
  const navigate = useNavigate();
  const [message, setMessage] = useState("");

  const handleUnavailable = (label) => {
    setMessage(`${label} settings are ready for the next configuration step.`);
  };

  const handleLogout = async () => {
    await logout();
  };

  return (
    <div className="min-h-screen bg-slate-50 pb-10">
      <header className="sticky top-0 z-30 border-b border-slate-200/80 bg-white/90 backdrop-blur-xl">
        <div className="mx-auto flex h-14 w-full max-w-3xl items-center justify-between px-3 sm:px-6">
          <button type="button" onClick={() => navigate(-1)} aria-label="Go back" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-slate-700 transition hover:bg-slate-100">
            <X size={20} className="rotate-45" />
          </button>
          <h1 className="text-sm font-bold text-slate-950">Settings</h1>
          <span className="h-9 w-9" aria-hidden="true" />
        </div>
      </header>

      <main className="mx-auto w-full max-w-3xl px-3 pt-5 sm:px-6 sm:pt-7">
        <div className="mb-5">
          <h2 className="text-xl font-bold tracking-tight text-slate-950">Notell settings</h2>
          <p className="mt-1 text-sm text-slate-500">Manage your account, activity, data, privacy and website preferences.</p>
        </div>

        {message && (
          <div className="mb-4 flex items-center justify-between gap-3 rounded-2xl border border-indigo-100 bg-indigo-50 px-4 py-3 text-xs font-medium text-indigo-700">
            <span>{message}</span>
            <button type="button" onClick={() => setMessage("")} aria-label="Dismiss" className="rounded-full p-1 hover:bg-indigo-100"><X size={15} /></button>
          </div>
        )}

        <div className="space-y-5">
          {sections.map((section) => (
            <section key={section.title} className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
              <div className="border-b border-slate-100 px-4 py-3 sm:px-5">
                <h3 className="text-xs font-bold uppercase tracking-[0.14em] text-slate-500">{section.title}</h3>
              </div>
              <div className="divide-y divide-slate-100">
                {section.items.map((item) => <SettingRow key={item.label} item={item} onUnavailable={handleUnavailable} />)}
              </div>
            </section>
          ))}

          <section className="overflow-hidden rounded-2xl border border-red-100 bg-white shadow-sm">
            <button type="button" onClick={handleLogout} className="group flex w-full items-center gap-3 px-4 py-4 text-left transition hover:bg-red-50 sm:px-5">
              <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-red-50 text-red-600 group-hover:bg-red-100"><LogOut size={18} /></span>
              <span className="flex-1"><span className="block text-sm font-bold text-red-600">Logout</span><span className="mt-0.5 block text-xs text-slate-500">Sign out of this Notell account</span></span>
            </button>
          </section>

          <div className="flex items-center justify-center gap-2 pb-4 text-[11px] text-slate-400">
            <ShieldCheck size={13} /> Your settings are saved to your Notell account
          </div>
        </div>
      </main>
    </div>
  );
}
