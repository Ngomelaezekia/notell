import { useEffect, useState } from "react";
import {
  ArrowLeft,
  Check,
  ChevronDown,
  Loader2,
  UserPlus,
  UserRound,
  Users,
} from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { userAPI } from "../services/user/userApi";
import { getApiErrorMessage, getFileUrl } from "../utils/api";

const PAGE_SIZE = 20;

const RelationshipRow = ({ user, type, onChanged }) => {
  const [following, setFollowing] = useState(Boolean(user.following));
  const [loading, setLoading] = useState(false);
  const canFollow = user.allowFollowers !== false;

  const toggleFollow = async () => {
    if (loading || !user.id || !canFollow) return;
    setLoading(true);
    try {
      if (following) {
        await userAPI.unfollowUser(user.id);
        setFollowing(false);
        onChanged?.(user.id, false);
      } else {
        await userAPI.followUser(user.id);
        setFollowing(true);
        onChanged?.(user.id, true);
      }
    } catch {
      // Keep the current state when the relationship request fails.
    } finally {
      setLoading(false);
    }
  };

  const avatar = getFileUrl(user.profilePicture);
  const isFollowerList = type === "followers";

  return (
    <div className="group flex items-center gap-3 px-4 py-3.5 transition hover:bg-white/[0.035] sm:px-5">
      <Link
        to={`/users/${user.id}`}
        className="flex min-w-0 flex-1 items-center gap-3"
      >
        <div className="flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-full bg-gradient-to-br from-orange-400/20 to-white/5 text-sm font-bold text-white ring-1 ring-white/10">
          {avatar ? (
            <img
              src={avatar}
              alt={`${user.username} avatar`}
              className="h-full w-full object-cover"
              loading="lazy"
            />
          ) : (
            user.username?.charAt(0).toUpperCase()
          )}
        </div>
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-white">
            {user.username}
          </p>
          <div className="mt-0.5 flex min-w-0 items-center gap-1.5 text-[11px] text-white/40">
            {(user.city || user.country) && (
              <span className="truncate">
                {[user.city, user.country].filter(Boolean).join(", ")}
              </span>
            )}
            {user.bio && (user.city || user.country) && <span>·</span>}
            {user.bio && <span className="truncate">{user.bio}</span>}
          </div>
        </div>
      </Link>

      {!user.isSelf && (
        <button
          type="button"
          onClick={toggleFollow}
          disabled={loading || !canFollow}
          aria-label={`${following ? "Unfollow" : "Follow"} ${user.username}`}
          className={`inline-flex h-9 min-w-[84px] shrink-0 items-center justify-center gap-1.5 rounded-xl px-3 text-[11px] font-bold transition active:scale-[0.97] disabled:cursor-not-allowed disabled:opacity-45 ${following ? "border border-white/10 bg-white/[0.06] text-white/75 hover:bg-white/10" : "bg-orange-500 text-white shadow-lg shadow-orange-950/20 hover:bg-orange-400"}`}
        >
          {loading ? (
            <Loader2 size={14} className="animate-spin" />
          ) : following ? (
            <Check size={14} />
          ) : (
            <UserPlus size={14} />
          )}
          <span>{following ? "Following" : "Follow"}</span>
        </button>
      )}
    </div>
  );
};

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
  const subtitle = type === "followers"
    ? "People who follow this profile"
    : "People this profile follows";

  return (
    <div className="min-h-screen w-full bg-slate-950 pb-20 text-white lg:pb-8">
      <div className="mx-auto w-full max-w-2xl px-3 py-3 sm:px-5 sm:py-6">
        <header className="mb-3 flex h-12 items-center gap-3 sm:mb-5">
          <Link
            to={`/users/${id}`}
            aria-label="Back to profile"
            className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-white/10 bg-white/[0.035] text-white/70 transition hover:bg-white/10 hover:text-white"
          >
            <ArrowLeft size={18} />
          </Link>
          <div className="min-w-0">
            <h1 className="text-base font-bold text-white sm:text-lg">{title}</h1>
            <p className="truncate text-[11px] text-white/40">{subtitle}</p>
          </div>
          <div className="ml-auto flex h-9 w-9 items-center justify-center rounded-full bg-orange-500/10 text-orange-300">
            {type === "followers" ? <Users size={17} /> : <UserRound size={17} />}
          </div>
        </header>

        <section className="overflow-hidden rounded-3xl border border-white/10 bg-white/[0.035] shadow-2xl shadow-black/20">
          {loading ? (
            <div className="flex min-h-56 items-center justify-center text-white/45">
              <Loader2 size={22} className="animate-spin" />
            </div>
          ) : error && users.length === 0 ? (
            <div className="p-10 text-center text-sm text-red-300">{error}</div>
          ) : users.length === 0 ? (
            <div className="flex min-h-56 flex-col items-center justify-center px-8 text-center">
              <div className="flex h-14 w-14 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.04] text-white/30">
                {type === "followers" ? <Users size={24} /> : <UserRound size={24} />}
              </div>
              <p className="mt-4 text-sm font-semibold text-white/75">No {type} yet</p>
              <p className="mt-1 max-w-xs text-xs leading-5 text-white/35">
                {type === "followers"
                  ? "When people follow this profile, they will appear here."
                  : "Profiles this user follows will appear here."}
              </p>
            </div>
          ) : (
            <>
              {error && (
                <div className="border-b border-red-500/15 bg-red-500/5 px-4 py-3 text-xs text-red-300 sm:px-5">
                  {error}
                </div>
              )}
              <div className="divide-y divide-white/[0.07]">
                {users.map((user) => (
                  <RelationshipRow
                    key={user.id}
                    user={user}
                    type={type}
                  />
                ))}
              </div>
              {hasMore && (
                <div className="border-t border-white/[0.07] p-4 text-center">
                  <button
                    type="button"
                    onClick={() => void loadMore()}
                    disabled={loadingMore}
                    className="inline-flex min-h-10 items-center gap-2 rounded-xl border border-white/10 bg-white/[0.04] px-4 text-xs font-bold text-white/70 transition hover:bg-white/10 hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {loadingMore ? <Loader2 size={15} className="animate-spin" /> : <ChevronDown size={15} />}
                    {loadingMore ? "Loading…" : "Load more"}
                  </button>
                </div>
              )}
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
