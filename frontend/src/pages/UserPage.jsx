import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ArrowLeft,
  Camera,
  Check,
  ExternalLink,
  Link2,
  Loader2,
  MapPin,
  MoreVertical,
  Play,
  Settings,
  Share2,
  UserPlus,
  UserRound,
  Users as UsersIcon,
  X,
} from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { userAPI } from "../services/user/userApi";
import { getFileUrl } from "../utils/api";

const PROFILE_PAGE_SIZE = 36;
const URL_PATTERN = /https?:\/\/[^\s]+/gi;
const cleanUrl = (value) => value.replace(/[),.;]+$/g, "");
const getLinkLabel = (value) => {
  try {
    const host = new URL(value).hostname.replace(/^www\./i, "");
    const name = host.split(".")[0] || "Link";
    return name.charAt(0).toUpperCase() + name.slice(1);
  } catch {
    return "Link";
  }
};
const extractBioLinks = (bio = "") =>
  [...new Set((bio.match(URL_PATTERN) ?? []).map(cleanUrl))].slice(0, 5);

function Stat({ label, value, icon: Icon, onClick }) {
  return (
    <button type="button" onClick={onClick} className="group flex min-w-0 flex-1 items-center justify-center gap-2 rounded-xl px-2 py-2 text-center transition hover:bg-white/10">
      <Icon size={17} strokeWidth={1.8} className="text-white/75" />
      <span><span className="block text-base font-bold leading-5 text-white">{value ?? 0}</span><span className="block text-[10px] font-medium uppercase tracking-wide text-white/55">{label}</span></span>
    </button>
  );
}

function MediaTile({ post }) {
  const mediaUrl = getFileUrl(post?.contentUrl);
  const isVideo = post?.contentType === "video";
  return (
    <Link to={`/posts/${post?.postId}`} aria-label={`Open ${isVideo ? "video" : "photo"}${post?.caption ? `: ${post.caption}` : ""}`} className="group relative aspect-square overflow-hidden rounded-xl bg-slate-900 ring-1 ring-white/10 transition duration-200 hover:-translate-y-0.5 hover:ring-orange-400/35 focus:outline-none focus-visible:ring-2 focus-visible:ring-orange-400/80 active:scale-[0.99] sm:rounded-2xl">
      {isVideo ? <video src={mediaUrl} muted playsInline preload="metadata" className="h-full w-full object-cover transition duration-300 group-hover:scale-105" /> : <img src={mediaUrl} alt={post?.caption || "Post"} loading="lazy" className="h-full w-full object-cover transition duration-300 group-hover:scale-105" />}
      <div className="absolute inset-0 bg-gradient-to-t from-black/75 via-black/5 to-transparent opacity-0 transition-opacity duration-200 group-hover:opacity-100" />
      {isVideo && <span className="absolute right-2 top-2 inline-flex h-7 w-7 items-center justify-center rounded-full bg-black/65 text-white shadow-lg"><Play size={12} fill="currentColor" /></span>}
      {post?.caption && <span className="absolute bottom-2 left-2 right-2 line-clamp-2 text-[10px] font-medium leading-4 text-white opacity-0 transition-opacity duration-200 group-hover:opacity-100">{post.caption}</span>}
    </Link>
  );
}

function RelationshipPanel({ type, users, loading, error, onClose }) {
  const title = type === "followers" ? "Followers" : "Following";
  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/60 p-0 backdrop-blur-sm sm:items-center sm:p-5" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <section className="flex max-h-[78vh] w-full flex-col overflow-hidden rounded-t-3xl border border-slate-200 bg-white shadow-2xl sm:max-w-md sm:rounded-3xl" role="dialog" aria-modal="true" aria-label={title}>
        <header className="flex items-center justify-between border-b border-slate-100 px-5 py-4"><div><h2 className="text-base font-bold text-slate-950">{title}</h2><p className="mt-0.5 text-[11px] text-slate-500">People connected to this profile</p></div><button type="button" onClick={onClose} aria-label="Close" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-slate-500 transition hover:bg-slate-100"><X size={18} /></button></header>
        {loading ? <div className="flex min-h-52 items-center justify-center text-slate-500"><Loader2 size={22} className="animate-spin" /></div> : error && users.length === 0 ? <div className="p-8 text-center text-sm text-red-600">{error}</div> : users.length === 0 ? <div className="flex min-h-52 flex-col items-center justify-center p-8 text-center"><UserRound size={32} className="text-slate-300" /><p className="mt-3 text-sm font-medium text-slate-600">No {type} yet</p></div> : <div className="overflow-y-auto">{error && <div className="border-b border-red-100 bg-red-50 px-5 py-3 text-xs text-red-600">{error}</div>}<div className="divide-y divide-slate-100">{users.map((person) => { const personAvatar = getFileUrl(person.profilePicture); return <Link key={person.id} to={`/users/${person.id}`} onClick={onClose} className="flex items-center gap-3 px-5 py-3.5 transition hover:bg-slate-50"><div className="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full bg-slate-100 font-semibold text-slate-600">{personAvatar ? <img src={personAvatar} alt={`${person.username} avatar`} className="h-full w-full object-cover" /> : person.username?.charAt(0).toUpperCase()}</div><div className="min-w-0"><p className="truncate text-sm font-semibold text-slate-900">{person.username}</p>{(person.city || person.country) && <p className="truncate text-[11px] text-slate-500">{[person.city, person.country].filter(Boolean).join(", ")}</p>}{person.bio && <p className="truncate text-[11px] text-slate-500">{person.bio}</p>}</div></Link>; })}</div></div>}
      </section>
    </div>
  );
}

function ProfileMenu({ onClose }) {
  const [copied, setCopied] = useState(false);
  const copyProfile = async () => { try { await navigator.clipboard.writeText(window.location.href); setCopied(true); window.setTimeout(() => setCopied(false), 1600); } catch {} };
  const shareProfile = async () => { const url = window.location.href; try { if (navigator.share) await navigator.share({ title: document.title, text: "Check out this Notell profile", url }); else { await navigator.clipboard.writeText(url); setCopied(true); window.setTimeout(() => setCopied(false), 1600); } } catch (error) { if (error?.name !== "AbortError") await copyProfile(); } };
  return <><button type="button" aria-label="Close profile menu" className="fixed inset-0 z-40 cursor-default" onClick={onClose} /><div className="absolute right-3 top-12 z-50 w-56 overflow-hidden rounded-2xl border border-white/10 bg-slate-900/95 p-1.5 shadow-2xl backdrop-blur-xl" role="menu"><button type="button" role="menuitem" onClick={shareProfile} className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left text-xs font-semibold text-white/80 hover:bg-white/10"><Share2 size={15} />Share profile</button><button type="button" role="menuitem" onClick={copyProfile} className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left text-xs font-semibold text-white/80 hover:bg-white/10"><Link2 size={15} />{copied ? "Profile link copied" : "Copy profile link"}</button></div></>;
}

export const Users = () => {
  const { id } = useParams();
  const [user, setUser] = useState(null);
  const [relationship, setRelationship] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);
  const [profilePage, setProfilePage] = useState(1);
  const [profileHasMore, setProfileHasMore] = useState(false);
  const [profileLoadingMore, setProfileLoadingMore] = useState(false);
  const [relationshipPanel, setRelationshipPanel] = useState(null);
  const [relationshipUsers, setRelationshipUsers] = useState([]);
  const [relationshipLoading, setRelationshipLoading] = useState(false);
  const [relationshipError, setRelationshipError] = useState(null);
  const [profileMenuOpen, setProfileMenuOpen] = useState(false);

  const loadProfile = useCallback(async () => {
    if (!id) return;
    setLoading(true); setError(null); setProfilePage(1); setProfileHasMore(false);
    try {
      const [profileResponse, relationshipResponse] = await Promise.all([userAPI.getProfile(id, 1, PROFILE_PAGE_SIZE), userAPI.getRelationship(id)]);
      setUser(profileResponse?.data?.user ?? null);
      setRelationship(relationshipResponse?.data ?? null);
      setProfileHasMore(Boolean(profileResponse?.data?.pagination?.hasMore));
    } catch (err) { setError(err.response?.data?.message || "Failed to load profile"); }
    finally { setLoading(false); }
  }, [id]);

  useEffect(() => { loadProfile(); }, [loadProfile]);

  useEffect(() => {
    const handleRelationshipChange = (event) => {
      const change = event.detail;
      if (!change || String(change.userId) !== String(id)) return;
      setRelationship((current) => ({ ...current, ...change }));
    };
    window.addEventListener("notell:relationship-changed", handleRelationshipChange);
    return () => window.removeEventListener("notell:relationship-changed", handleRelationshipChange);
  }, [id]);

  const loadMorePosts = useCallback(async () => {
    if (!id || profileLoadingMore || !profileHasMore) return;
    const nextPage = profilePage + 1;
    setProfileLoadingMore(true);
    try {
      const response = await userAPI.getProfile(id, nextPage, PROFILE_PAGE_SIZE);
      const incoming = Array.isArray(response?.data?.user?.posts) ? response.data.user.posts : [];
      setUser((current) => {
        if (!current) return current;
        const existing = Array.isArray(current.posts) ? current.posts : [];
        const seen = new Set(existing.map((post) => post.postId));
        const merged = [...existing, ...incoming.filter((post) => post?.postId && !seen.has(post.postId))];
        return { ...current, posts: merged, postCount: response?.data?.user?.postCount ?? current.postCount };
      });
      setProfilePage(nextPage);
      setProfileHasMore(Boolean(response?.data?.pagination?.hasMore));
    } catch (err) { setError(err.response?.data?.message || "Failed to load more posts"); }
    finally { setProfileLoadingMore(false); }
  }, [id, profileHasMore, profileLoadingMore, profilePage]);

  const openRelationshipPanel = async (type) => {
    if (!id) return;
    setRelationshipPanel(type); setRelationshipUsers([]); setRelationshipError(null); setRelationshipLoading(true);
    try {
      const response = type === "followers" ? await userAPI.getFollowers(id, 1, 100) : await userAPI.getFollowing(id, 1, 100);
      setRelationshipUsers(Array.isArray(response?.data) ? response.data : []);
    } catch (err) { setRelationshipError(err.response?.data?.message || `Failed to load ${type}`); }
    finally { setRelationshipLoading(false); }
  };

  const handleFollow = async () => {
    if (!id || actionLoading || (!relationship?.following && !relationship?.allowFollowers)) return;
    setActionLoading(true); setError(null);
    try {
      const response = relationship.following
        ? await userAPI.unfollowUser(id)
        : await userAPI.followUser(id);
      const data = response?.data?.data ?? response?.data ?? {};
      setRelationship((current) => ({
        ...current,
        ...data,
        following: data.following ?? !current?.following,
      }));
    } catch (err) { setError(err.response?.data?.message || "Failed to update follow status"); }
    finally { setActionLoading(false); }
  };

  const avatar = getFileUrl(user?.profilePicture);
  const cover = getFileUrl(user?.coverPicture);
  const posts = Array.isArray(user?.posts) ? user.posts : [];
  const postCount = user?.postCount ?? posts.length;
  const bioLinks = useMemo(() => extractBioLinks(user?.bio), [user?.bio]);
  const joined = user?.createdAt ? new Date(user.createdAt).toLocaleDateString(undefined, { month: "short", year: "numeric" }) : "";
  const isSelf = Boolean(relationship?.isSelf);
  const canFollow = Boolean(relationship?.following || relationship?.allowFollowers);

  if (loading) return <div className="flex min-h-[60vh] items-center justify-center bg-slate-950 text-white"><Loader2 className="animate-spin" size={24} /></div>;
  if (error && !user) return <div className="mx-auto max-w-xl p-6"><Link to="/" className="mb-5 inline-flex items-center gap-2 text-sm font-medium text-white/70"><ArrowLeft size={17} />Back to home</Link><div className="rounded-3xl border border-red-500/20 bg-slate-900 p-8 text-center text-red-300">{error}</div></div>;

  return <div className="min-h-screen w-full bg-slate-950 pb-24 text-white lg:pb-8"><div className="mx-auto w-full max-w-6xl">
    <div className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-white/10 bg-slate-950/85 px-4 backdrop-blur-xl sm:px-6">
      <Link to="/" aria-label="Back to home" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-white/80 hover:bg-white/10"><ArrowLeft size={20} /></Link>
      <div className="min-w-0 text-center"><span className="truncate text-sm font-semibold text-white">{user?.username}</span>{user?.status && user.status !== "free" && <span className="ml-2 rounded-full border border-orange-300/40 bg-orange-400/10 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide text-orange-300">{user.status}</span>}</div>
      {isSelf ? <Link to="/settings" aria-label="Profile settings" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-white/80 hover:bg-white/10"><Settings size={19} /></Link> : <button type="button" onClick={() => setProfileMenuOpen((v) => !v)} aria-label="Profile menu" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-white/70 hover:bg-white/10"><MoreVertical size={20} /></button>}
      {profileMenuOpen && <ProfileMenu onClose={() => setProfileMenuOpen(false)} />}
    </div>

    <section className="relative overflow-hidden border-b border-white/10 bg-black lg:rounded-b-[32px] lg:border-x lg:border-white/10">
      <div className="relative h-36 overflow-hidden bg-gradient-to-br from-slate-800 via-slate-700 to-slate-950 sm:h-48 lg:h-64">{cover && <img src={cover} alt={`${user?.username} cover`} className="h-full w-full object-cover" />}<div className="absolute inset-0 bg-gradient-to-b from-black/5 via-black/10 to-black/90" /><div className="absolute inset-x-0 bottom-0 h-24 bg-gradient-to-t from-black to-transparent" /></div>
      <div className="relative mx-auto max-w-4xl px-4 pb-4 sm:px-6 lg:px-8"><div className="-mt-12 flex flex-col items-center sm:-mt-14 lg:-mt-16">
        <div className="relative"><div className="flex h-24 w-24 items-center justify-center overflow-hidden rounded-full border-4 border-black bg-slate-800 text-3xl font-bold shadow-2xl sm:h-28 sm:w-28">{avatar ? <img src={avatar} alt={`${user?.username} avatar`} className="h-full w-full object-cover" /> : user?.username?.charAt(0).toUpperCase()}</div>{isSelf && <Link to="/profile" aria-label="Edit profile picture" className="absolute bottom-0 right-0 inline-flex h-8 w-8 items-center justify-center rounded-full border-2 border-black bg-slate-800 hover:bg-orange-500"><Camera size={14} /></Link>}</div>
        <div className="mt-2.5 text-center"><div className="flex flex-wrap items-center justify-center gap-2"><h1 className="text-xl font-bold tracking-tight sm:text-2xl">{user?.username}</h1>{user?.status && user.status !== "free" && <span className="rounded-full border border-orange-400/60 bg-orange-400/10 px-2 py-1 text-[9px] font-bold uppercase tracking-widest text-orange-300">{user.status}</span>}</div>{joined && <p className="mt-0.5 text-[10px] text-white/50">Member since {joined}</p>}{(user?.city || user?.country) && <div className="mt-1.5 inline-flex items-center gap-1.5 text-[10px] text-white/55"><MapPin size={12} />{[user.city, user.country].filter(Boolean).join(", ")}</div>}{user?.bio && <p className="mx-auto mt-2 max-w-xl whitespace-pre-line text-xs leading-5 text-white/70">{user.bio}</p>}{bioLinks.length > 0 && <div className="mt-2 flex max-w-xl flex-wrap justify-center gap-1.5">{bioLinks.map((url) => <a key={url} href={url} target="_blank" rel="noreferrer noopener" className="inline-flex min-h-8 max-w-full items-center gap-1.5 rounded-full border border-white/10 bg-white/[0.06] px-3 py-1 text-[10px] font-semibold text-white/80 hover:border-orange-400/40"><Link2 size={12} className="text-orange-300" /><span className="max-w-[9rem] truncate">{getLinkLabel(url)}</span><ExternalLink size={10} className="text-white/45" /></a>)}</div>}</div>
        {!isSelf && <div className="mt-3 flex w-full max-w-lg items-stretch"><button type="button" onClick={handleFollow} disabled={actionLoading || !canFollow} aria-pressed={Boolean(relationship?.following)} className={`group relative inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-xl px-5 text-sm font-bold shadow-lg disabled:cursor-not-allowed disabled:opacity-50 sm:min-h-11 sm:text-[14px] ${relationship?.following ? "border border-white/20 bg-white/10 text-white" : "bg-orange-500 text-white hover:bg-orange-400"}`}>{actionLoading ? <Loader2 size={17} className="animate-spin" /> : relationship?.following ? <Check size={17} /> : <UserPlus size={17} />}<span>{actionLoading ? "Updating…" : relationship?.following ? "Following" : canFollow ? "Follow" : "Followers disabled"}</span></button></div>}
      </div>
      <div className="mx-auto mt-3 flex w-full max-w-xl items-center rounded-2xl border border-white/10 bg-white/[0.04] p-1"><Stat label="Posts" value={postCount} icon={Camera} /><div className="h-8 w-px bg-white/10" /><Stat label="Followers" value={relationship?.followerCount} icon={UsersIcon} onClick={() => void openRelationshipPanel("followers")} /><div className="h-8 w-px bg-white/10" /><Stat label="Following" value={relationship?.followingCount} icon={UserRound} onClick={() => void openRelationshipPanel("following")} /></div>
      {relationship?.follower && !relationship?.following && <div className="mt-2 text-center text-[10px] text-white/45"><UsersIcon size={12} className="mr-1 inline" />This user follows you</div>}{error && <p className="mt-2 text-center text-xs text-red-400">{error}</p>}
      </div>
    </section>

    <section className="mx-auto mt-3 max-w-6xl px-3 sm:px-6"><div className="overflow-hidden rounded-3xl border border-white/10 bg-white/[0.035] shadow-2xl shadow-black/20"><div className="flex items-center justify-between border-b border-white/10 px-3 py-3 sm:px-4"><div><div className="flex items-center gap-2"><h2 className="text-sm font-bold">Posts</h2>{postCount > 0 && <span className="rounded-full bg-white/10 px-2 py-0.5 text-[9px] font-semibold text-white/55">{postCount}</span>}</div><p className="mt-0.5 text-[10px] text-white/40">Photos and videos from {user?.username}</p></div><button type="button" aria-label="Post view options" className="inline-flex h-8 w-8 items-center justify-center rounded-full text-white/45 hover:bg-white/10"><MoreVertical size={16} className="rotate-90" /></button></div>
      {posts.length === 0 ? <div className="flex min-h-64 flex-col items-center justify-center px-6 py-12 text-center sm:min-h-72"><div className="relative flex h-16 w-16 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.05] text-white/35"><Camera size={25} /><span className="absolute -right-1.5 -top-1.5 flex h-6 w-6 items-center justify-center rounded-full border border-slate-950 bg-orange-500 text-white"><Play size={10} fill="currentColor" /></span></div><h3 className="mt-4 text-sm font-semibold text-white/85">Nothing here yet</h3><p className="mt-1 max-w-xs text-xs leading-5 text-white/40">When {user?.username} shares photos or videos, they’ll appear here.</p>{isSelf && <Link to="/create-post" className="mt-5 inline-flex min-h-10 items-center gap-2 rounded-xl bg-orange-500 px-4 text-xs font-bold text-white hover:bg-orange-400"><Camera size={14} />Create your first post</Link>}</div> : <><div className="grid grid-cols-2 gap-1.5 p-2 sm:grid-cols-3 sm:gap-2.5 sm:p-3 lg:grid-cols-4 lg:gap-3">{posts.map((post) => <MediaTile key={post.postId} post={post} />)}</div>{profileHasMore && <div className="flex justify-center border-t border-white/10 p-4"><button type="button" onClick={() => void loadMorePosts()} disabled={profileLoadingMore} className="inline-flex min-h-10 items-center gap-2 rounded-xl border border-white/10 bg-white/[0.05] px-5 text-xs font-bold text-white/80 transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-60">{profileLoadingMore && <Loader2 size={15} className="animate-spin" />}{profileLoadingMore ? "Loading more…" : "Load more posts"}</button></div>}</>}
    </div></section>
  </div>
  {relationshipPanel && <RelationshipPanel type={relationshipPanel} users={relationshipUsers} loading={relationshipLoading} error={relationshipError} onClose={() => setRelationshipPanel(null)} />}
  </div>;
};

export default Users;
