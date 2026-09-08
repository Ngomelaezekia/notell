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

const Stat = ({ label, value, onClick }) => (
  <button
    type="button"
    onClick={onClick}
    className="group min-w-[92px] rounded-xl px-2 py-1.5 text-center transition hover:bg-slate-50"
  >
    <div className="text-base font-bold text-slate-900 group-hover:text-indigo-600">{value ?? 0}</div>
    <div className="text-[11px] font-medium text-slate-500">{label}</div>
  </button>
);

const MediaTile = ({ post }) => {
  const mediaUrl = getFileUrl(post?.contentUrl);
  const isVideo = post?.contentType === "video";

  return (
    <Link
      to={`/posts/${post?.postId}`}
      className="group relative aspect-square overflow-hidden rounded-xl bg-slate-100 shadow-sm ring-1 ring-slate-200 transition hover:-translate-y-0.5 hover:shadow-md"
    >
      {isVideo ? (
        <video src={mediaUrl} muted playsInline preload="metadata" className="h-full w-full object-cover transition duration-300 group-hover:scale-105" />
      ) : (
        <img src={mediaUrl} alt={post?.caption || "Post"} loading="lazy" className="h-full w-full object-cover transition duration-300 group-hover:scale-105" />
      )}
      <div className="absolute inset-0 bg-gradient-to-t from-black/45 via-transparent to-transparent opacity-0 transition group-hover:opacity-100" />
      {isVideo && <span className="absolute right-2.5 top-2.5 inline-flex h-8 w-8 items-center justify-center rounded-full bg-black/55 text-white backdrop-blur-sm"><Play size={14} fill="currentColor" /></span>}
      {post?.caption && <span className="absolute bottom-2.5 left-2.5 right-2.5 line-clamp-2 text-[11px] font-medium text-white opacity-0 transition group-hover:opacity-100">{post.caption}</span>}
    </Link>
  );
};

const RelationshipPanel = ({ type, users, loading, error, onClose }) => {
  const title = type === "followers" ? "Followers" : "Following";
  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-slate-950/35 p-0 backdrop-blur-[2px] sm:items-center sm:p-5" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <section className="flex max-h-[78vh] w-full flex-col overflow-hidden rounded-t-3xl border border-slate-200 bg-white shadow-2xl sm:max-w-md sm:rounded-3xl" role="dialog" aria-modal="true" aria-label={title}>
        <header className="flex items-center justify-between border-b border-slate-100 px-5 py-4">
          <div>
            <h2 className="text-base font-bold text-slate-950">{title}</h2>
            <p className="mt-0.5 text-[11px] text-slate-500">People connected to this profile</p>
          </div>
          <button type="button" onClick={onClose} aria-label="Close" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-slate-500 transition hover:bg-slate-100 hover:text-slate-900"><X size={18} /></button>
        </header>

        {loading ? (
          <div className="flex min-h-52 items-center justify-center text-slate-500"><Loader2 size={22} className="animate-spin" /></div>
        ) : error && users.length === 0 ? (
          <div className="p-8 text-center text-sm text-red-600">{error}</div>
        ) : users.length === 0 ? (
          <div className="flex min-h-52 flex-col items-center justify-center p-8 text-center"><UserRound size={32} className="text-slate-300" /><p className="mt-3 text-sm font-medium text-slate-600">No {type} yet</p></div>
        ) : (
          <div className="overflow-y-auto">
            {error && <div className="border-b border-red-100 bg-red-50 px-5 py-3 text-xs text-red-600">{error}</div>}
            <div className="divide-y divide-slate-100">
              {users.map((person) => {
                const avatar = getFileUrl(person.profilePicture);
                return (
                  <Link key={person.id} to={`/users/${person.id}`} onClick={onClose} className="flex items-center gap-3 px-5 py-3.5 transition hover:bg-slate-50">
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
      const [profileResponse, relationshipResponse] = await Promise.all([userAPI.getProfile(id), userAPI.getRelationship(id)]);
      setUser(profileResponse?.data?.user ?? null);
      setRelationship(relationshipResponse?.data ?? null);
    } catch (err) {
      setError(err.response?.data?.message || "Failed to load profile");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { loadProfile(); }, [loadProfile]);

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
        setRelationship((current) => ({ ...current, following: false, followerCount: Math.max(0, (current?.followerCount ?? 1) - 1) }));
      } else {
        await userAPI.followUser(id);
        setRelationship((current) => ({ ...current, following: true, followerCount: (current?.followerCount ?? 0) + 1 }));
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
  const joined = user?.createdAt ? new Date(user.createdAt).toLocaleDateString(undefined, { month: "long", year: "numeric" }) : "";
  const isSelf = Boolean(relationship?.isSelf);

  if (loading) return <div className="flex min-h-[60vh] items-center justify-center text-slate-500"><Loader2 className="animate-spin" size={24} /></div>;

  if (error && !user) {
    return <div className="mx-auto max-w-xl p-6"><Link to="/" className="mb-5 inline-flex items-center gap-2 text-sm font-medium text-slate-600"><ArrowLeft size={17} /> Back to home</Link><div className="rounded-3xl border border-red-100 bg-white p-8 text-center text-red-600 shadow-sm">{error}</div></div>;
  }

  return (
    <div className="min-h-screen bg-slate-50 pb-8">
      <div className="sticky top-0 z-30 border-b border-slate-200/80 bg-white/90 backdrop-blur-xl">
        <div className="mx-auto flex h-14 w-full max-w-4xl items-center justify-between px-3 sm:px-6">
          <Link to="/" aria-label="Back to home" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-slate-700 transition hover:bg-slate-100 hover:text-slate-950"><ArrowLeft size={20} /></Link>
          <div className="flex min-w-0 items-center gap-2"><span className="truncate text-sm font-semibold text-slate-900">{user?.username}</span>{user?.status && user.status !== "free" && <span className="rounded-full bg-indigo-50 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-indigo-600">{user.status}</span>}</div>
          {isSelf ? <Link to="/settings" aria-label="Website settings" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-slate-700 transition hover:bg-slate-100 hover:text-slate-950"><MoreVertical size={20} /></Link> : <button type="button" aria-label="Profile menu" className="inline-flex h-9 w-9 items-center justify-center rounded-full text-slate-400"><MoreVertical size={20} /></button>}
        </div>
      </div>

      <main className="mx-auto w-full max-w-4xl px-2.5 pt-2 sm:px-6 sm:pt-4">
        <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
          <div className="relative h-24 overflow-hidden bg-gradient-to-br from-slate-950 via-indigo-950 to-slate-800 sm:h-32">{cover && <img src={cover} alt={`${user?.username} cover`} className="h-full w-full object-cover" />}<div className="absolute inset-0 bg-gradient-to-t from-black/40 via-black/5 to-transparent" /></div>

          <div className="px-4 pb-4 sm:px-7 sm:pb-5">
            <div className="-mt-10 flex flex-col gap-3 sm:-mt-12 sm:flex-row sm:items-end sm:justify-between">
              <div className="flex min-w-0 items-end gap-3">
                <div className="relative h-20 w-20 shrink-0 sm:h-24 sm:w-24">
                  <div className="flex h-full w-full items-center justify-center overflow-hidden rounded-full border-4 border-white bg-slate-100 text-2xl font-bold text-slate-700 shadow-md">{avatar ? <img src={avatar} alt={`${user?.username} avatar`} className="h-full w-full object-cover" /> : user?.username?.charAt(0).toUpperCase()}</div>
                  {isSelf && <Link to="/profile" aria-label="Edit profile picture" className="absolute bottom-[-2px] right-[-2px] inline-flex h-8 w-8 items-center justify-center rounded-full border-2 border-white bg-slate-900 text-white shadow transition hover:scale-105 hover:bg-indigo-600"><Camera size={14} /></Link>}
                </div>
                <div className="min-w-0 pb-0.5"><h1 className="truncate text-xl font-bold tracking-tight text-slate-950 sm:text-2xl">{user?.username}</h1>{joined && <p className="mt-0.5 text-[11px] text-slate-500">Joined {joined}</p>}</div>
              </div>

              {!isSelf && <button type="button" onClick={handleFollow} disabled={actionLoading || (!relationship?.following && !relationship?.allowFollowers)} className={`inline-flex items-center justify-center gap-2 rounded-xl px-4 py-2 text-xs font-semibold shadow-sm transition disabled:cursor-not-allowed disabled:opacity-50 ${relationship?.following ? "border border-slate-300 bg-white text-slate-700 hover:bg-slate-50" : "bg-slate-950 text-white hover:bg-indigo-700"}`}>{actionLoading ? <Loader2 size={15} className="animate-spin" /> : relationship?.following ? <Check size={15} /> : <UserPlus size={15} />}{relationship?.following ? "Following" : relationship?.allowFollowers ? "Follow" : "Followers disabled"}</button>}
            </div>

            <div className="mt-3 max-w-3xl">
              {user?.bio && <p className="whitespace-pre-line text-sm leading-5 text-slate-700">{user.bio}</p>}
              {(user?.city || user?.country) && <div className="mt-2 inline-flex items-center gap-1.5 text-[11px] font-medium text-slate-500"><MapPin size={13} />{[user.city, user.country].filter(Boolean).join(", ")}</div>}
              {bioLinks.length > 0 && <div className="mt-2 flex flex-wrap gap-1.5">{bioLinks.map((url) => <a key={url} href={url} target="_blank" rel="noreferrer noopener" className="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2.5 py-1 text-[11px] font-semibold text-slate-700 transition hover:border-indigo-200 hover:bg-indigo-50 hover:text-indigo-700"><Link2 size={12} />{getLinkLabel(url)}<ExternalLink size={10} /></a>)}</div>}
            </div>

            <div className="mt-3 flex flex-wrap items-center gap-1 border-t border-slate-100 pt-2.5 sm:gap-2">
              <Stat label="Followers" value={relationship?.followerCount} onClick={() => void openRelationshipPanel("followers")} />
              <Stat label="Following" value={relationship?.followingCount} onClick={() => void openRelationshipPanel("following")} />
              <div className="flex min-w-[92px] flex-col items-center px-2 py-1.5 text-center"><div className="text-base font-bold text-slate-900">{posts.length}</div><div className="text-[11px] font-medium text-slate-500">Posts</div></div>
            </div>

            {relationship?.follower && !relationship?.following && <div className="mt-2 inline-flex items-center gap-1.5 text-[11px] text-slate-500"><UsersIcon size={13} /> This user follows you</div>}
            {error && <p className="mt-3 text-xs text-red-600">{error}</p>}
          </div>
        </section>

        <section className="mt-4 rounded-3xl border border-slate-200 bg-white px-2.5 py-3.5 shadow-sm sm:px-5">
          <div className="flex items-center justify-between border-b border-slate-100 px-1.5 pb-3"><div><h2 className="text-sm font-bold text-slate-900">Gallery</h2><p className="mt-0.5 text-[11px] text-slate-500">Photos and videos shared by {user?.username}</p></div><span className="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-600"><Link2 size={11} /> {posts.length}</span></div>
          {posts.length > 0 ? <div className="mt-3 grid grid-cols-3 gap-1 sm:gap-2.5">{posts.map((post) => <MediaTile key={post.postId} post={post} />)}</div> : <div className="flex min-h-48 flex-col items-center justify-center px-6 py-10 text-center"><div className="mb-3 inline-flex h-12 w-12 items-center justify-center rounded-full bg-slate-100 text-slate-500"><Link2 size={20} /></div><h3 className="text-sm font-bold text-slate-900">No posts yet</h3><p className="mt-1 max-w-sm text-[11px] leading-5 text-slate-500">{isSelf ? "Share your first photo or video and it will appear here." : "Posts from this profile will appear here when they are shared."}</p></div>}
        </section>
      </main>

      {relationshipPanel && <RelationshipPanel type={relationshipPanel} users={relationshipUsers} loading={relationshipLoading} error={relationshipError} onClose={() => setRelationshipPanel(null)} />}
    </div>
  );
};

export default Users;
