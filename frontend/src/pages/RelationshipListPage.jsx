import { useEffect, useState } from "react";
import { ArrowLeft, ChevronDown, Loader2, UserRound, Users } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { userAPI } from "../services/user/userApi";
import { getApiErrorMessage, getFileUrl } from "../utils/api";

const PAGE_SIZE = 20;

const RelationshipListPage = ({ type }) => {
  const { id } = useParams();
  const [users, setUsers] = useState([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    let active = true;
    const load = async () => {
      if (!id) return;
      setLoading(true); setError(null); setPage(1);
      try {
        const response = type === "followers" ? await userAPI.getFollowers(id, 1, PAGE_SIZE) : await userAPI.getFollowing(id, 1, PAGE_SIZE);
        if (active) { setUsers(response?.data ?? []); setPage(response?.pagination?.page ?? 1); setHasMore(Boolean(response?.pagination?.hasMore)); }
      } catch (err) { if (active) setError(getApiErrorMessage(err, `Failed to load ${type}`)); }
      finally { if (active) setLoading(false); }
    };
    void load(); return () => { active = false; };
  }, [id, type]);

  const loadMore = async () => {
    if (!id || loadingMore || !hasMore) return;
    const nextPage = page + 1; setLoadingMore(true); setError(null);
    try {
      const response = type === "followers" ? await userAPI.getFollowers(id, nextPage, PAGE_SIZE) : await userAPI.getFollowing(id, nextPage, PAGE_SIZE);
      const nextUsers = response?.data ?? [];
      setUsers((current) => { const ids = new Set(current.map((user) => user.id)); return [...current, ...nextUsers.filter((user) => !ids.has(user.id))]; });
      setPage(response?.pagination?.page ?? nextPage); setHasMore(Boolean(response?.pagination?.hasMore));
    } catch (err) { setError(getApiErrorMessage(err, `Failed to load more ${type}`)); }
    finally { setLoadingMore(false); }
  };

  const title = type === "followers" ? "Followers" : "Following";
  const subtitle = type === "followers" ? "People who follow this profile" : "People this profile follows";

  return (
    <div className="min-h-screen w-full bg-neutral-950 pb-32 text-neutral-100 lg:pb-10">
      <div className="mx-auto w-full max-w-2xl px-3 py-4 sm:px-5 sm:py-6">
        <header className="mb-4 flex h-12 items-center gap-3 sm:mb-5">
          <Link to={`/users/${id}`} aria-label="Back to profile" className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-neutral-800 bg-neutral-900 text-neutral-400 transition hover:border-neutral-700 hover:text-white"><ArrowLeft size={18} /></Link>
          <div className="min-w-0"><h1 className="text-base font-bold sm:text-lg">{title}</h1><p className="truncate text-[11px] text-neutral-500">{subtitle}</p></div>
          <div className="ml-auto flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-neutral-800 bg-neutral-900 text-neutral-400">{type === "followers" ? <Users size={17} /> : <UserRound size={17} />}</div>
        </header>

        <section className="overflow-hidden rounded-3xl border border-neutral-800/90 bg-neutral-900/70 shadow-xl shadow-black/20">
          {loading ? <div className="flex min-h-64 items-center justify-center text-neutral-500"><Loader2 size={22} className="animate-spin" /></div> : error && users.length === 0 ? <div className="p-10 text-center text-sm text-red-300">{error}</div> : users.length === 0 ? (
            <div className="flex min-h-64 flex-col items-center justify-center px-8 text-center"><div className="flex h-14 w-14 items-center justify-center rounded-2xl border border-neutral-800 bg-neutral-950 text-neutral-600">{type === "followers" ? <Users size={24} /> : <UserRound size={24} />}</div><p className="mt-4 text-sm font-semibold text-neutral-300">No {type} yet</p><p className="mt-1 max-w-xs text-xs leading-5 text-neutral-500">{type === "followers" ? "When people follow this profile, they will appear here." : "Profiles this user follows will appear here."}</p></div>
          ) : (
            <>
              {error && <div className="border-b border-red-900/50 bg-red-950/20 px-4 py-3 text-xs text-red-300">{error}</div>}
              <div className="divide-y divide-neutral-800/70">
                {users.map((user) => {
                  const avatar = getFileUrl(user.profilePicture);
                  return <Link key={user.id} to={`/users/${user.id}`} className="group flex items-center gap-3 px-4 py-3.5 transition hover:bg-neutral-800/50 sm:px-5">
                    <div className="flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-full bg-neutral-800 text-sm font-bold text-neutral-400 ring-1 ring-neutral-700 group-hover:ring-neutral-600">{avatar ? <img src={avatar} alt={`${user.username} avatar`} className="h-full w-full object-cover" loading="lazy" /> : user.username?.charAt(0).toUpperCase()}</div>
                    <div className="min-w-0 flex-1"><p className="truncate text-sm font-semibold text-neutral-200 group-hover:text-white">{user.username}</p><div className="mt-0.5 flex min-w-0 items-center gap-1.5 text-[11px] text-neutral-500">{(user.city || user.country) && <span className="truncate">{[user.city, user.country].filter(Boolean).join(", ")}</span>}{user.bio && (user.city || user.country) && <span>·</span>}{user.bio && <span className="truncate">{user.bio}</span>}</div></div>
                    <span className="shrink-0 rounded-lg border border-neutral-800 bg-neutral-950/60 px-2.5 py-1.5 text-[10px] font-semibold text-neutral-500 group-hover:text-neutral-300">View</span>
                  </Link>;
                })}
              </div>
              {hasMore && <div className="border-t border-neutral-800/70 p-4 text-center"><button type="button" onClick={() => void loadMore()} disabled={loadingMore} className="inline-flex min-h-10 items-center gap-2 rounded-xl border border-neutral-800 bg-neutral-950 px-4 text-xs font-bold text-neutral-300 transition hover:bg-neutral-800 hover:text-white disabled:opacity-50">{loadingMore ? <Loader2 size={15} className="animate-spin" /> : <ChevronDown size={15} />}{loadingMore ? "Loading…" : "Load more"}</button></div>}
            </>
          )}
        </section>
      </div>
    </div>
  );
};

export const FollowersPage = () => <RelationshipListPage type="followers" />;
export const FollowingPage = () => <RelationshipListPage type="following" />;
export default RelationshipListPage;
