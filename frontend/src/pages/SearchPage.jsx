import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { ArrowLeft, ChevronDown, FileText, Loader2, MapPin, Search, UserRound, UserPlus, Check, X } from "lucide-react";
import { userAPI } from "../services/user/userApi";
import { postsAPI } from "../services/post/postsApi";
import { getApiErrorMessage, getFileUrl } from "../utils/api";
import { PostCard } from "../components/PostCard";

const PAGE_SIZE = 20;
const SUGGESTION_LIMIT = 5;
const MIN_QUERY_LENGTH = 2;
const TABS = [
  { id: "all", label: "All" },
  { id: "people", label: "People" },
  { id: "posts", label: "Posts" },
];

const formatRelativeTime = (value) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const seconds = Math.max(0, Math.floor((Date.now() - date.getTime()) / 1000));
  if (seconds < 60) return "now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d`;
  const weeks = Math.floor(days / 7);
  if (weeks < 5) return `${weeks}w`;
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
};

const Avatar = ({ user, size = "h-12 w-12" }) => (
  <div className={`${size} shrink-0 overflow-hidden rounded-full bg-neutral-800 ring-1 ring-neutral-700`}>
    {user?.profilePicture ? (
      <img src={getFileUrl(user.profilePicture)} alt="" className="h-full w-full object-cover" />
    ) : (
      <div className="flex h-full w-full items-center justify-center text-neutral-500"><UserRound size={size.includes("9") ? 16 : 21} /></div>
    )}
  </div>
);

const PeopleSkeleton = () => (
  <div className="flex animate-pulse items-center gap-3 rounded-2xl border border-neutral-800 bg-neutral-900/60 p-3">
    <div className="h-12 w-12 shrink-0 rounded-full bg-neutral-800" />
    <div className="min-w-0 flex-1 space-y-2"><div className="h-3 w-32 rounded bg-neutral-800" /><div className="h-2.5 w-48 max-w-full rounded bg-neutral-800" /></div>
    <div className="h-9 w-16 rounded-xl bg-neutral-800" />
  </div>
);

const PostSkeleton = () => (
  <div className="animate-pulse overflow-hidden rounded-2xl border border-neutral-800 bg-neutral-900/60">
    <div className="flex items-center gap-3 p-3"><div className="h-10 w-10 rounded-full bg-neutral-800" /><div className="flex-1 space-y-2"><div className="h-3 w-28 rounded bg-neutral-800" /><div className="h-2.5 w-16 rounded bg-neutral-800" /></div></div>
    <div className="aspect-[4/3] bg-neutral-800/80" /><div className="space-y-2 p-3"><div className="h-3 w-3/4 rounded bg-neutral-800" /><div className="h-2.5 w-1/3 rounded bg-neutral-800" /></div>
  </div>
);

export default function SearchPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [query, setQuery] = useState(searchParams.get("q") || "");
  const [activeTab, setActiveTab] = useState(searchParams.get("type") || "all");
  const [users, setUsers] = useState([]);
  const [posts, setPosts] = useState([]);
  const [userPage, setUserPage] = useState(1);
  const [postPage, setPostPage] = useState(1);
  const [userHasMore, setUserHasMore] = useState(false);
  const [postHasMore, setPostHasMore] = useState(false);
  const [loading, setLoading] = useState(false);
  const [loadingMoreUsers, setLoadingMoreUsers] = useState(false);
  const [loadingMorePosts, setLoadingMorePosts] = useState(false);
  const [userError, setUserError] = useState(null);
  const [postError, setPostError] = useState(null);
  const [searched, setSearched] = useState(Boolean(searchParams.get("q")));
  const [suggestions, setSuggestions] = useState({ users: [], posts: [] });
  const [suggestionLoading, setSuggestionLoading] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [followState, setFollowState] = useState({});
  const [followLoading, setFollowLoading] = useState({});
  const searchTimerRef = useRef(null);
  const requestIdRef = useRef(0);

  const onSearching = (value) => {
    const nextQuery = value.trim();
    setQuery(value);
    setShowSuggestions(Boolean(nextQuery));
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current);
    if (nextQuery.length < MIN_QUERY_LENGTH) {
      setSuggestions({ users: [], posts: [] });
      setSuggestionLoading(false);
      return;
    }
    setSuggestionLoading(true);
    const requestId = ++requestIdRef.current;
    searchTimerRef.current = setTimeout(async () => {
      try {
        const [userResult, postResult] = await Promise.allSettled([
          userAPI.searchUsers(nextQuery, 1, SUGGESTION_LIMIT),
          postsAPI.searchPosts(nextQuery, 1, SUGGESTION_LIMIT),
        ]);
        if (requestId !== requestIdRef.current) return;
        setSuggestions({
          users: userResult.status === "fulfilled" ? userResult.value?.data?.users || [] : [],
          posts: postResult.status === "fulfilled" ? postResult.value?.data?.posts || [] : [],
        });
      } finally {
        if (requestId === requestIdRef.current) setSuggestionLoading(false);
      }
    }, 300);
  };

  useEffect(() => () => searchTimerRef.current && clearTimeout(searchTimerRef.current), []);

  useEffect(() => {
    const value = searchParams.get("q")?.trim() || "";
    const requestedType = searchParams.get("type") || "all";
    const type = TABS.some((tab) => tab.id === requestedType) ? requestedType : "all";
    setQuery(value);
    setActiveTab(type);
    setShowSuggestions(false);
    setFollowState({});
    setFollowLoading({});
    if (value.length < MIN_QUERY_LENGTH) {
      setUsers([]); setPosts([]); setUserPage(1); setPostPage(1);
      setUserHasMore(false); setPostHasMore(false); setSearched(false);
      setUserError(null); setPostError(null);
      return;
    }
    let cancelled = false;
    setLoading(true); setUsers([]); setPosts([]); setUserPage(1); setPostPage(1);
    setUserHasMore(false); setPostHasMore(false); setUserError(null); setPostError(null);
    Promise.allSettled([
      userAPI.searchUsers(value, 1, PAGE_SIZE),
      postsAPI.searchPosts(value, 1, PAGE_SIZE),
    ]).then(([userResult, postResult]) => {
      if (cancelled) return;
      if (userResult.status === "fulfilled") {
        const data = userResult.value?.data || {};
        setUsers(data.users || []); setUserHasMore(Boolean(data.pagination?.hasMore));
      } else {
        setUsers([]); setUserHasMore(false); setUserError(getApiErrorMessage(userResult.reason, "People search failed."));
      }
      if (postResult.status === "fulfilled") {
        const data = postResult.value?.data || {};
        setPosts(data.posts || []); setPostHasMore(Boolean(data.pagination?.hasMore));
      } else {
        setPosts([]); setPostHasMore(false); setPostError(getApiErrorMessage(postResult.reason, "Post search failed."));
      }
      setSearched(true);
    }).finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [searchParams]);

  const handleFollow = async (user) => {
    if (!user?.id || followLoading[user.id]) return;
    setFollowLoading((current) => ({ ...current, [user.id]: true }));
    try {
      await userAPI.followUser(user.id);
      setFollowState((current) => ({ ...current, [user.id]: true }));
    } catch (error) {
      if (error?.response?.status === 409) {
        setFollowState((current) => ({ ...current, [user.id]: true }));
      } else {
        setUserError(getApiErrorMessage(error, `Could not follow @${user.username}.`));
      }
    } finally {
      setFollowLoading((current) => ({ ...current, [user.id]: false }));
    }
  };

  const submitSearch = (event) => {
    event.preventDefault();
    const value = query.trim();
    setShowSuggestions(false);
    if (value.length < MIN_QUERY_LENGTH) return setSearchParams({});
    const nextParams = { q: value };
    if (activeTab !== "all") nextParams.type = activeTab;
    setSearchParams(nextParams);
  };

  const changeTab = (tab) => {
    setActiveTab(tab); setShowSuggestions(false);
    const value = query.trim();
    const nextParams = value.length >= MIN_QUERY_LENGTH ? { q: value } : {};
    if (tab !== "all" && value.length >= MIN_QUERY_LENGTH) nextParams.type = tab;
    setSearchParams(nextParams);
  };

  const clearSearch = () => {
    setQuery(""); setSuggestions({ users: [], posts: [] }); setShowSuggestions(false); setSearchParams({});
  };

  const loadMoreUsers = async () => {
    if (loadingMoreUsers || !userHasMore) return;
    const nextPage = userPage + 1; setLoadingMoreUsers(true); setUserError(null);
    try {
      const data = (await userAPI.searchUsers(query.trim(), nextPage, PAGE_SIZE))?.data || {};
      setUsers((current) => { const ids = new Set(current.map((u) => u.id)); return [...current, ...(data.users || []).filter((u) => !ids.has(u.id))]; });
      setUserPage(data.pagination?.page ?? nextPage); setUserHasMore(Boolean(data.pagination?.hasMore));
    } catch (error) { setUserError(getApiErrorMessage(error, "Failed to load more people.")); }
    finally { setLoadingMoreUsers(false); }
  };

  const loadMorePosts = async () => {
    if (loadingMorePosts || !postHasMore) return;
    const nextPage = postPage + 1; setLoadingMorePosts(true); setPostError(null);
    try {
      const data = (await postsAPI.searchPosts(query.trim(), nextPage, PAGE_SIZE))?.data || {};
      setPosts((current) => { const ids = new Set(current.map((p) => p.postId)); return [...current, ...(data.posts || []).filter((p) => !ids.has(p.postId))]; });
      setPostPage(data.pagination?.page ?? nextPage); setPostHasMore(Boolean(data.pagination?.hasMore));
    } catch (error) { setPostError(getApiErrorMessage(error, "Failed to load more posts.")); }
    finally { setLoadingMorePosts(false); }
  };

  const showPeople = activeTab === "all" || activeTab === "people";
  const showPosts = activeTab === "all" || activeTab === "posts";
  const hasResults = (showPeople && users.length) || (showPosts && posts.length);
  const hasErrors = Boolean((showPeople && userError) || (showPosts && postError));
  const hasSuggestions = suggestions.users.length > 0 || suggestions.posts.length > 0;

  return (
    <section className="min-h-[calc(100dvh-6rem)] w-full bg-neutral-950 px-3 py-3 text-neutral-100 sm:px-5 sm:py-5">
      <div className="mx-auto w-full max-w-4xl">
        <header className="mb-4 flex items-center gap-2 sm:mb-5">
          <Link to="/" className="rounded-full p-2 text-neutral-400 transition hover:bg-neutral-900 hover:text-white focus:outline-none focus:ring-2 focus:ring-white/20" aria-label="Back to feed"><ArrowLeft size={19} /></Link>
          <div><h1 className="text-lg font-bold tracking-tight sm:text-xl">Discover</h1><p className="text-xs text-neutral-500 sm:text-sm">Find people and posts on Notell.</p></div>
        </header>

        <form onSubmit={submitSearch} className="relative">
          <Search className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-neutral-500" size={19} />
          <input value={query} onChange={(e) => onSearching(e.target.value)} onFocus={() => query.trim() && setShowSuggestions(true)} onBlur={() => setTimeout(() => setShowSuggestions(false), 150)} placeholder="Search people or posts" maxLength={100} autoFocus enterKeyHint="search" autoComplete="off" aria-label="Search people or posts" className="w-full rounded-2xl border border-neutral-800 bg-neutral-900 py-3.5 pl-11 pr-24 text-sm text-white shadow-sm outline-none transition placeholder:text-neutral-600 focus:border-neutral-600 focus:ring-2 focus:ring-white/5" />
          {query && <button type="button" onMouseDown={(e) => e.preventDefault()} onClick={clearSearch} className="absolute right-20 top-1/2 -translate-y-1/2 rounded-full p-1.5 text-neutral-500 hover:bg-neutral-800 hover:text-white" aria-label="Clear search"><X size={16} /></button>}
          <button type="submit" className="absolute right-2 top-1/2 -translate-y-1/2 rounded-xl bg-white px-3.5 py-2 text-xs font-bold text-neutral-950 transition hover:bg-neutral-200 active:scale-95 sm:text-sm">Search</button>

          {showSuggestions && query.trim().length >= MIN_QUERY_LENGTH && <div className="absolute left-0 right-0 top-[calc(100%+8px)] z-40 overflow-hidden rounded-2xl border border-neutral-800 bg-neutral-950 shadow-2xl shadow-black/50">
            {suggestionLoading ? <div className="flex items-center gap-2 px-4 py-4 text-sm text-neutral-500"><Loader2 size={16} className="animate-spin" />Searching…</div> : hasSuggestions ? <div className="max-h-[min(60vh,420px)] overflow-y-auto py-2">
              {suggestions.users.length > 0 && <div><p className="px-4 py-2 text-[10px] font-bold uppercase tracking-wider text-neutral-600">People</p>{suggestions.users.map((user) => <Link key={user.id} to={`/users/${user.id}`} onMouseDown={(e) => e.preventDefault()} onClick={() => setShowSuggestions(false)} className="flex items-center gap-3 px-4 py-2.5 transition hover:bg-neutral-900"><Avatar user={user} size="h-9 w-9" /><div className="min-w-0"><p className="truncate text-sm font-semibold">@{user.username}</p><p className="truncate text-xs text-neutral-500">{user.bio || [user.city, user.country].filter(Boolean).join(", ") || "View profile"}</p></div></Link>)}</div>}
              {suggestions.posts.length > 0 && <div className="border-t border-neutral-800/80 pt-1"><p className="px-4 py-2 text-[10px] font-bold uppercase tracking-wider text-neutral-600">Posts</p>{suggestions.posts.map((post) => <button key={post.postId} type="button" onMouseDown={(e) => e.preventDefault()} onClick={() => { setShowSuggestions(false); navigate(`/posts/${post.postId}`); }} className="flex w-full items-start gap-3 px-4 py-2.5 text-left transition hover:bg-neutral-900"><FileText size={16} className="mt-0.5 shrink-0 text-neutral-500" /><span className="min-w-0 truncate text-sm text-neutral-300">{post.caption || "Post matching your search"}</span></button>)}</div>}
            </div> : <div className="px-4 py-4 text-sm text-neutral-500">No matching suggestions yet.</div>}
          </div>}
        </form>

        <div className="mt-4 flex gap-1 overflow-x-auto rounded-xl border border-neutral-800 bg-neutral-900/60 p-1" role="tablist" aria-label="Search result type">
          {TABS.map((tab) => <button key={tab.id} type="button" role="tab" aria-selected={activeTab === tab.id} onClick={() => changeTab(tab.id)} className={`min-w-20 flex-1 rounded-lg px-3 py-2 text-xs font-semibold transition sm:text-sm ${activeTab === tab.id ? "bg-white text-neutral-950 shadow" : "text-neutral-400 hover:bg-neutral-800 hover:text-white"}`}>{tab.label}</button>)}
        </div>

        {!searched && !query.trim() && <div className="mx-auto mt-10 max-w-xl rounded-3xl border border-neutral-800 bg-neutral-900/40 px-5 py-12 text-center sm:mt-14"><div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-neutral-800 text-neutral-500"><Search size={25} /></div><h2 className="mt-4 text-base font-semibold">Search Notell</h2><p className="mx-auto mt-1 max-w-sm text-sm leading-6 text-neutral-500">Find people by username, bio or location, or discover posts by their content.</p></div>}

        <main className="mt-5 space-y-7 sm:mt-6">
          {loading && <div className="space-y-6"><div className="space-y-2"><PeopleSkeleton /><PeopleSkeleton /></div><div className="space-y-3"><PostSkeleton /><PostSkeleton /></div></div>}
          {!loading && searched && showPeople && userError && <div className="rounded-2xl border border-red-900/60 bg-red-950/30 p-4 text-sm text-red-300">{userError}</div>}
          {!loading && searched && showPosts && postError && <div className="rounded-2xl border border-red-900/60 bg-red-950/30 p-4 text-sm text-red-300">{postError}</div>}

          {!loading && searched && showPeople && users.length > 0 && <section>
            <div className="mb-3 flex items-end justify-between"><div><h2 className="text-sm font-bold sm:text-base">People</h2><p className="mt-0.5 text-xs text-neutral-600">{users.length}{userHasMore ? "+" : ""} result{users.length === 1 ? "" : "s"}</p></div>{activeTab === "all" && <button type="button" onClick={() => changeTab("people")} className="text-xs font-semibold text-neutral-400 hover:text-white sm:text-sm">See all</button>}</div>
            <div className="space-y-2">{users.map((user) => <div key={user.id} className="flex items-center gap-3 rounded-2xl border border-neutral-800 bg-neutral-900/70 p-3 transition hover:border-neutral-700 hover:bg-neutral-900 sm:p-3.5">
              <Link to={`/users/${user.id}`} className="shrink-0 rounded-full focus:outline-none focus:ring-2 focus:ring-white/20" aria-label={`View @${user.username}'s public profile`}><Avatar user={user} /></Link>
              <Link to={`/users/${user.id}`} className="min-w-0 flex-1 rounded-xl focus:outline-none focus:ring-2 focus:ring-white/10" aria-label={`View @${user.username}'s public profile`}>
                <p className="truncate text-sm font-bold text-white">@{user.username}</p>
                {user.bio ? <p className="mt-0.5 truncate text-xs text-neutral-500">{user.bio}</p> : (user.city || user.country) ? <p className="mt-1 flex items-center gap-1 truncate text-xs text-neutral-500"><MapPin size={12} />{[user.city, user.country].filter(Boolean).join(", ")}</p> : <p className="mt-1 text-xs text-neutral-600">View public profile</p>}
              </Link>
              <div className="flex shrink-0 items-center gap-1.5">
                <Link to={`/users/${user.id}`} className="hidden rounded-xl border border-neutral-700 px-3 py-2 text-xs font-semibold text-neutral-300 transition hover:border-neutral-500 hover:text-white sm:inline-flex" aria-label={`View @${user.username}'s public profile`}>View</Link>
                <button type="button" onClick={() => handleFollow(user)} disabled={followLoading[user.id] || followState[user.id]} className={`inline-flex min-w-[82px] items-center justify-center gap-1.5 rounded-xl px-3 py-2 text-xs font-bold transition active:scale-95 disabled:cursor-default ${followState[user.id] ? "border border-emerald-500/20 bg-emerald-500/10 text-emerald-300" : "bg-white text-neutral-950 hover:bg-neutral-200 disabled:opacity-60"}`} aria-label={followState[user.id] ? `Following @${user.username}` : `Follow @${user.username}`}>
                  {followLoading[user.id] ? <Loader2 size={14} className="animate-spin" /> : followState[user.id] ? <Check size={14} /> : <UserPlus size={14} />}
                  {followLoading[user.id] ? "Following…" : followState[user.id] ? "Following" : "Follow"}
                </button>
              </div>
            </div>)}</div>
            {userHasMore && <button type="button" onClick={loadMoreUsers} disabled={loadingMoreUsers} className="mx-auto mt-4 flex items-center gap-2 rounded-xl border border-neutral-800 bg-neutral-900 px-4 py-2.5 text-xs font-semibold text-neutral-300 transition hover:bg-neutral-800 disabled:opacity-50">{loadingMoreUsers ? <Loader2 size={15} className="animate-spin" /> : <ChevronDown size={15} />}{loadingMoreUsers ? "Loading…" : "Load more people"}</button>}
          </section>}

          {!loading && searched && showPosts && posts.length > 0 && <section>
            <div className="mb-3 flex items-end justify-between"><div><h2 className="text-sm font-bold sm:text-base">Posts</h2><p className="mt-0.5 text-xs text-neutral-600">{posts.length}{postHasMore ? "+" : ""} result{posts.length === 1 ? "" : "s"}</p></div>{activeTab === "all" && <button type="button" onClick={() => changeTab("posts")} className="text-xs font-semibold text-neutral-400 hover:text-white sm:text-sm">See all</button>}</div>
            <div className="space-y-4">{posts.map((post) => <div key={post.postId} className="group"><PostCard post={post} onPostDeleted={(id) => setPosts((current) => current.filter((item) => item.postId !== id))} /><Link to={`/posts/${post.postId}`} className="-mt-1 flex items-center justify-between rounded-b-2xl border-x border-b border-transparent px-3 py-2 text-xs font-semibold text-neutral-500 transition hover:border-neutral-800 hover:bg-neutral-900/50 hover:text-white" aria-label={`Open post by ${post.user?.username || "user"}`}><span>View post</span><span className="text-neutral-600">{formatRelativeTime(post.createdAt)}</span></Link></div>)}</div>
            {postHasMore && <button type="button" onClick={loadMorePosts} disabled={loadingMorePosts} className="mx-auto mt-4 flex items-center gap-2 rounded-xl border border-neutral-800 bg-neutral-900 px-4 py-2.5 text-xs font-semibold text-neutral-300 transition hover:bg-neutral-800 disabled:opacity-50">{loadingMorePosts ? <Loader2 size={15} className="animate-spin" /> : <ChevronDown size={15} />}{loadingMorePosts ? "Loading…" : "Load more posts"}</button>}
          </section>}

          {!loading && searched && !hasResults && !hasErrors && <div className="rounded-3xl border border-neutral-800 bg-neutral-900/50 px-5 py-14 text-center"><div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-neutral-800 text-neutral-500">{activeTab === "people" ? <UserRound size={25} /> : activeTab === "posts" ? <FileText size={25} /> : <Search size={25} />}</div><p className="mt-4 font-semibold text-neutral-200">No {activeTab === "people" ? "people" : activeTab === "posts" ? "posts" : "results"} found</p><p className="mt-1 text-sm text-neutral-500">Try another keyword or check the spelling.</p></div>}
        </main>
      </div>
    </section>
  );
}
