import { useCallback, useEffect, useState } from "react";
import { ArrowLeft, Check, Loader2, MapPin, UserPlus, Users } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { userAPI } from "../services/user/userApi";
import { getFileUrl } from "../utils/api";

const Stat = ({ label, value }) => (
  <div className="min-w-24 text-center">
    <div className="text-lg font-bold text-slate-900">{value ?? 0}</div>
    <div className="text-xs text-slate-500">{label}</div>
  </div>
);

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

  useEffect(() => { loadProfile(); }, [loadProfile]);

  const handleFollow = async () => {
    if (!id || actionLoading || !relationship?.allowFollowers) return;
    setActionLoading(true);
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

  if (loading) {
    return <div className="flex min-h-[60vh] items-center justify-center text-slate-500"><Loader2 className="animate-spin" /></div>;
  }

  if (error && !user) {
    return <div className="mx-auto max-w-xl p-6"><Link to="/" className="mb-6 inline-flex items-center gap-2 text-sm text-slate-600"><ArrowLeft size={16} /> Back</Link><div className="rounded-2xl border bg-white p-8 text-center text-red-600">{error}</div></div>;
  }

  const avatar = getFileUrl(user?.profilePicture);
  const joined = user?.createdAt ? new Date(user.createdAt).toLocaleDateString(undefined, { month: "long", year: "numeric" }) : "";

  return (
    <div className="mx-auto w-full max-w-3xl p-4 sm:p-6">
      <Link to="/" className="mb-5 inline-flex items-center gap-2 text-sm font-medium text-slate-600 hover:text-slate-900"><ArrowLeft size={16} /> Back</Link>

      <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
        <div className="h-28 bg-gradient-to-r from-slate-900 via-indigo-900 to-slate-800 sm:h-36" />
        <div className="px-5 pb-6 sm:px-8">
          <div className="-mt-12 flex flex-col gap-4 sm:-mt-14 sm:flex-row sm:items-end sm:justify-between">
            <div className="flex items-end gap-4">
              <div className="flex h-24 w-24 shrink-0 items-center justify-center overflow-hidden rounded-full border-4 border-white bg-slate-100 text-2xl font-bold text-slate-700 shadow">
                {avatar ? <img src={avatar} alt={`${user?.username} avatar`} className="h-full w-full object-cover" /> : user?.username?.charAt(0).toUpperCase()}
              </div>
              <div className="pb-1">
                <h1 className="text-xl font-bold text-slate-900">{user?.username}</h1>
                {joined && <p className="text-xs text-slate-500">Joined {joined}</p>}
              </div>
            </div>

            {!relationship?.isSelf && (
              <button type="button" onClick={handleFollow} disabled={actionLoading || !relationship?.allowFollowers} className={`inline-flex items-center justify-center gap-2 rounded-xl px-5 py-2.5 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-50 ${relationship?.following ? "border border-slate-300 bg-white text-slate-700 hover:bg-slate-50" : "bg-slate-900 text-white hover:bg-slate-800"}`}>
                {actionLoading ? <Loader2 size={16} className="animate-spin" /> : relationship?.following ? <Check size={16} /> : <UserPlus size={16} />}
                {relationship?.following ? "Following" : relationship?.allowFollowers ? "Follow" : "Followers disabled"}
              </button>
            )}
          </div>

          {user?.bio && <p className="mt-5 max-w-2xl text-sm leading-6 text-slate-700">{user.bio}</p>}
          {(user?.city || user?.country) && <div className="mt-3 flex items-center gap-1.5 text-xs text-slate-500"><MapPin size={14} /> {[user.city, user.country].filter(Boolean).join(", ")}</div>}

          <div className="mt-6 flex items-center gap-8 border-t border-slate-100 pt-5">
            <Stat label="Followers" value={relationship?.followerCount} />
            <Stat label="Following" value={relationship?.followingCount} />
          </div>

          {relationship?.follower && !relationship?.following && <div className="mt-4 flex items-center gap-2 text-xs text-slate-500"><Users size={14} /> This user follows you</div>}
          {error && <p className="mt-4 text-sm text-red-600">{error}</p>}
        </div>
      </section>
    </div>
  );
};

export default Users;
