import { useEffect, useState } from "react";
import { ArrowLeft, ChevronDown, Loader2, UserRound } from "lucide-react";
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
      setLoading(true);
      setError(null);
      setPage(1);
      try {
        const response = type === "followers"
          ? await userAPI.getFollowers(id, 1, PAGE_SIZE)
          : await userAPI.getFollowing(id, 1, PAGE_SIZE);
        if (active) {
          setUsers(response?.data ?? []);
          setPage(response?.pagination?.page ?? 1);
          setHasMore(Boolean(response?.pagination?.hasMore));
        }
      } catch (err) {
        if (active) setError(getApiErrorMessage(err, `Failed to load ${type}`));
      } finally {
        if (active) setLoading(false);
      }
    };
    void load();
    return () => { active = false; };
  }, [id, type]);

  const loadMore = async () => {
    if (!id || loadingMore || !hasMore) return;
    const nextPage = page + 1;
    setLoadingMore(true);
    setError(null);
    try {
      const response = type === "followers"
        ? await userAPI.getFollowers(id, nextPage, PAGE_SIZE)
        : await userAPI.getFollowing(id, nextPage, PAGE_SIZE);
      const nextUsers = response?.data ?? [];
      setUsers((current) => {
        const existingIds = new Set(current.map((user) => user.id));
        return [...current, ...nextUsers.filter((user) => !existingIds.has(user.id))];
      });
      setPage(response?.pagination?.page ?? nextPage);
      setHasMore(Boolean(response?.pagination?.hasMore));
    } catch (err) {
      setError(getApiErrorMessage(err, `Failed to load more ${type}`));
    } finally {
      setLoadingMore(false);
    }
  };

  const title = type === "followers" ? "Followers" : "Following";

  return (
    <div className="mx-auto w-full max-w-2xl p-4 sm:p-6">
      <Link to={`/users/${id}`} className="mb-5 inline-flex items-center gap-2 text-sm font-medium text-slate-600 hover:text-slate-900">
        <ArrowLeft size={16} /> Back to profile
      </Link>

      <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div className="border-b border-slate-100 px-5 py-4 sm:px-6">
          <h1 className="text-lg font-bold text-slate-900">{title}</h1>
          <p className="mt-1 text-xs text-slate-500">People connected to this profile</p>
        </div>

        {loading ? (
          <div className="flex min-h-48 items-center justify-center text-slate-500"><Loader2 className="animate-spin" /></div>
        ) : error && users.length === 0 ? (
          <div className="p-8 text-center text-sm text-red-600">{error}</div>
        ) : users.length === 0 ? (
          <div className="p-10 text-center">
            <UserRound className="mx-auto text-slate-300" size={34} />
            <p className="mt-3 text-sm font-medium text-slate-600">No {type} yet</p>
          </div>
        ) : (
          <>
            {error && <div className="border-b border-red-100 bg-red-50 px-5 py-3 text-sm text-red-600">{error}</div>}
            <div className="divide-y divide-slate-100">
              {users.map((user) => {
                const avatar = getFileUrl(user.profilePicture);
                return (
                  <Link key={user.id} to={`/users/${user.id}`} className="flex items-center gap-3 px-5 py-4 transition hover:bg-slate-50 sm:px-6">
                    <div className="flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-full bg-slate-100 font-semibold text-slate-600">
                      {avatar ? <img src={avatar} alt={`${user.username} avatar`} className="h-full w-full object-cover" /> : user.username?.charAt(0).toUpperCase()}
                    </div>
                    <div className="min-w-0">
                      <p className="truncate text-sm font-semibold text-slate-900">{user.username}</p>
                      {(user.city || user.country) && <p className="truncate text-xs text-slate-500">{[user.city, user.country].filter(Boolean).join(", ")}</p>}
                      {user.bio && <p className="truncate text-xs text-slate-500">{user.bio}</p>}
                    </div>
                  </Link>
                );
              })}
            </div>
            {hasMore && (
              <div className="border-t border-slate-100 p-4 text-center">
                <button type="button" onClick={() => void loadMore()} disabled={loadingMore} className="mx-auto inline-flex items-center gap-2 rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-semibold text-slate-700 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60">
                  {loadingMore ? <Loader2 size={16} className="animate-spin" /> : <ChevronDown size={16} />}
                  {loadingMore ? "Loading..." : "Load more"}
                </button>
              </div>
            )}
          </>
        )}
      </section>
    </div>
  );
};

export const FollowersPage = () => <RelationshipListPage type="followers" />;
export const FollowingPage = () => <RelationshipListPage type="following" />;
export default RelationshipListPage;
