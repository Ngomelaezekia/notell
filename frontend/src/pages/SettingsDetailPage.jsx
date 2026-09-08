import { useEffect, useMemo, useState } from "react";
import { ArrowLeft, CheckCircle2, ChevronRight, CircleHelp, Database, Mail, RotateCcw, ShieldCheck, Trash2 } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { userAPI } from "../services/user/userApi";

const SETTINGS_PREFIX = "notell:";

const sections = {
  privacy: { title: "Manage and privacy", description: "Control who can follow you and how your account is shared." },
  content: { title: "Contents", description: "Choose how media and content behave in Notell." },
  notifications: { title: "Notifications", description: "Choose the activity notifications you want to receive." },
  ads: { title: "Ads", description: "Manage local ad-personalization preferences." },
  history: { title: "History", description: "Review and clear activity stored locally on this device." },
  downloads: { title: "Media downloads", description: "Control automatic media download behavior." },
  storage: { title: "Cache and storage", description: "See local Notell storage and remove saved preferences." },
  about: { title: "About Notell", description: "Information about this website." },
  terms: { title: "Rules and terms", description: "Community rules and general terms for using Notell." },
  more: { title: "More", description: "Additional website preferences and maintenance tools." },
  help: { title: "Help center", description: "Quick answers for common Notell problems." },
  email: { title: "Email", description: "Contact information for Notell support." },
};

function getBool(key, fallback = false) {
  const value = localStorage.getItem(`${SETTINGS_PREFIX}${key}`);
  return value == null ? fallback : value === "true";
}

function Toggle({ title, description, checked, onChange }) {
  return (
    <label className="flex cursor-pointer items-center gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <span className="min-w-0 flex-1"><span className="block text-sm font-semibold text-slate-900">{title}</span><span className="mt-1 block text-xs leading-5 text-slate-500">{description}</span></span>
      <input className="sr-only" type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} />
      <span className={`relative h-6 w-11 shrink-0 rounded-full transition ${checked ? "bg-indigo-600" : "bg-slate-200"}`}><span className={`absolute top-1 h-4 w-4 rounded-full bg-white shadow transition ${checked ? "left-6" : "left-1"}`} /></span>
    </label>
  );
}

function Action({ icon: Icon, title, description, onClick, danger = false }) {
  return <button type="button" onClick={onClick} className="flex w-full items-center gap-3 rounded-2xl border border-slate-200 bg-white p-4 text-left shadow-sm hover:bg-slate-50">
    <span className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${danger ? "bg-red-50 text-red-600" : "bg-slate-100 text-slate-700"}`}><Icon size={18} /></span>
    <span className="min-w-0 flex-1"><span className={`block text-sm font-semibold ${danger ? "text-red-700" : "text-slate-900"}`}>{title}</span><span className="mt-1 block text-xs leading-5 text-slate-500">{description}</span></span>
    <ChevronRight size={17} className="text-slate-300" />
  </button>;
}

export default function SettingsDetailPage({ section = "privacy" }) {
  const navigate = useNavigate();
  const { user, updateUser } = useAuth();
  const meta = sections[section] || sections.privacy;
  const [message, setMessage] = useState("");
  const [allowFollowers, setAllowFollowers] = useState(user?.allowFollowers !== false);
  const [autoplay, setAutoplay] = useState(() => getBool("autoplay", true));
  const [notifications, setNotifications] = useState(() => getBool("notifications", true));
  const [likes, setLikes] = useState(() => getBool("notifications:likes", true));
  const [comments, setComments] = useState(() => getBool("notifications:comments", true));
  const [follows, setFollows] = useState(() => getBool("notifications:follows", true));
  const [personalizedAds, setPersonalizedAds] = useState(() => getBool("ads:personalized", true));
  const [wifiOnly, setWifiOnly] = useState(() => getBool("downloads:wifiOnly", true));
  const [autoDownload, setAutoDownload] = useState(() => getBool("downloads:auto", false));

  useEffect(() => { if (user?.allowFollowers != null) setAllowFollowers(user.allowFollowers); }, [user?.allowFollowers]);
  const storageBytes = useMemo(() => {
    let total = 0;
    for (let i = 0; i < localStorage.length; i += 1) { const key = localStorage.key(i); const value = key ? localStorage.getItem(key) || "" : ""; if (key?.startsWith(SETTINGS_PREFIX)) total += key.length + value.length; }
    return total;
  }, [message]);

  const save = (key, value, setter) => { setter(value); localStorage.setItem(`${SETTINGS_PREFIX}${key}`, String(value)); setMessage("Saved on this device."); };
  const clearLocalData = () => { Object.keys(localStorage).filter((key) => key.startsWith(SETTINGS_PREFIX)).forEach((key) => localStorage.removeItem(key)); setMessage("Notell local preferences and cache were cleared."); };
  const resetPreferences = () => { clearLocalData(); window.location.reload(); };
  const updatePrivacy = async (value) => {
    setAllowFollowers(value);
    try { await userAPI.updateProfile({ allowFollowers: value }); updateUser({ allowFollowers: value }); setMessage("Privacy setting saved."); }
    catch { setAllowFollowers(!value); setMessage("Could not save the privacy setting. Please try again."); }
  };

  const content = {
    privacy: <div className="space-y-3"><Toggle title="Allow followers" description="Let other users follow your account." checked={allowFollowers} onChange={updatePrivacy} /><div className="rounded-2xl border border-slate-200 bg-white p-4 text-xs leading-5 text-slate-500">Your profile editing tools are available from <button className="font-semibold text-indigo-600" onClick={() => navigate("/profile")}>Edit profile</button>. Privacy changes are saved to your account.</div></div>,
    content: <div className="space-y-3"><Toggle title="Autoplay videos" description="Automatically play videos while browsing the feed." checked={autoplay} onChange={(v) => save("autoplay", v, setAutoplay)} /><Toggle title="Data saver" description="Reduce media loading and bandwidth usage." checked={getBool("dataSaver")} onChange={(v) => { localStorage.setItem(`${SETTINGS_PREFIX}dataSaver`, String(v)); setMessage("Saved on this device."); }} /></div>,
    notifications: <div className="space-y-3"><Toggle title="Notifications" description="Enable Notell activity notifications on this device." checked={notifications} onChange={(v) => save("notifications", v, setNotifications)} /><Toggle title="Likes" description="Notify me when someone likes my content." checked={likes} onChange={(v) => save("notifications:likes", v, setLikes)} /><Toggle title="Comments" description="Notify me about comments on my content." checked={comments} onChange={(v) => save("notifications:comments", v, setComments)} /><Toggle title="New followers" description="Notify me when someone follows me." checked={follows} onChange={(v) => save("notifications:follows", v, setFollows)} /></div>,
    ads: <div className="space-y-3"><Toggle title="Personalized ads" description="Allow local preference signals to be used for ad personalization when supported." checked={personalizedAds} onChange={(v) => save("ads:personalized", v, setPersonalizedAds)} /><div className="rounded-2xl border border-amber-100 bg-amber-50 p-4 text-xs leading-5 text-amber-800">This preference controls the Notell setting only. It does not disable ads or change third-party advertising policies.</div></div>,
    history: <div className="space-y-3"><div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"><p className="text-sm font-semibold text-slate-900">Local activity</p><p className="mt-1 text-xs leading-5 text-slate-500">Notell can keep lightweight preferences on this browser. Server-side post history is part of your account and is not deleted by this button.</p></div><Action icon={Trash2} title="Clear local history" description="Remove Notell activity and preference keys stored on this device." danger onClick={clearLocalData} /></div>,
    downloads: <div className="space-y-3"><Toggle title="Wi-Fi only" description="Only allow automatic media downloads when a Wi-Fi connection is available." checked={wifiOnly} onChange={(v) => save("downloads:wifiOnly", v, setWifiOnly)} /><Toggle title="Automatic downloads" description="Allow supported media to be downloaded automatically." checked={autoDownload} onChange={(v) => save("downloads:auto", v, setAutoDownload)} /></div>,
    storage: <div className="space-y-3"><div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"><div className="flex items-center gap-3"><Database size={20} className="text-slate-600" /><div><p className="text-sm font-semibold text-slate-900">Notell local data</p><p className="text-xs text-slate-500">Approximately {(storageBytes / 1024).toFixed(1)} KB of Notell preferences.</p></div></div></div><Action icon={Trash2} title="Clear cache and preferences" description="Remove only keys created by Notell in this browser." danger onClick={clearLocalData} /></div>,
    about: <div className="space-y-3"><div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"><h3 className="font-bold text-slate-950">Notell</h3><p className="mt-2 text-sm leading-6 text-slate-600">Notell is a social media platform for sharing content, connecting with people and discovering posts.</p></div><div className="rounded-2xl border border-slate-200 bg-white p-5 text-xs leading-5 text-slate-500">Website information and feature availability may change as Notell continues to develop.</div></div>,
    terms: <div className="space-y-3"><div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"><h3 className="font-semibold text-slate-900">Community rules</h3><ul className="mt-3 list-disc space-y-2 pl-5 text-sm leading-6 text-slate-600"><li>Respect other people and do not harass or threaten them.</li><li>Do not upload illegal, abusive, deceptive or harmful material.</li><li>Do not impersonate people or misuse another person’s private information.</li><li>Only share content you have the right to publish.</li><li>Follow applicable laws and Notell platform policies.</li></ul></div><div className="rounded-2xl border border-slate-200 bg-white p-5 text-xs leading-5 text-slate-500">These are the current in-app community guidelines. A formal legal terms document can be added here when the website publishes one.</div></div>,
    more: <div className="space-y-3"><Action icon={RotateCcw} title="Reset preferences" description="Restore local Notell settings to their defaults." onClick={resetPreferences} /><Action icon={ShieldCheck} title="Privacy settings" description="Return to account privacy controls." onClick={() => navigate("/settings/privacy")} /></div>,
    help: <div className="space-y-3"><div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"><h3 className="font-semibold text-slate-900">Common fixes</h3><div className="mt-3 space-y-3 text-sm leading-6 text-slate-600"><p><b>Media does not load:</b> check your connection and Data saver settings.</p><p><b>Settings do not persist:</b> make sure browser storage is enabled for Notell.</p><p><b>Account changes fail:</b> refresh your session and try again.</p></div></div><Action icon={Mail} title="Contact Notell" description="Open the contact page while support email configuration is being finalized." onClick={() => navigate("/settings/email")} /></div>,
    email: <div className="space-y-3"><div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"><Mail size={22} className="text-slate-600" /><h3 className="mt-3 font-semibold text-slate-900">Support email</h3><p className="mt-1 text-sm leading-6 text-slate-600">A public Notell support email is not currently configured in the repository, so no address is being invented here.</p></div></div>,
  }[section] || null;

  return <div className="min-h-screen bg-slate-50 pb-10"><header className="sticky top-0 z-30 border-b border-slate-200/80 bg-white/90 backdrop-blur-xl"><div className="mx-auto flex h-14 max-w-3xl items-center gap-3 px-3 sm:px-6"><button type="button" onClick={() => navigate(-1)} aria-label="Back" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-slate-700 hover:bg-slate-100"><ArrowLeft size={20} /></button><div><h1 className="text-base font-bold text-slate-950">{meta.title}</h1><p className="text-[10px] text-slate-500">Notell settings</p></div></div></header><main className="mx-auto max-w-3xl space-y-4 px-3 pt-5 sm:px-6 sm:pt-7"><div><h2 className="text-xl font-bold text-slate-950">{meta.title}</h2><p className="mt-1 text-sm text-slate-500">{meta.description}</p></div>{message && <div className="flex items-center gap-2 rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-3 text-xs font-medium text-emerald-700"><CheckCircle2 size={15} />{message}</div>}{content}<button type="button" onClick={() => navigate("/settings")} className="flex w-full items-center justify-center gap-2 rounded-2xl border border-slate-200 bg-white px-4 py-3 text-sm font-semibold text-slate-700 shadow-sm hover:bg-slate-50"><CircleHelp size={16} /> Back to all settings</button></main></div>;
}
