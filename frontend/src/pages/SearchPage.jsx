import { useEffect, useRef, useState } from "react";
import { ArrowLeft, Check, Loader2, MapPin, Search, UserPlus, UserRound, X } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { userAPI } from "../services/user/userApi";
import { getApiErrorMessage, getFileUrl } from "../utils/api";

const PAGE_SIZE = 20;
const SUGGESTION_LIMIT = 5;
const MIN_QUERY_LENGTH = 2;

const Avatar = ({ user, size = "h-12 w-12" }) => (
  <div className={`${size} shrink-0 overflow-hidden rounded-full bg-neutral-800 ring-1 ring-neutral-700`}>
    {user?.profilePicture ? <img src={getFileUrl(user.profilePicture)} alt="" className="h-full w-full object-cover" loading="lazy" /> : <div className="flex h-full w-full items-center justify-center text-neutral-500"><UserRound size={20} /></div>}
  </div>
);

const PersonRow = ({ user, following, loading, onFollow }) => {
  const location = [user?.city, user?.country].filter(Boolean).join(", ");
  return (
    <div className="group flex items-center gap-3 rounded-2xl border border-neutral-800/80 bg-neutral-900/70 p-3 transition hover:border-neutral-700 hover:bg-neutral-900">
      <Link to={`/users/${user.id}`} className="shrink-0"><Avatar user={user} /></Link>
      <Link to={`/users/${user.id}`} className="min-w-0 flex-1">
        <p className="truncate text-sm font-semibold text-neutral-100 group-hover:text-white">{user.username}</p>
        <div className="mt-1 flex min-w-0 items-center gap-1.5 text-[11px] text-neutral-500">
          {location && <><MapPin size={12} className="shrink-0" /><span className="truncate">{location}</span></>}
          {user?.bio && <span className="truncate">{location ? "· " : ""}{user.bio}</span>}
        </div>
      </Link>
      <button type="button" onClick={() => onFollow(user)} disabled={loading} className={`inline-flex h-9 min-w-[84px] shrink-0 items-center justify-center gap-1.5 rounded-xl px-3 text-xs font-bold transition active:scale-95 disabled:cursor-not-allowed disabled:opacity-60 ${following ? "border border-neutral-700 bg-neutral-800 text-neutral-200" : "bg-neutral-100 text-neutral-950 hover:bg-white"}`}>
        {loading ? <Loader2 size={14} className="animate-spin" /> : following ? <Check size={14} /> : <UserPlus size={14} />}{following ? "Following" : "Follow"}
      </button>
    </div>
  );
};

export default function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [query, setQuery] = useState(searchParams.get("q") || "");
  const [users, setUsers] = useState([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState(null);
  const [suggestions, setSuggestions] = useState([]);
  const [suggestionLoading, setSuggestionLoading] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [followState, setFollowState] = useState({});
  const [followLoading, setFollowLoading] = useState({});
  const timerRef = useRef(null);
  const requestRef = useRef(0);

  const onSearching = (value) => {
    setQuery(value);
    const trimmed = value.trim();
    setShowSuggestions(Boolean(trimmed));
    if (timerRef.current) clearTimeout(timerRef.current);
    if (trimmed.length < MIN_QUERY_LENGTH) { setSuggestions([]); setSuggestionLoading(false); return; }
    const requestId = ++requestRef.current;
    setSuggestionLoading(true);
    timerRef.current = setTimeout(async () => {
      try { const result = await userAPI.searchUsers(trimmed, 1, SUGGESTION_LIMIT); if (requestId === requestRef.current) setSuggestions(result?.data?.users || []); }
      catch { if (requestId === requestRef.current) setSuggestions([]); }
      finally { if (requestId === requestRef.current) setSuggestionLoading(false); }
    }, 250);
  };

  useEffect(() => () => timerRef.current && clearTimeout(timerRef.current), []);

  useEffect(() => {
    const handleRelationshipChange = (event) => {
      const id = String(event.detail?.userId ?? ""); if (!id) return;
      const following = Boolean(event.detail?.following);
      setFollowState((current) => ({ ...current, [id]: following }));
      setUsers((current) => current.map((u) => String(u.id) === id ? { ...u, following } : u));
      setSuggestions((current) => current.map((u) => String(u.id) === id ? { ...u, following } : u));
    };
    window.addEventListener("notell:relationship-changed", handleRelationshipChange);
    return () => window.removeEventListener("notell:relationship-changed", handleRelationshipChange);
  }, []);

  useEffect(() => {
    const value = searchParams.get("q")?.trim() || "";
    setQuery(value); setShowSuggestions(false);
    if (value.length < MIN_QUERY_LENGTH) { setUsers([]); setPage(1); setHasMore(false); setLoading(false); setError(null); return; }
    let cancelled = false;
    setLoading(true); setError(null); setUsers([]); setPage(1); setHasMore(false);
    userAPI.searchUsers(value, 1, PAGE_SIZE).then((result) => {
      if (cancelled) return;
      const data = result?.data || {}; const resultUsers = data.users || [];
      setUsers(resultUsers);
      setFollowState(Object.fromEntries(resultUsers.filter((u) => u?.id).map((u) => [String(u.id), Boolean(u.following)])));
      setHasMore(Boolean(data.pagination?.hasMore));
    }).catch((err) => { if (!cancelled) setError(getApiErrorMessage(err, "People search failed.")); }).finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [searchParams]);

  const handleFollow = async (user) => {
    const id = String(user?.id ?? ""); if (!id || followLoading[id]) return;
    const following = Boolean(followState[id] ?? user.following);
    setFollowLoading((current) => ({ ...current, [id]: true }));
    try {
      const response = following ? await userAPI.unfollowUser(user.id) : await userAPI.followUser(user.id);
      const next = Boolean(response?.data?.data?.following ?? response?.data?.following ?? !following);
      setFollowState((current) => ({ ...current, [id]: next }));
      setUsers((current) => current.map((u) => String(u.id) === id ? { ...u, following: next } : u));
    } catch (err) { setError(getApiErrorMessage(err, `Could not ${following ? "unfollow" : "follow"} @${user.username}.`)); }
    finally { setFollowLoading((current) => ({ ...current, [id]: false })); }
  };

  const submitSearch = (event) => { event.preventDefault(); const value = query.trim(); setShowSuggestions(false); if (value.length < MIN_QUERY_LENGTH) return setSearchParams({}); setSearchParams({ q: value }); };
  const clearSearch = () => { setQuery(""); setSuggestions([]); setShowSuggestions(false); setSearchParams({}); };
  const loadMore = async () => {
    if (loadingMore || !hasMore) return;
    const nextPage = page + 1; setLoadingMore(true); setError(null);
    try {
      const data = (await userAPI.searchUsers(query.trim(), nextPage, PAGE_SIZE))?.data || {};
      const incoming = data.users || [];
      setUsers((current) => { const ids = new Set(current.map((u) => u.id)); return [...current, ...incoming.filter((u) => !ids.has(u.id))]; });
      setFollowState((current) => ({ ...current, ...Object.fromEntries(incoming.map((u) => [String(u.id), Boolean(u.following)])) }));
      setPage(data.pagination?.page ?? nextPage); setHasMore(Boolean(data.pagination?.hasMore));
    } catch (err) { setError(getApiErrorMessage(err, "Failed to load more people.")); } finally { setLoadingMore(false); }
  };

  return (
    <section className="min-h-[calc(100dvh-4rem)] w-full bg-neutral-950 px-3 pb-28 pt-4 text-neutral-100 sm:px-5 sm:pb-12 sm:pt-6">
      <div className="mx-auto w-full max-w-3xl">
        <header className="mb-5 flex items-center gap-3"><Link to="/" aria-label="Back to feed" className="rounded-full p-2 text-neutral-400 transition hover:bg-neutral-900 hover:text-white"><ArrowLeft size={19} /></Link><div><h1 className="text-xl font-bold tracking-tight">Find people</h1><p className="mt-0.5 text-xs text-neutral-500 sm:text-sm">Search Notell accounts by username and profile details.</p></div></header>
        <form onSubmit={submitSearch} className="relative">
          <Search className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-neutral-500" size={19} />
          <input value={query} onChange={(e) => onSearching(e.target.value)} onFocus={() => query.trim() && setShowSuggestions(true)} onBlur={() => setTimeout(() => setShowSuggestions(false), 150)} placeholder="Search people" maxLength={100} autoComplete="off" enterKeyHint="search" aria-label="Search people" className="w-full rounded-2xl border border-neutral-800 bg-neutral-900 py-3.5 pl-11 pr-24 text-sm text-white outline-none transition placeholder:text-neutral-600 focus:border-neutral-600 focus:ring-2 focus:ring-white/5" />
          {query && <button type="button" onMouseDown={(e) => e.preventDefault()} onClick={clearSearch} aria-label="Clear search" className="absolute right-20 top-1/2 -translate-y-1/2 rounded-full p-1.5 text-neutral-500 hover:bg-neutral-800 hover:text-white"><X size={16} /></button>}
          <button type="submit" className="absolute right-2 top-1/2 -translate-y-1/2 rounded-xl bg-neutral-100 px-3.5 py-2 text-xs font-bold text-neutral-950 hover:bg-white">Search</button>
          {showSuggestions && query.trim().length >= MIN_QUERY_LENGTH && <div className="absolute inset-x-0 top-[calc(100%+8px)] z-40 overflow-hidden rounded-2xl border border-neutral-800 bg-neutral-950/98 p-1.5 shadow-2xl shadow-black/40 backdrop-blur-xl">{suggestionLoading ? <div className="flex justify-center p-5 text-neutral-500"><Loader2 size={18} className="animate-spin" /></div> : suggestions.length ? suggestions.map((user) => <Link key={user.id} to={`/users/${user.id}`} className="flex items-center gap-3 rounded-xl p-2.5 hover:bg-neutral-900"><Avatar user={user} size="h-9 w-9" /><div className="min-w-0"><p className="truncate text-sm font-semibold">{user.username}</p><p className="text-[11px] text-neutral-500">View profile</p></div></Link>) : <p className="p-4 text-center text-xs text-neutral-500">No people found</p>}</div>}
        </form>
        {error && <div className="mt-4 rounded-2xl border border-red-900/50 bg-red-950/30 px-4 py-3 text-xs text-red-300">{error}</div>}
        <div className="mt-6 flex items-center justify-between"><div><h2 className="text-sm font-bold text-neutral-200">People</h2><p className="text-[11px] text-neutral-500">{query.trim() ? `Results for “${query.trim()}”` : "Start with a name or username"}</p></div>{users.length > 0 && <span className="rounded-full border border-neutral-800 bg-neutral-900 px-2.5 py-1 text-[10px] font-semibold text-neutral-500">{users.length}{hasMore ? "+" : ""}</span>}</div>
        <div className="mt-3 space-y-2.5">
          {loading ? Array.from({ length: 5 }).map((_, i) => <div key={i} className="flex animate-pulse items-center gap-3 rounded-2xl border border-neutral-800 bg-neutral-900/60 p-3"><div className="h-12 w-12 rounded-full bg-neutral-800" /><div className="flex-1 space-y-2"><div className="h-3 w-32 rounded bg-neutral-800" /><div className="h-2.5 w-48 rounded bg-neutral-800" /></div><div className="h-9 w-20 rounded-xl bg-neutral-800" /></div>) : users.length ? users.map((user) => <PersonRow key={user.id} user={user} following={Boolean(followState[String(user.id)] ?? user.following)} loading={Boolean(followLoading[String(user.id)])} onFollow={handleFollow} />) : <div className="rounded-3xl border border-dashed border-neutral-800 bg-neutral-900/40 px-6 py-16 text-center"><div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl border border-neutral-800 bg-neutral-900 text-neutral-600"><UserRound size={24} /></div><p className="mt-4 text-sm font-semibold text-neutral-300">{query.trim() ? "No people found" : "Search for a person"}</p><p className="mx-auto mt-1 max-w-sm text-xs leading-5 text-neutral-500">Search only looks through Notell accounts now. Post content is no longer part of search.</p></div>}
        </div>
        {hasMore && <div className="flex justify-center py-5"><button type="button" onClick={() => void loadMore()} disabled={loadingMore} className="inline-flex h-10 items-center gap-2 rounded-xl border border-neutral-800 bg-neutral-900 px-4 text-xs font-bold text-neutral-300 hover:bg-neutral-800 disabled:opacity-50">{loadingMore && <Loader2 size={15} className="animate-spin" />}{loadingMore ? "Loading…" : "Load more people"}</button></div>}
      </div>
    </section>
  );
}
