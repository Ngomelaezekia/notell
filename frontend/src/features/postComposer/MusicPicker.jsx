import { useEffect, useMemo, useRef, useState } from "react";
import { Check, ChevronDown, Clock3, Headphones, Loader2, Music2, Search, Upload, X } from "lucide-react";
import musicServiceAPI from "../../services/music/musicApi";

const country = import.meta.env.VITE_MUSIC_COUNTRY || "TZ";
const formatTime = (value) => {
  const seconds = Math.max(0, Math.round(Number(value) || 0));
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}`;
};

export const MusicPicker = ({ value, onChange, disabled = false }) => {
  const [open, setOpen] = useState(false);
  const [tab, setTab] = useState("cloud");
  const [query, setQuery] = useState("");
  const [tracks, setTracks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [playingId, setPlayingId] = useState(null);
  const [selected, setSelected] = useState(value || null);
  const [startSec, setStartSec] = useState(value?.startSec ?? 0);
  const [endSec, setEndSec] = useState(value?.endSec ?? 0);
  const [localFile, setLocalFile] = useState(value?.source === "local" ? value : null);
  const audioRef = useRef(null);

  useEffect(() => setSelected(value || null), [value]);

  useEffect(() => {
    if (!open || tab !== "cloud") return undefined;
    let cancelled = false;
    const timer = window.setTimeout(async () => {
      setLoading(true);
      try {
        const data = query.trim()
          ? await musicServiceAPI.search(query.trim(), { country, limit: 24 })
          : await musicServiceAPI.trending({ country, days: 7 });
        if (!cancelled) setTracks(data?.tracks || []);
      } catch {
        if (!cancelled) setTracks([]);
      } finally {
        if (!cancelled) setLoading(false);
      }
    }, query.trim() ? 280 : 0);
    return () => { cancelled = true; window.clearTimeout(timer); };
  }, [open, query, tab]);

  const previewURL = selected?.previewUrl || selected?.previewURL || "";
  const maxDuration = Number(selected?.durationSec || 0);
  const effectiveEnd = endSec > startSec ? endSec : Math.min(maxDuration, startSec + 30);
  const canUse = Boolean(selected?.canUseInPost && selected?.rights?.licensed && selected?.rights?.ugcUse && selected?.rights?.streaming);

  const chooseCloud = async (track) => {
    if (!track?.canUseInPost || !track?.rights?.licensed || !track?.rights?.ugcUse || !track?.rights?.streaming) return;
    setSelected(track);
    setStartSec(0);
    setEndSec(Math.min(Number(track.durationSec) || 30, 30));
    setPlayingId(track.id);
  };

  const commitCloud = async () => {
    if (!selected || !canUse) return;
    try {
      const segment = await musicServiceAPI.segment({ trackId: selected.id, startSec, endSec: effectiveEnd, country });
      onChange({ source: "cloud", ...selected, provider: segment.provider || selected.provider, startSec: segment.startSec, endSec: segment.endSec });
      await musicServiceAPI.event({ trackId: selected.id, type: "use", country });
      setOpen(false);
    } catch {
      // Keep the picker open so the user can correct the selection.
    }
  };

  const chooseLocal = (file) => {
    if (!file || !file.type.startsWith("audio/")) return;
    const url = URL.createObjectURL(file);
    const next = { source: "local", file, title: file.name, artist: "From device", previewUrl: url, startSec: 0, endSec: 0 };
    setLocalFile(next);
    setSelected(next);
    onChange(next);
    setOpen(false);
  };

  const remove = () => { onChange(null); setSelected(null); setLocalFile(null); };

  return (
    <>
      <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-3">
        {value ? (
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-white/10">
              {value.artworkUrl ? <img src={value.artworkUrl} alt="" className="h-full w-full object-cover" /> : <Music2 size={18} className="text-white/70" />}
            </div>
            <div className="min-w-0 flex-1"><p className="truncate text-sm font-bold text-white">{value.title}</p><p className="truncate text-[11px] text-white/45">{value.artist}{value.source === "cloud" ? " · Licensed library" : " · Device"}</p></div>
            <button type="button" onClick={() => setOpen(true)} disabled={disabled} className="rounded-full bg-white/10 px-3 py-2 text-[11px] font-bold text-white">Change</button>
            <button type="button" onClick={remove} disabled={disabled} className="rounded-full p-2 text-white/45 hover:bg-white/10 hover:text-white" aria-label="Remove music"><X size={16} /></button>
          </div>
        ) : (
          <button type="button" onClick={() => setOpen(true)} disabled={disabled} className="flex w-full items-center gap-3 text-left">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-white text-black"><Music2 size={19} /></div>
            <div className="flex-1"><p className="text-sm font-bold text-white">Add music</p><p className="mt-0.5 text-[11px] text-white/45">Search the Notell licensed library or choose audio from your device</p></div>
            <ChevronDown size={17} className="text-white/35" />
          </button>
        )}
      </div>

      {open && (
        <div className="fixed inset-0 z-[80] flex items-end justify-center bg-black/70 p-0 backdrop-blur-sm sm:items-center sm:p-4" onMouseDown={(event) => { if (event.target === event.currentTarget) setOpen(false); }}>
          <section className="flex max-h-[92vh] w-full max-w-2xl flex-col overflow-hidden rounded-t-[28px] bg-neutral-950 text-white ring-1 ring-white/10 sm:rounded-[28px]">
            <header className="flex items-center justify-between border-b border-white/10 px-4 py-3"><div><h2 className="text-base font-bold">Add sound</h2><p className="text-[11px] text-white/40">Use licensed cloud music or your own audio</p></div><button type="button" onClick={() => setOpen(false)} className="rounded-full p-2 text-white/50 hover:bg-white/10 hover:text-white"><X size={18} /></button></header>
            <div className="grid grid-cols-2 gap-1 border-b border-white/10 p-2"><button type="button" onClick={() => setTab("cloud")} className={`rounded-xl py-2.5 text-xs font-bold ${tab === "cloud" ? "bg-white text-black" : "text-white/50"}`}>Notell music</button><button type="button" onClick={() => setTab("local")} className={`rounded-xl py-2.5 text-xs font-bold ${tab === "local" ? "bg-white text-black" : "text-white/50"}`}>From device</button></div>
            {tab === "cloud" ? (
              <>
                <div className="p-3"><div className="flex items-center gap-2 rounded-2xl bg-white/[0.06] px-3"><Search size={17} className="text-white/35" /><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search songs, artists, albums" className="h-11 min-w-0 flex-1 bg-transparent text-sm outline-none placeholder:text-white/30" /></div></div>
                <div className="min-h-0 flex-1 overflow-y-auto px-3 pb-3">{loading ? <div className="flex justify-center py-10"><Loader2 className="animate-spin text-white/40" /></div> : tracks.length ? <div className="space-y-1">{tracks.map((track) => <button key={track.id} type="button" onClick={() => chooseCloud(track)} className={`flex w-full items-center gap-3 rounded-2xl p-2 text-left hover:bg-white/[0.06] ${selected?.id === track.id ? "bg-white/[0.08]" : ""}`}><div className="h-12 w-12 shrink-0 overflow-hidden rounded-xl bg-white/10">{track.artworkUrl ? <img src={track.artworkUrl} alt="" className="h-full w-full object-cover" /> : <div className="flex h-full items-center justify-center"><Music2 size={17} className="text-white/40" /></div>}</div><div className="min-w-0 flex-1"><p className="truncate text-sm font-semibold">{track.title}</p><p className="truncate text-[11px] text-white/45">{track.artist}{track.album ? ` · ${track.album}` : ""}</p></div><span className="shrink-0 text-[10px] text-white/35">{formatTime(track.durationSec)}</span>{track.canUseInPost && track.rights?.licensed && track.rights?.ugcUse && track.rights?.streaming ? <Check size={15} className="text-emerald-300" /> : <span className="text-[9px] text-white/25">Unavailable</span>}</button>)}</div> : <div className="py-12 text-center text-xs text-white/35">No cleared tracks found.</div>}</div>
                {selected?.source !== "local" && selected && canUse && <div className="border-t border-white/10 bg-neutral-950 p-3"><div className="flex items-center gap-3"><div className="h-12 w-12 overflow-hidden rounded-xl bg-white/10">{selected.artworkUrl ? <img src={selected.artworkUrl} alt="" className="h-full w-full object-cover" /> : <Music2 size={18} className="m-3" />}</div><div className="min-w-0 flex-1"><p className="truncate text-sm font-bold">{selected.title}</p><p className="truncate text-[11px] text-white/45">{selected.artist} · choose a clip</p></div>{previewURL && <audio ref={audioRef} src={previewURL} controls className="h-8 max-w-[150px]" />}</div><div className="mt-3 rounded-2xl bg-white/[0.04] p-3"><div className="flex items-center justify-between text-[10px] text-white/40"><span>Start {formatTime(startSec)}</span><span><Clock3 size={11} className="mr-1 inline" />Up to 60s</span><span>End {formatTime(effectiveEnd)}</span></div><input type="range" min="0" max={Math.max(0, maxDuration - 1)} step="0.1" value={startSec} onChange={(event) => { const next = Math.min(Number(event.target.value), Math.max(0, maxDuration - 1)); setStartSec(next); setEndSec(Math.min(maxDuration, next + Math.min(30, maxDuration - next))); }} className="mt-3 w-full accent-white" /></div><button type="button" onClick={commitCloud} className="mt-3 flex h-11 w-full items-center justify-center gap-2 rounded-full bg-white text-sm font-bold text-black"><Check size={16} /> Use this sound</button></div>}
              </>
            ) : (
              <div className="p-5"><label className="flex min-h-52 cursor-pointer flex-col items-center justify-center rounded-3xl border border-dashed border-white/15 bg-white/[0.03] text-center"><Upload size={26} className="text-white/45" /><p className="mt-3 text-sm font-bold">Choose audio from this device</p><p className="mt-1 text-[11px] text-white/35">MP3 and browser-supported audio files</p><input type="file" accept="audio/*" className="hidden" onChange={(event) => chooseLocal(event.target.files?.[0])} /></label>{localFile && <div className="mt-3 text-xs text-white/50">{localFile.title}</div>}</div>
            )}
          </section>
        </div>
      )}
    </>
  );
};

export default MusicPicker;
