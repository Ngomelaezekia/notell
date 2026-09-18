import { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, BarChart3, CalendarDays, Loader2, Plus, Radio, Save, Settings2 } from "lucide-react";
import { channelAPI } from "../services/channel/channelApi";
import { paymentAPI } from "../services/payment/paymentApi";
import LiveStreamStudio from "../components/LiveStreamStudio";
import { getApiErrorMessage } from "../utils/api";

const tabs = ["overview", "programs", "schedule", "audience", "content", "settings"];

export default function ChannelStudioPage() {
  const { id } = useParams();
  const [channel, setChannel] = useState(null);
  const [tab, setTab] = useState("overview");
  const [programs, setPrograms] = useState([]);
  const [schedule, setSchedule] = useState([]);
  const [audience, setAudience] = useState(null);
  const [content, setContent] = useState([]);
  const [settings, setSettings] = useState([]);
  const [programTitle, setProgramTitle] = useState("");
  const [programDescription, setProgramDescription] = useState("");
  const [scheduleTitle, setScheduleTitle] = useState("");
  const [scheduleStart, setScheduleStart] = useState("");
  const [postId, setPostId] = useState("");
  const [settingKey, setSettingKey] = useState("");
  const [settingValue, setSettingValue] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    try {
      setLoading(true);
      const result = await Promise.all([
        channelAPI.get(id),
        channelAPI.studioPrograms(id),
        channelAPI.studioSchedule(id),
        channelAPI.audience(id),
        channelAPI.studioContent(id),
        channelAPI.studioSettings(id),
        paymentAPI.entitlement("platform", "channel"),
      ]);
      setChannel(result[0]);
      setPrograms(result[1].programs || []);
      setSchedule(result[2].schedule || []);
      setAudience(result[3]);
      setContent(result[4].content || []);
      setSettings(result[5].settings || []);
      if (!result[6]?.active && !result[6]?.allowed) setError("Creator access is not currently active.");
    } catch (e) {
      setError(getApiErrorMessage(e, "Unable to load channel studio."));
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { void load(); }, [load]);

  const publishedPrograms = useMemo(() => programs.filter((item) => item.status === "PUBLISHED"), [programs]);

  const createProgram = async (event) => {
    event.preventDefault();
    if (!programTitle.trim()) return;
    setSaving(true); setError("");
    try {
      const created = await channelAPI.createProgram(id, { title: programTitle.trim(), description: programDescription.trim(), status: "PUBLISHED" });
      setPrograms((current) => [created, ...current]);
      setProgramTitle(""); setProgramDescription("");
    } catch (e) { setError(getApiErrorMessage(e, "Could not create program.")); }
    finally { setSaving(false); }
  };

  const createSchedule = async (event) => {
    event.preventDefault();
    if (!scheduleTitle.trim() || !scheduleStart) return;
    setSaving(true); setError("");
    try {
      const created = await channelAPI.createSchedule(id, { title: scheduleTitle.trim(), startsAt: new Date(scheduleStart).toISOString(), status: "SCHEDULED" });
      setSchedule((current) => [...current, created].sort((a, b) => new Date(a.startsAt) - new Date(b.startsAt)));
      setScheduleTitle(""); setScheduleStart("");
    } catch (e) { setError(getApiErrorMessage(e, "Could not create schedule item.")); }
    finally { setSaving(false); }
  };

  const linkContent = async (event) => {
    event.preventDefault();
    const numericId = Number(postId);
    if (!numericId) return;
    setSaving(true); setError("");
    try {
      const created = await channelAPI.linkContent(id, { postId: numericId, status: "PUBLISHED" });
      setContent((current) => [created, ...current]);
      setPostId("");
    } catch (e) { setError(getApiErrorMessage(e, "Could not link post content.")); }
    finally { setSaving(false); }
  };

  const saveSetting = async (event) => {
    event.preventDefault();
    if (!settingKey.trim()) return;
    setSaving(true); setError("");
    try {
      const saved = await channelAPI.setStudioSetting(id, settingKey.trim(), settingValue);
      setSettings((current) => {
        const next = current.filter((item) => item.key !== saved.key);
        return [...next, saved].sort((a, b) => String(a.key).localeCompare(String(b.key)));
      });
      setSettingKey(""); setSettingValue("");
    } catch (e) { setError(getApiErrorMessage(e, "Could not save studio setting.")); }
    finally { setSaving(false); }
  };

  if (loading) return <div className="flex min-h-screen w-full items-center justify-center bg-neutral-950 text-neutral-600"><Loader2 className="animate-spin"/></div>;
  if (!channel) return <div className="min-h-screen w-full bg-neutral-950 p-6 text-neutral-300">{error || "Channel not found."}</div>;

  return (
    <section className="min-h-screen w-full bg-neutral-950 pb-28 text-neutral-100">
      <div className="mx-auto max-w-6xl px-4 pt-5 sm:px-5">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Link to="/channels/manage" className="inline-flex items-center gap-2 text-xs font-bold text-neutral-500 hover:text-white"><ArrowLeft size={16}/> Channel management</Link>
          <Link to={`/channels/${id}`} className="rounded-xl border border-neutral-700 px-3 py-2 text-[10px] font-bold">Public channel</Link>
        </div>
        <header className="mt-5"><h1 className="text-2xl font-black">{channel.name} Studio</h1><p className="mt-1 text-xs text-neutral-500">Manage channel content, programming, schedule, audience, settings and live broadcasting.</p></header>

        {error && <div className="mt-4 rounded-xl border border-red-900/40 bg-red-950/20 p-3 text-xs text-red-300">{error}</div>}

        <div className="mt-6 flex gap-2 overflow-x-auto pb-2 scrollbar-none">
          {tabs.map((item) => <button type="button" key={item} onClick={() => setTab(item)} className={`whitespace-nowrap rounded-full px-4 py-2 text-xs font-bold ${tab===item?"bg-neutral-100 text-neutral-950":"border border-neutral-800 bg-neutral-900 text-neutral-500"}`}>{item.charAt(0).toUpperCase()+item.slice(1)}</button>)}
        </div>

        {tab === "overview" && <div className="mt-5 grid gap-4 md:grid-cols-3"><div className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><Radio size={18}/><p className="mt-3 text-xs text-neutral-500">Live studio</p><p className="mt-1 text-sm font-black">Broadcast from browser camera and mic.</p></div><div className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><CalendarDays size={18}/><p className="mt-3 text-xs text-neutral-500">Programs</p><p className="mt-1 text-sm font-black">{publishedPrograms.length} published</p></div><div className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><BarChart3 size={18}/><p className="mt-3 text-xs text-neutral-500">Audience</p><p className="mt-1 text-sm font-black">{audience?.followers ?? 0} followers · {audience?.subscribers ?? 0} subscribers</p></div></div>}

        {tab === "overview" && <LiveStreamStudio channelId={id}/>}
        {tab === "programs" && <div className="mt-5 grid gap-5 lg:grid-cols-[1fr_320px]"><div className="space-y-3">{programs.length ? programs.map((item) => <article key={item.id} className="rounded-2xl border border-neutral-800 bg-neutral-900/60 p-4"><div className="flex items-center justify-between gap-3"><div><h3 className="font-bold">{item.title}</h3><p className="mt-1 text-xs text-neutral-500">{item.description || "No description"}</p></div><span className="rounded-full border border-neutral-800 px-2 py-1 text-[9px] font-black text-neutral-500">{item.status}</span></div></article>) : <p className="text-sm text-neutral-600">No programs yet.</p>}</div><form onSubmit={createProgram} className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><h2 className="font-black">Create program</h2><input required value={programTitle} onChange={(e)=>setProgramTitle(e.target.value)} placeholder="Program title" className="mt-4 w-full rounded-xl border border-neutral-800 bg-neutral-950 px-3 py-3 text-sm outline-none"/><textarea value={programDescription} onChange={(e)=>setProgramDescription(e.target.value)} placeholder="Description" className="mt-3 min-h-24 w-full rounded-xl border border-neutral-800 bg-neutral-950 px-3 py-3 text-sm outline-none"/><button disabled={saving} className="mt-3 flex w-full items-center justify-center gap-2 rounded-xl bg-neutral-100 py-3 text-xs font-black text-neutral-950"><Plus size={14}/> Create published program</button></form></div>}

        {tab === "schedule" && <div className="mt-5 grid gap-5 lg:grid-cols-[1fr_320px]"><div className="space-y-3">{schedule.length ? schedule.map((item) => <article key={item.id} className="rounded-2xl border border-neutral-800 bg-neutral-900/60 p-4"><p className="text-sm font-bold">{item.title}</p><p className="mt-1 text-xs text-neutral-500">{new Date(item.startsAt).toLocaleString()} · {item.status}</p></article>) : <p className="text-sm text-neutral-600">No scheduled items.</p>}</div><form onSubmit={createSchedule} className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><h2 className="font-black">Schedule broadcast</h2><input required value={scheduleTitle} onChange={(e)=>setScheduleTitle(e.target.value)} placeholder="Schedule title" className="mt-4 w-full rounded-xl border border-neutral-800 bg-neutral-950 px-3 py-3 text-sm outline-none"/><input required type="datetime-local" value={scheduleStart} onChange={(e)=>setScheduleStart(e.target.value)} className="mt-3 w-full rounded-xl border border-neutral-800 bg-neutral-950 px-3 py-3 text-sm outline-none"/><button disabled={saving} className="mt-3 flex w-full items-center justify-center gap-2 rounded-xl bg-neutral-100 py-3 text-xs font-black text-neutral-950"><CalendarDays size={14}/> Schedule</button></form></div>}

        {tab === "audience" && <div className="mt-5 grid gap-4 sm:grid-cols-3"><div className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><p className="text-xs text-neutral-500">Followers</p><p className="mt-2 text-3xl font-black">{audience?.followers ?? 0}</p></div><div className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><p className="text-xs text-neutral-500">Subscribers</p><p className="mt-2 text-3xl font-black">{audience?.subscribers ?? 0}</p></div><div className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><p className="text-xs text-neutral-500">Team members</p><p className="mt-2 text-3xl font-black">{audience?.teamMembers ?? 0}</p></div></div>}

        {tab === "content" && <div className="mt-5 grid gap-5 lg:grid-cols-[1fr_320px]"><div className="space-y-3">{content.length ? content.map((item) => <article key={item.id} className="flex items-center justify-between gap-3 rounded-2xl border border-neutral-800 bg-neutral-900/60 p-4"><div><p className="text-sm font-bold">Post #{item.postId}</p><p className="mt-1 text-xs text-neutral-500">{item.status}</p></div><Link to={`/posts/${item.postId}`} className="rounded-lg border border-neutral-700 px-3 py-2 text-[10px] font-bold">Open</Link></article>) : <p className="text-sm text-neutral-600">No linked content.</p>}</div><form onSubmit={linkContent} className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><h2 className="font-black">Link post</h2><input required inputMode="numeric" value={postId} onChange={(e)=>setPostId(e.target.value.replace(/\D/g,""))} placeholder="Post ID" className="mt-4 w-full rounded-xl border border-neutral-800 bg-neutral-950 px-3 py-3 text-sm outline-none"/><button disabled={saving} className="mt-3 flex w-full items-center justify-center gap-2 rounded-xl bg-neutral-100 py-3 text-xs font-black text-neutral-950"><Plus size={14}/> Link published post</button></form></div>}

        {tab === "settings" && <div className="mt-5 grid gap-5 lg:grid-cols-[1fr_320px]"><div className="space-y-3">{settings.length ? settings.map((item) => <article key={item.key} className="rounded-2xl border border-neutral-800 bg-neutral-900/60 p-4"><div className="flex items-center justify-between gap-3"><div><p className="text-xs font-black">{item.key}</p><p className="mt-1 text-xs text-neutral-500 break-all">{item.value}</p></div><Settings2 size={16} className="text-neutral-600"/></div></article>) : <p className="text-sm text-neutral-600">No studio settings saved yet.</p>}</div><form onSubmit={saveSetting} className="rounded-3xl border border-neutral-800 bg-neutral-900/60 p-5"><h2 className="font-black">Save setting</h2><input required value={settingKey} onChange={(e)=>setSettingKey(e.target.value)} placeholder="Setting key" className="mt-4 w-full rounded-xl border border-neutral-800 bg-neutral-950 px-3 py-3 text-sm outline-none"/><textarea value={settingValue} onChange={(e)=>setSettingValue(e.target.value)} placeholder="Value" className="mt-3 min-h-24 w-full rounded-xl border border-neutral-800 bg-neutral-950 px-3 py-3 text-sm outline-none"/><button disabled={saving} className="mt-3 flex w-full items-center justify-center gap-2 rounded-xl bg-neutral-100 py-3 text-xs font-black text-neutral-950"><Save size={14}/> Save setting</button></form></div>}
      </div>
    </section>
  );
}
