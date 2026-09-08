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
  UserPlus,
  Users as UsersIcon,
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

const Stat = ({ label, value, to }) => (
  <Link
    to={to}
    className="group min-w-[100px] rounded-2xl px-3 py-2 text-center transition hover:bg-slate-50"
  >
    <div className="text-lg font-bold text-slate-900 group-hover:text-indigo-600">{value ?? 0}</div>
    <div className="text-xs font-medium text-slate-500">{label}</div>
  </Link>
);

const MediaTile = ({ post }) => {
  const mediaUrl = getFileUrl(post?.contentUrl);
  const isVideo = post?.contentType === "video";

  return (
    <Link
      to={`/posts/${post?.postId}`}
      className="group relative aspect-square overflow-hidden rounded-2xl bg-slate-100 shadow-sm ring-1 ring-slate-200 transition hover:-translate-y-0.5 hover:shadow-md"
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
      <div className="absolute inset-0 bg-gradient-to-t from-black/45 via-transparent to-transparent opacity-0 transition group-hover:opacity-100" />
      {isVideo && (
        <span className="absolute right-3 top-3 inline-flex h-9 w-9 items-center justify-center rounded-full bg-black/55 text-white backdrop-blur-sm">
          <Play size={16} fill="currentColor" />
        </span>
      )}
      {post?.caption && (
        <span className="absolute bottom-3 left-3 right-3 line-clamp-2 text-xs font-medium text-white opacity-0 transition group-hover:opacity-100">
          {post.caption}
        </span>
      )}
    </Link>
  );
};

export const Users = () => {
  const { id } = useParams();
  const [user, setUser] = useState(null);
  const [relationship, setRelationship] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState(null);

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
    ? new Date(user.createdAt).toLocaleDateString(undefined, { month: "long", year: "numeric" })
    : "";
  const isSelf = Boolean(relationship?.isSelf);

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center text-slate-500">
        <Loader2 className="animate-spin" size={24} />
      </div>
    );
  }

  if (error && !user) {
    return (
      <div className="mx-auto max-w-xl p-6">
        <Link to="/" className="mb-6 inline-flex items-center gap-2 text-sm font-medium text-slate-600">
          <ArrowLeft size={17} /> Back to home
        </Link>
        <div className="rounded-3xl border border-red-100 bg-white p-8 text-center text-red-600 shadow-sm">
          {error}
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50 pb-10">
      <div className="sticky top-0 z-30 border-b border-slate-200/80 bg-white/90 backdrop-blur-xl">
        <div className="mx-auto flex h-16 w-full max-w-4xl items-center justify-between px-4 sm:px-6">
          <Link
            to="/"
            aria-label="Back to home"
            className="inline-flex h-10 w-10 items-center justify-center rounded-full text-slate-700 transition hover:bg-slate-100 hover:text-slate-950"
          >
            <ArrowLeft size={21} />
          </Link>
          <div className="flex min-w-0 items-center gap-2">
            <span className="truncate text-sm font-semibold text-slate-900">{user?.username}</span>
            {user?.status && user.status !== "free" && (
              <span className="rounded-full bg-indigo-50 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-indigo-600">
                {user.status}
              </span>
            )}
          </div>
          {isSelf ? (
            <Link
              to="/profile"
              aria-label="Profile settings"
              className="inline-flex h-10 w-10 items-center justify-center rounded-full text-slate-700 transition hover:bg-slate-100 hover:text-slate-950"
            >
              <MoreVertical size={21} />
            </Link>
          ) : (
            <button
              type="button"
              aria-label="Profile menu"
              className="inline-flex h-10 w-10 items-center justify-center rounded-full text-slate-400"
            >
              <MoreVertical size={21} />
            </button>
          )}
        </div>
      </div>

      <main className="mx-auto w-full max-w-4xl px-3 pt-3 sm:px-6 sm:pt-5">
        <section className="overflow-hidden rounded-[28px] border border-slate-200 bg-white shadow-sm">
          <div className="relative h-40 overflow-hidden bg-gradient-to-br from-slate-950 via-indigo-950 to-slate-800 sm:h-56">
            {cover && (
              <img
                src={cover}
                alt={`${user?.username} cover`}
                className="h-full w-full object-cover"
              />
            )}
            <div className="absolute inset-0 bg-gradient-to-t from-black/45 via-black/5 to-transparent" />
          </div>

          <div className="px-4 pb-5 sm:px-8 sm:pb-7">
            <div className="-mt-14 flex flex-col gap-5 sm:-mt-16 sm:flex-row sm:items-end sm:justify-between">
              <div className="flex min-w-0 items-end gap-4">
                <div className="relative h-28 w-28 shrink-0 sm:h-32 sm:w-32">
                  <div className="flex h-full w-full items-center justify-center overflow-hidden rounded-full border-4 border-white bg-slate-100 text-3xl font-bold text-slate-700 shadow-lg">
                    {avatar ? (
                      <img src={avatar} alt={`${user?.username} avatar`} className="h-full w-full object-cover" />
                    ) : (
                      user?.username?.charAt(0).toUpperCase()
                    )}
                  </div>
                  {isSelf && (
                    <Link
                      to="/profile"
                      aria-label="Edit profile picture"
                      className="absolute bottom-0 right-0 inline-flex h-10 w-10 items-center justify-center rounded-full border-4 border-white bg-slate-900 text-white shadow-md transition hover:scale-105 hover:bg-indigo-600"
                    >
                      <Camera size={17} />
                    </Link>
                  )}
                </div>

                <div className="min-w-0 pb-1">
                  <h1 className="truncate text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">{user?.username}</h1>
                  {joined && <p className="mt-1 text-xs text-slate-500">Joined {joined}</p>}
                </div>
              </div>

              {!isSelf && (
                <button
                  type="button"
                  onClick={handleFollow}
                  disabled={actionLoading || (!relationship?.following && !relationship?.allowFollowers)}
                  className={`inline-flex items-center justify-center gap-2 rounded-2xl px-5 py-2.5 text-sm font-semibold shadow-sm transition disabled:cursor-not-allowed disabled:opacity-50 ${
                    relationship?.following
                      ? "border border-slate-300 bg-white text-slate-700 hover:bg-slate-50"
                      : "bg-slate-950 text-white hover:bg-indigo-700"
                  }`}
                >
                  {actionLoading ? (
                    <Loader2 size={17} className="animate-spin" />
                  ) : relationship?.following ? (
                    <Check size={17} />
                  ) : (
                    <UserPlus size={17} />
                  )}
                  {relationship?.following
                    ? "Following"
                    : relationship?.allowFollowers
                      ? "Follow"
                      : "Followers disabled"}
                </button>
              )}
            </div>

            <div className="mt-5 max-w-3xl">
              {user?.bio && (
                <p className="whitespace-pre-line text-sm leading-6 text-slate-700">{user.bio}</p>
              )}

              {(user?.city || user?.country) && (
                <div className="mt-3 inline-flex items-center gap-1.5 text-xs font-medium text-slate-500">
                  <MapPin size={15} />
                  {[user.city, user.country].filter(Boolean).join(", ")}
                </div>
              )}

              {bioLinks.length > 0 && (
                <div className="mt-4 flex flex-wrap gap-2">
                  {bioLinks.map((url) => (
                    <a
                      key={url}
                      href={url}
                      target="_blank"
                      rel="noreferrer noopener"
                      className="inline-flex items-center gap-1.5 rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs font-semibold text-slate-700 transition hover:border-indigo-200 hover:bg-indigo-50 hover:text-indigo-700"
                    >
                      <Link2 size={13} />
                      {getLinkLabel(url)}
                      <ExternalLink size={12} />
                    </a>
                  ))}
                </div>
              )}
            </div>

            <div className="mt-6 flex flex-wrap items-center gap-2 border-t border-slate-100 pt-4 sm:gap-4">
              <Stat label="Followers" value={relationship?.followerCount} to={`/users/${id}/followers`} />
              <Stat label="Following" value={relationship?.followingCount} to={`/users/${id}/following`} />
              <div className="flex min-w-[100px] flex-col items-center px-3 py-2 text-center">
                <div className="text-lg font-bold text-slate-900">{posts.length}</div>
                <div className="text-xs font-medium text-slate-500">Posts</div>
              </div>
            </div>

            {relationship?.follower && !relationship?.following && (
              <div className="mt-3 inline-flex items-center gap-2 text-xs text-slate-500">
                <UsersIcon size={14} /> This user follows you
              </div>
            )}

            {error && <p className="mt-4 text-sm text-red-600">{error}</p>}
          </div>
        </section>

        <section className="mt-5 rounded-[28px] border border-slate-200 bg-white px-3 py-4 shadow-sm sm:px-5">
          <div className="flex items-center justify-between border-b border-slate-100 px-2 pb-4">
            <div>
              <h2 className="text-base font-bold text-slate-900">Gallery</h2>
              <p className="mt-0.5 text-xs text-slate-500">Photos and videos shared by {user?.username}</p>
            </div>
            <span className="inline-flex items-center gap-1.5 rounded-full bg-slate-100 px-2.5 py-1 text-[11px] font-semibold text-slate-600">
              <Link2 size={12} /> {posts.length}
            </span>
          </div>

          {posts.length > 0 ? (
            <div className="mt-4 grid grid-cols-3 gap-1.5 sm:gap-3">
              {posts.map((post) => (
                <MediaTile key={post.postId} post={post} />
              ))}
            </div>
          ) : (
            <div className="flex min-h-56 flex-col items-center justify-center px-6 py-12 text-center">
              <div className="mb-4 inline-flex h-14 w-14 items-center justify-center rounded-full bg-slate-100 text-slate-500">
                <Link2 size={22} />
              </div>
              <h3 className="text-sm font-bold text-slate-900">No posts yet</h3>
              <p className="mt-1 max-w-sm text-xs leading-5 text-slate-500">
                {isSelf ? "Share your first photo or video and it will appear here." : "Posts from this profile will appear here when they are shared."}
              </p>
            </div>
          )}
        </section>
      </main>
    </div>
  );
};

export default Users;
