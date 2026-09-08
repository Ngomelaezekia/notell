import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ArrowLeft,
  Camera,
  Check,
  ExternalLink,
  Link2,
  Loader2,
  MapPin,
  MessageCircle,
  MoreVertical,
  Play,
  Settings,
  UserPlus,
  UserRound,
  Users as UsersIcon,
  X,
} from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { userAPI } from "../services/user/userApi";
import { getFileUrl } from "../utils/api";

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

const extractBioLinks = (bio = "") => {
  const matches = bio.match(URL_PATTERN) ?? [];
  return [...new Set(matches.map(cleanUrl))].slice(0, 5);
};

const Stat = ({ label, value, icon: Icon, onClick }) => (
  <button
    type="button"
    onClick={onClick}
    className="group flex min-w-0 flex-1 items-center justify-center gap-2 rounded-xl px-2 py-2 text-center transition hover:bg-white/10"
  >
    <Icon size={17} strokeWidth={1.8} className="text-white/75" />
    <span>
      <span className="block text-base font-bold leading-5 text-white">{value ?? 0}</span>
      <span className="block text-[10px] font-medium uppercase tracking-wide text-white/55 group-hover:text-white/75">{label}</span>
    </span>
  </button>
);

const MediaTile = ({ post }) => {
  const mediaUrl = getFileUrl(post?.contentUrl);
  const isVideo = post?.contentType === "video";

  return (
    <Link
      to={`/posts/${post?.postId}`}
      className="group relative aspect-square overflow-hidden rounded-2xl bg-slate-900 ring-1 ring-white/10 transition hover:-translate-y-0.5 hover:ring-white/20"
    >
      {isVideo ? (
        <video
          src={mediaUrl}
          muted
          playsInline
          preload="metadata"
          className="h-full w-full object-cover transition duration-300 group-hover:scale-105"
        />
      ) : (
        <img
          src={mediaUrl}
          alt={post?.caption || "Post"}
          loading="lazy"
          className="h-full w-full object-cover transition duration-300 group-hover:scale-105"
        />
      )}
      <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent opacity-0 transition group-hover:opacity-100" />
      {isVideo && (
        <span className="absolute right-2.5 top-2.5 inline-flex h-8 w-8 items-center justify-center rounded-full bg-black/60 text-white backdrop-blur-sm">
          <Play size={14} fill="currentColor" />
        </span>
      )}
      {post?.caption && (
        <span className="absolute bottom-2.5 left-2.5 right-2.5 line-clamp-2 text-[11px] font-medium text-white opacity-0 transition group-hover:opacity-100">
          {post.caption}
        </span>
      )}
    </Link>
  );
};

const RelationshipPanel = ({ type, users, loading, error, onClose }) => {
  const title = type === "followers" ? "Followers" : "Following";

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/60 p-0 backdrop-blur-sm sm:items-center sm:p-5"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose();
      }}
    >
      <section
        className="flex max-h-[78vh] w-full flex-col overflow-hidden rounded-t-3xl border border-slate-200 bg-white shadow-2xl sm:max-w-md sm:rounded-3xl"
        role="dialog"
        aria-modal="true"
        aria-label={title}
      >
        <header className="flex items-center justify-between border-b border-slate-100 px-5 py-4">
          <div>
            <h2 className="text-base font-bold text-slate-950">{title}</h2>
            <p className="mt-0.5 text-[11px] text-slate-500">People connected to this profile</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="inline-flex h-9 w-9 items-center justify-center rounded-full text-slate-500 transition hover:bg-slate-100 hover:text-slate-900"
          >
            <X size={18} />
          </button>
        </header>

        {loading ? (
          <div className="flex min-h-52 items-center justify-center text-slate-500">
            <Loader2 size={22} className="animate-spin" />
          </div>
        ) : error && users.length === 0 ? (
          <div className="p-8 text-center text-sm text-red-600">{error}</div>
        ) : users.length === 0 ? (
          <div className="flex min-h-52 flex-col items-center justify-center p-8 text-center">
            <UserRound size={32} className="text-slate-300" />
            <p className="mt-3 text-sm font-medium text-slate-600">No {type} yet</p>
          </div>
        ) : (
          <div className="overflow-y-auto">
            {error && <div className="border-b border-red-100 bg-red-50 px-5 py-3 text-xs text-red-600">{error}</div>}
            <div className="divide-y divide-slate-100">
              {users.map((person) => {
                const avatar = getFileUrl(person.profilePicture);
                return (
                  <Link
                    key={person.id}
                    to={`/users/${person.id}`}
                    onClick={onClose}
                    className="flex items-center gap-3 px-5 py-3.5 transition hover:bg-slate-50"
                  >
                    <div className="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-full bg-slate-100 font-semibold text-slate-600">
                      {avatar ? <img src={avatar} alt={`${person.username} avatar`} className="h-full w-full object-cover" /> : person.username?.charAt(0).toUpperCase()}
                    </div>
                    <div className="min-w-0">
                      <p className="truncate text-sm font-semibold text-slate-900">{person.username}</p>
                      {(person.city || person.country) && <p className="truncate text-[11px] text-slate-500">{[person.city, person.country].filter(Boolean).join(", ")}</p>}
                      {person.bio && <p className="truncate text-[11px] text-slate-500">{person.bio}</p>}
                    </div>
                  </Link>
                );
              })}
            </div>
          </div>
        )}
      </section>
    </div>
  );
};

export const Users = () => {
  const { id } = useParams();
  const [user, setUser] = useState(null);
  const [relationship, setRelationship] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);
  const [relationshipPanel, setRelationshipPanel] = useState(null);
  const [relationshipUsers, setRelationshipUsers] = useState([]);
  const [relationshipLoading, setRelationshipLoading] = useState(false);
  const [relationshipError, setRelationshipError] = useState(null);

  const loadProfile = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    try {
      const [profileResponse, relationshipResponse] = await Promise.all([
        userAPI.getProfile(id),
        userAPI.getRelationship(id),
      ]);
      setUser(profileResponse?.data?.user ?? null);
      setRelationship(relationshipResponse?.data ?? null);
    } catch (err) {
      setError(err.response?.data?.message || "Failed to load profile");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    loadProfile();
  }, [loadProfile]);

  const openRelationshipPanel = async (type) => {
    if (!id) return;
    setRelationshipPanel(type);
    setRelationshipUsers([]);
    setRelationshipError(null);
    setRelationshipLoading(true);
    try {
      const response = type === "followers"
        ? await userAPI.getFollowers(id, 1, 100)
        : await userAPI.getFollowing(id, 1, 100);
      setRelationshipUsers(Array.isArray(response?.data) ? response.data : []);
    } catch (err) {
      setRelationshipError(err.response?.data?.message || `Failed to load ${type}`);
    } finally {
      setRelationshipLoading(false);
    }
  };

  const handleFollow = async () => {
    if (!id || actionLoading || (!relationship?.following && !relationship?.allowFollowers)) return;
    setActionLoading(true);
    setError(null);
    try {
      if (relationship.following) {
        await userAPI.unfollowUser(id);
        setRelationship((current) => ({
          ...current,
          following: false,
          followerCount: Math.max(0, (current?.followerCount ?? 1) - 1),
        }));
      } else {
        await userAPI.followUser(id);
        setRelationship((current) => ({
          ...current,
          following: true,
          followerCount: (current?.followerCount ?? 0) + 1,
        }));
      }
    } catch (err) {
      setError(err.response?.data?.message || "Failed to update follow status");
    } finally {
      setActionLoading(false);
    }
  };

  const avatar = getFileUrl(user?.profilePicture);
  const cover = getFileUrl(user?.coverPicture);
  const posts = Array.isArray(user?.posts) ? user.posts : [];
  const bioLinks = useMemo(() => extractBioLinks(user?.bio), [user?.bio]);
  const joined = user?.createdAt
    ? new Date(user.createdAt).toLocaleDateString(undefined, { month: "short", year: "numeric" })
    : "";
  const isSelf = Boolean(relationship?.isSelf);
  const canFollow = Boolean(relationship?.following || relationship?.allowFollowers);

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center bg-slate-950 text-white">
        <Loader2 className="animate-spin" size={24} />
      </div>
    );
  }

  if (error && !user) {
    return (
      <div className="mx-auto max-w-xl p-6">
        <Link to="/" className="mb-5 inline-flex items-center gap-2 text-sm font-medium text-slate-600">
          <ArrowLeft size={17} /> Back to home
        </Link>
        <div className="rounded-3xl border border-red-100 bg-white p-8 text-center text-red-600 shadow-sm">{error}</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen w-full bg-slate-950 pb-24 text-white lg:pb-8">
      <div className="mx-auto w-full max-w-6xl">
        <div className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-white/10 bg-slate-950/85 px-4 backdrop-blur-xl sm:px-6">
          <Link
            to="/"
            aria-label="Back to home"
            className="inline-flex h-9 w-9 items-center justify-center rounded-full text-white/80 transition hover:bg-white/10 hover:text-white"
          >
            <ArrowLeft size={20} />
          </Link>
          <div className="min-w-0 text-center">
            <span className="truncate text-sm font-semibold text-white">{user?.username}</span>
            {user?.status && user.status !== "free" && (
              <span className="ml-2 rounded-full border border-orange-300/40 bg-orange-400/10 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wide text-orange-300">
                {user.status}
              </span>
            )}
          </div>
          {isSelf ? (
            <Link
              to="/settings"
              aria-label="Profile settings"
              className="inline-flex h-9 w-9 items-center justify-center rounded-full text-white/80 transition hover:bg-white/10 hover:text-white"
            >
              <Settings size={19} />
            </Link>
          ) : (
            <button type="button" aria-label="Profile menu" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-white/45">
              <MoreVertical size={20} />
            </button>
          )}
        </div>

        <section className="relative overflow-hidden border-b border-white/10 bg-black lg:rounded-b-[32px] lg:border-x lg:border-white/10">
          <div className="relative h-48 overflow-hidden bg-gradient-to-br from-slate-800 via-slate-700 to-slate-950 sm:h-64 lg:h-80">
            {cover && <img src={cover} alt={`${user?.username} cover`} className="h-full w-full object-cover" />}
            <div className="absolute inset-0 bg-gradient-to-b from-black/5 via-black/10 to-black/90" />
            <div className="absolute inset-x-0 bottom-0 h-32 bg-gradient-to-t from-black to-transparent" />
          </div>

          <div className="relative mx-auto max-w-4xl px-4 pb-5 sm:px-6 lg:px-8">
            <div className="-mt-14 flex flex-col items-center sm:-mt-16 lg:-mt-20">
              <div className="relative">
                <div className="flex h-28 w-28 items-center justify-center overflow-hidden rounded-full border-4 border-black bg-slate-800 text-3xl font-bold text-white shadow-2xl sm:h-32 sm:w-32">
                  {avatar ? <img src={avatar} alt={`${user?.username} avatar`} className="h-full w-full object-cover" /> : user?.username?.charAt(0).toUpperCase()}
                </div>
                {isSelf && (
                  <Link
                    to="/profile"
                    aria-label="Edit profile picture"
                    className="absolute bottom-1 right-0 inline-flex h-9 w-9 items-center justify-center rounded-full border-2 border-black bg-slate-800 text-white shadow-lg transition hover:bg-orange-500"
                  >
                    <Camera size={15} />
                  </Link>
                )}
              </div>

              <div className="mt-3 text-center">
                <div className="flex flex-wrap items-center justify-center gap-2">
                  <h1 className="text-2xl font-bold tracking-tight text-white sm:text-3xl">{user?.username}</h1>
                  {user?.status && user.status !== "free" && (
                    <span className="rounded-full border border-orange-400/60 bg-orange-400/10 px-2.5 py-1 text-[9px] font-bold uppercase tracking-widest text-orange-300">
                      {user.status}
                    </span>
                  )}
                </div>
                {joined && <p className="mt-1 text-[11px] text-white/50">Member since {joined}</p>}
                {(user?.city || user?.country) && (
                  <div className="mt-2 inline-flex items-center gap-1.5 text-[11px] text-white/55">
                    <MapPin size={13} />
                    {[user.city, user.country].filter(Boolean).join(", ")}
                  </div>
                )}
                {user?.bio && <p className="mx-auto mt-3 max-w-xl whitespace-pre-line text-sm leading-5 text-white/70">{user.bio}</p>}
                {bioLinks.length > 0 && (
                  <div className="mt-2 flex flex-wrap justify-center gap-1.5">
                    {bioLinks.map((url) => (
                      <a
                        key={url}
                        href={url}
                        target="_blank"
                        rel="noreferrer noopener"
                        className="inline-flex items-center gap-1 rounded-full border border-white/10 bg-white/5 px-2.5 py-1 text-[10px] font-semibold text-white/75 transition hover:bg-white/10 hover:text-white"
                      >
                        <Link2 size={11} />
                        {getLinkLabel(url)}
                        <ExternalLink size={9} />
                      </a>
                    ))}
                  </div>
                )}
              </div>

              {!isSelf && (
                <div className="mt-5 flex w-full max-w-lg items-stretch gap-2 sm:gap-3">
                  <button
                    type="button"
                    onClick={handleFollow}
                    disabled={actionLoading || !canFollow}
                    aria-pressed={Boolean(relationship?.following)}
                    aria-label={relationship?.following ? `Unfollow ${user?.username}` : `Follow ${user?.username}`}
                    className={`group relative inline-flex min-h-12 flex-[1.5] items-center justify-center gap-2.5 rounded-2xl px-5 text-sm font-bold shadow-lg transition-all duration-200 active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50 sm:min-h-14 sm:text-[15px] ${relationship?.following ? "border border-white/20 bg-white/10 text-white shadow-black/20 hover:bg-white/15" : "bg-orange-500 text-white shadow-orange-950/30 hover:bg-orange-400 hover:shadow-orange-950/40"}`}
                  >
                    {actionLoading ? (
                      <Loader2 size={18} className="animate-spin" />
                    ) : relationship?.following ? (
                      <Check size={18} strokeWidth={2.5} />
                    ) : (
                      <UserPlus size={18} strokeWidth={2.5} />
                    )}
                    <span>{actionLoading ? "Updating…" : relationship?.following ? "Following" : canFollow ? "Follow" : "Followers disabled"}</span>
                  </button>
                  <button
                    type="button"
                    aria-label={`Chat with ${user?.username}`}
                    className="inline-flex min-h-12 w-12 shrink-0 items-center justify-center rounded-2xl border border-orange-500/60 bg-orange-500/5 text-orange-300 transition-all hover:bg-orange-500/10 hover:text-orange-200 active:scale-[0.98] sm:min-h-14 sm:w-14"
                  >
                    <MessageCircle size={19} />
                  </button>
                </div>
              )}
            </div>

            <div className="mt-5 flex w-full max-w-xl mx-auto items-center rounded-2xl border border-white/10 bg-white/[0.04] p-1">
              <Stat label="Posts" value={posts.length} icon={Camera} />
              <div className="h-8 w-px bg-white/10" />
              <Stat label="Followers" value={relationship?.followerCount} icon={UsersIcon} onClick={() => void openRelationshipPanel("followers")} />
              <div className="h-8 w-px bg-white/10" />
              <Stat label="Following" value={relationship?.followingCount} icon={UserRound} onClick={() => void openRelationshipPanel("following")} />
            </div>

            {relationship?.follower && !relationship?.following && (
              <div className="mt-3 text-center text-[10px] text-white/45">
                <UsersIcon size={12} className="mr-1 inline" /> This user follows you
              </div>
            )}
            {error && <p className="mt-3 text-center text-xs text-red-400">{error}</p>}
          </div>
        </section>

        <section className="mx-auto mt-3 max-w-6xl px-3 sm:px-6">
          <div className="rounded-3xl border border-white/10 bg-white/[0.035] p-2 shadow-2xl shadow-black/20 sm:p-4">
            <div className="flex items-center justify-between border-b border-white/10 px-2 pb-3 sm:px-3">
              <div>
                <h2 className="text-sm font-bold text-white">All Posts</h2>
                <p className="mt-0.5 text-[10px] text-white/40">Photos and videos from {user?.username}</p>
              </div>
              <button type="button" className="inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-[10px] font-semibold text-white/60 transition hover:bg-white/10 hover:text-white">
                Newest
                <MoreVertical size={13} className="rotate-90" />
              </button>
            </div>

            {posts.length === 0 ? (
              <div className="flex min-h-64 flex-col items-center justify-center px-6 text-center">
                <div className="flex h-14 w-14 items-center justify-center rounded-full bg-white/5 text-white/35">
                  <Camera size={23} />
                </div>
                <h3 className="mt-4 text-sm font-semibold text-white/80">No posts yet</h3>
                <p className="mt-1 max-w-xs text-xs leading-5 text-white/40">Posts, images and videos will appear here when this profile shares them.</p>
              </div>
            ) : (
              <div className="grid grid-cols-2 gap-2 pt-3 sm:grid-cols-3 sm:gap-3 lg:grid-cols-4">
                {posts.map((post) => <MediaTile key={post.postId} post={post} />)}
              </div>
            )}
          </div>
        </section>
      </div>

      {relationshipPanel && (
        <RelationshipPanel
          type={relationshipPanel}
          users={relationshipUsers}
          loading={relationshipLoading}
          error={relationshipError}
          onClose={() => setRelationshipPanel(null)}
        />
      )}
    </div>
  );
};

export default Users;
