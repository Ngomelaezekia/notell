import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, CalendarDays, Loader2, Play, UserPlus, Users } from "lucide-react";
import { channelAPI } from "../services/channel/channelApi";
import LiveStreamViewer from "../components/LiveStreamViewer";
import { getApiErrorMessage, getFileUrl } from "../utils/api";

export default function ChannelDetailPage() {
  const { id } = useParams();
  const [channel, setChannel] = useState(null);
  const [programs, setPrograms] = useState([]);
  const [content, setContent] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [following, setFollowing] = useState(false);

  useEffect(() => {
    let cancelled = false;
    Promise.all([channelAPI.get(id), channelAPI.programs(id), channelAPI.content(id)])
      .then(([c, p, x]) => {
        if (cancelled) return;
        setChannel(c);
        setPrograms(p.programs || []);
        setContent(x.content || []);
      })
      .catch((e) => !cancelled && setError(getApiErrorMessage(e, "Unable to load channel.")))
      .finally(() => !cancelled && setLoading(false));
    return () => { cancelled = true; };
  }, [id]);

  const follow = async () => {
    try {
      if (following) await channelAPI.unfollow(id);
      else await channelAPI.follow(id);
      setFollowing((value) => !value);
    } catch (e) {
      setError(getApiErrorMessage(e, "Could not update channel follow."));
    }
  };

  if (loading) return <div className="flex min-h-screen w-full items-center justify-center bg-neutral-950 text-neutral-600"><Loader2 className="animate-spin" /></div>;
  if (!channel) return <div className="min-h-screen w-full bg-neutral-950 p-6 text-neutral-300">{error || "Channel not found."}</div>;

  return (
    <section className="min-h-screen w-full bg-neutral-950 pb-28 text-neutral-100">
      <div className="relative h-48 bg-neutral-900 sm:h-64">
        {channel.bannerMediaId && <img src={getFileUrl(channel.bannerMediaId)} alt="" className="h-full w-full object-cover opacity-70" />}
        <div className="absolute inset-0 bg-gradient-to-t from-neutral-950 via-neutral-950/20 to-transparent" />
        <Link to="/channels" className="absolute left-4 top-4 rounded-full bg-black/50 p-2 text-white backdrop-blur"><ArrowLeft size={18} /></Link>
      </div>
      <div className="mx-auto max-w-5xl px-4">
        <div className="-mt-12 flex flex-wrap items-end justify-between gap-4">
          <div className="flex items-end gap-4">
            <div className="h-24 w-24 overflow-hidden rounded-3xl border-4 border-neutral-950 bg-neutral-800">
              {channel.avatarMediaId ? <img src={getFileUrl(channel.avatarMediaId)} alt="" className="h-full w-full object-cover" /> : <div className="flex h-full items-center justify-center text-3xl font-black">{channel.name[0]}</div>}
            </div>
            <div className="pb-1"><h1 className="text-2xl font-black">{channel.name}</h1><p className="text-xs text-neutral-500">{channel.category || "General"} · {channel.type}</p></div>
          </div>
          <button onClick={follow} className="inline-flex h-10 items-center gap-2 rounded-xl border border-neutral-700 bg-neutral-900 px-4 text-xs font-black hover:bg-neutral-800"><UserPlus size={15} />{following ? "Following" : "Follow"}</button>
        </div>
        <p className="mt-5 max-w-3xl text-sm leading-6 text-neutral-400">{channel.description || "This channel has not added a description yet."}</p>
        {error && <div className="mt-4 rounded-xl bg-red-950/30 p-3 text-xs text-red-300">{error}</div>}

        <LiveStreamViewer channelId={id} />

        <section className="mt-8">
          <div className="mb-3 flex items-center gap-2"><CalendarDays size={17} /><h2 className="font-black">Programs</h2></div>
          {programs.length ? <div className="grid gap-3 sm:grid-cols-2">{programs.filter((p) => p.status === "PUBLISHED").map((p) => <div key={p.id} className="rounded-2xl border border-neutral-800 bg-neutral-900/60 p-4"><h3 className="font-bold">{p.title}</h3><p className="mt-1 text-xs leading-5 text-neutral-500">{p.description || "Program from this channel."}</p></div>)}</div> : <p className="text-sm text-neutral-600">No published programs yet.</p>}
        </section>

        <section className="mt-8">
          <div className="mb-3 flex items-center gap-2"><Play size={17} /><h2 className="font-black">Channel content</h2></div>
          {content.length ? <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">{content.map((item) => <Link key={item.id} to={`/channels/${id}/player/${item.postId}`} className="group overflow-hidden rounded-2xl border border-neutral-800 bg-neutral-900"><div className="flex aspect-video items-center justify-center bg-neutral-800"><Play size={28} className="text-neutral-500 transition group-hover:text-white" /></div><div className="p-3"><p className="text-xs font-bold">Watch channel content</p><p className="mt-1 flex items-center gap-1 text-[10px] text-neutral-600"><Users size={11} /> Post #{item.postId}</p></div></Link>)}</div> : <p className="text-sm text-neutral-600">No channel content yet.</p>}
        </section>
      </div>
    </section>
  );
}
