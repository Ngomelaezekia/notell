import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { Search, UserRound, FileText, Loader2, ChevronDown } from "lucide-react";
import { userAPI } from "../services/user/userApi";
import { postsAPI } from "../services/post/postsApi";
import { getFileUrl } from "../utils/api";
import { PostCard } from "../components/PostCard";

const PAGE_SIZE = 20;

const TABS = [
  { id: "all", label: "All" },
  { id: "people", label: "People" },
  { id: "posts", label: "Posts" },
];

export default function SearchPage() {
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

  useEffect(() => {
    const value = searchParams.get("q")?.trim() || "";
    const requestedType = searchParams.get("type") || "all";
    const type = TABS.some((tab) => tab.id === requestedType) ? requestedType : "all";

    setQuery(value);
    setActiveTab(type);

    if (value.length < 2) {
      setUsers([]);
      setPosts([]);
      setUserPage(1);
      setPostPage(1);
      setUserHasMore(false);
      setPostHasMore(false);
      setSearched(false);
      setUserError(null);
      setPostError(null);
      return;
    }

    let cancelled = false;
    setLoading(true);
    setUsers([]);
    setPosts([]);
    setUserPage(1);
    setPostPage(1);
    setUserHasMore(false);
    setPostHasMore(false);
    setUserError(null);
    setPostError(null);

    Promise.allSettled([
      userAPI.searchUsers(value, 1, PAGE_SIZE),
      postsAPI.searchPosts(value, 1, PAGE_SIZE),
    ])
      .then(([userResult, postResult]) => {
        if (cancelled) return;

        if (userResult.status === "fulfilled") {
          const data = userResult.value?.data || {};
          setUsers(data.users || []);
          setUserHasMore(Boolean(data.pagination?.hasMore));
        } else {
          setUsers([]);
          setUserHasMore(false);
          setUserError(userResult.reason?.response?.data?.message || "People search failed.");
        }

        if (postResult.status === "fulfilled") {
          const data = postResult.value?.data || {};
          setPosts(data.posts || []);
          setPostHasMore(Boolean(data.pagination?.hasMore));
        } else {
          setPosts([]);
          setPostHasMore(false);
          setPostError(postResult.reason?.response?.data?.message || "Post search failed.");
        }

        setSearched(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [searchParams]);

  const submitSearch = (event) => {
    event.preventDefault();
    const value = query.trim();

    if (value.length < 2) {
      setSearchParams({});
      return;
    }

    const nextParams = { q: value };
    if (activeTab !== "all") nextParams.type = activeTab;
    setSearchParams(nextParams);
  };

  const changeTab = (tab) => {
    setActiveTab(tab);
    const value = query.trim();
    const nextParams = value.length >= 2 ? { q: value } : {};
    if (tab !== "all" && value.length >= 2) nextParams.type = tab;
    setSearchParams(nextParams);
  };

  const loadMoreUsers = async () => {
    if (loadingMoreUsers || !userHasMore || query.trim().length < 2) return;

    const nextPage = userPage + 1;
    setLoadingMoreUsers(true);
    setUserError(null);

    try {
      const result = await userAPI.searchUsers(query.trim(), nextPage, PAGE_SIZE);
      const data = result?.data || {};
      const nextUsers = data.users || [];

      setUsers((current) => {
        const existingIds = new Set(current.map((user) => user.id));
        return [...current, ...nextUsers.filter((user) => !existingIds.has(user.id))];
      });
      setUserPage(nextPage);
      setUserHasMore(Boolean(data.pagination?.hasMore));
    } catch (error) {
      setUserError(error?.response?.data?.message || "Failed to load more people.");
    } finally {
      setLoadingMoreUsers(false);
    }
  };

  const loadMorePosts = async () => {
    if (loadingMorePosts || !postHasMore || query.trim().length < 2) return;

    const nextPage = postPage + 1;
    setLoadingMorePosts(true);
    setPostError(null);

    try {
      const result = await postsAPI.searchPosts(query.trim(), nextPage, PAGE_SIZE);
      const data = result?.data || {};
      const nextPosts = data.posts || [];

      setPosts((current) => {
        const existingIds = new Set(current.map((post) => post.postId));
        return [...current, ...nextPosts.filter((post) => !existingIds.has(post.postId))];
      });
      setPostPage(nextPage);
      setPostHasMore(Boolean(data.pagination?.hasMore));
    } catch (error) {
      setPostError(error?.response?.data?.message || "Failed to load more posts.");
    } finally {
      setLoadingMorePosts(false);
    }
  };

  const showPeople = activeTab === "all" || activeTab === "people";
  const showPosts = activeTab === "all" || activeTab === "posts";
  const hasResults = (showPeople && users.length > 0) || (showPosts && posts.length > 0);
  const hasErrors = Boolean((showPeople && userError) || (showPosts && postError));

  return (
    <section className="mx-auto w-full max-w-4xl">
      <div className="rounded-3xl border border-white/70 bg-white/80 p-5 shadow-xl backdrop-blur-xl sm:p-7">
        <div className="mb-6">
          <h1 className="text-2xl font-black text-slate-900">Discover</h1>
          <p className="mt-1 text-sm text-slate-500">Find people and posts across Notell.</p>
        </div>

        <form onSubmit={submitSearch} className="relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400" size={20} />
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search people or posts..."
            maxLength={100}
            autoFocus
            className="w-full rounded-2xl border border-slate-200 bg-white py-3.5 pl-12 pr-28 text-slate-900 outline-none transition focus:border-indigo-400 focus:ring-4 focus:ring-indigo-100"
          />
          <button
            type="submit"
            className="absolute right-2 top-1/2 -translate-y-1/2 rounded-xl bg-slate-900 px-4 py-2 text-sm font-bold text-white transition hover:bg-slate-700"
          >
            Search
          </button>
        </form>

        <div className="mt-5 flex gap-2 overflow-x-auto border-b border-slate-100 pb-3">
          {TABS.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => changeTab(tab.id)}
              className={`shrink-0 rounded-xl px-4 py-2 text-sm font-bold transition ${
                activeTab === tab.id
                  ? "bg-slate-900 text-white"
                  : "bg-slate-100 text-slate-600 hover:bg-slate-200"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <div className="mt-6">
          {loading && (
            <div className="flex items-center justify-center gap-2 py-12 text-sm text-slate-500">
              <Loader2 size={18} className="animate-spin" /> Searching...
            </div>
          )}

          {!loading && searched && showPeople && userError && (
            <p className="mb-4 rounded-2xl bg-red-50 p-4 text-sm text-red-600">{userError}</p>
          )}

          {!loading && searched && showPosts && postError && (
            <p className="mb-4 rounded-2xl bg-red-50 p-4 text-sm text-red-600">{postError}</p>
          )}

          {!loading && !hasResults && !hasErrors && searched && (
            <div className="py-12 text-center">
              {activeTab === "people" ? (
                <UserRound className="mx-auto text-slate-300" size={42} />
              ) : activeTab === "posts" ? (
                <FileText className="mx-auto text-slate-300" size={42} />
              ) : (
                <Search className="mx-auto text-slate-300" size={42} />
              )}
              <p className="mt-3 font-semibold text-slate-700">No results found</p>
              <p className="mt-1 text-sm text-slate-500">Try a different search term.</p>
            </div>
          )}

          {!loading && searched && showPeople && users.length > 0 && (
            <section className={showPosts && posts.length > 0 ? "mb-8" : ""}>
              <div className="mb-3 flex items-center justify-between">
                <h2 className="text-lg font-black text-slate-900">People</h2>
                {activeTab === "all" && (
                  <button
                    type="button"
                    onClick={() => changeTab("people")}
                    className="text-sm font-bold text-indigo-600 hover:text-indigo-800"
                  >
                    See all
                  </button>
                )}
              </div>

              <div className="space-y-3">
                {users.map((user) => (
                  <Link
                    key={user.id}
                    to={`/users/${user.id}`}
                    className="flex items-center gap-4 rounded-2xl border border-slate-100 bg-white p-4 transition hover:border-indigo-200 hover:shadow-md"
                  >
                    <div className="h-12 w-12 shrink-0 overflow-hidden rounded-full bg-slate-100">
                      {user.profilePicture ? (
                        <img src={getFileUrl(user.profilePicture)} alt="" className="h-full w-full object-cover" />
                      ) : (
                        <div className="flex h-full w-full items-center justify-center text-slate-400">
                          <UserRound size={22} />
                        </div>
                      )}
                    </div>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-bold text-slate-900">@{user.username}</p>
                      {(user.city || user.country) && (
                        <p className="truncate text-xs text-slate-500">
                          {[user.city, user.country].filter(Boolean).join(", ")}
                        </p>
                      )}
                      {user.bio && <p className="mt-1 truncate text-sm text-slate-600">{user.bio}</p>}
                    </div>
                  </Link>
                ))}
              </div>

              {userHasMore && (
                <button
                  type="button"
                  onClick={loadMoreUsers}
                  disabled={loadingMoreUsers}
                  className="mx-auto mt-4 flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-bold text-slate-700 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {loadingMoreUsers ? <Loader2 size={16} className="animate-spin" /> : <ChevronDown size={16} />}
                  {loadingMoreUsers ? "Loading..." : "Load more people"}
                </button>
              )}
            </section>
          )}

          {!loading && searched && showPosts && posts.length > 0 && (
            <section>
              <div className="mb-3 flex items-center justify-between">
                <h2 className="text-lg font-black text-slate-900">Posts</h2>
                {activeTab === "all" && (
                  <button
                    type="button"
                    onClick={() => changeTab("posts")}
                    className="text-sm font-bold text-indigo-600 hover:text-indigo-800"
                  >
                    See all
                  </button>
                )}
              </div>

              <div>{posts.map((post) => <PostCard key={post.postId} post={post} />)}</div>

              {postHasMore && (
                <button
                  type="button"
                  onClick={loadMorePosts}
                  disabled={loadingMorePosts}
                  className="mx-auto mt-4 flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-bold text-slate-700 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  {loadingMorePosts ? <Loader2 size={16} className="animate-spin" /> : <ChevronDown size={16} />}
                  {loadingMorePosts ? "Loading..." : "Load more posts"}
                </button>
              )}
            </section>
          )}
        </div>
      </div>
    </section>
  );
}
