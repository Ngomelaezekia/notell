import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { Search, UserRound } from "lucide-react";
import { userAPI } from "../services/user/userApi";
import { getFileUrl } from "../utils/api";

export default function SearchPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [query, setQuery] = useState(searchParams.get("q") || "");
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [searched, setSearched] = useState(Boolean(searchParams.get("q")));

  useEffect(() => {
    const value = searchParams.get("q")?.trim() || "";
    setQuery(value);
    if (value.length < 2) {
      setUsers([]);
      setSearched(false);
      setError(null);
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError(null);
    userAPI.searchUsers(value)
      .then((response) => {
        if (cancelled) return;
        setUsers(response?.data?.users || []);
        setSearched(true);
      })
      .catch((err) => {
        if (cancelled) return;
        setUsers([]);
        setError(err?.response?.data?.message || "Search failed. Please try again.");
        setSearched(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => { cancelled = true; };
  }, [searchParams]);

  const submitSearch = (event) => {
    event.preventDefault();
    const value = query.trim();
    if (value.length < 2) {
      setSearchParams({});
      return;
    }
    setSearchParams({ q: value });
  };

  return (
    <section className="mx-auto w-full max-w-3xl">
      <div className="rounded-3xl border border-white/70 bg-white/80 p-5 shadow-xl backdrop-blur-xl sm:p-7">
        <div className="mb-6">
          <h1 className="text-2xl font-black text-slate-900">Discover people</h1>
          <p className="mt-1 text-sm text-slate-500">Find people by username, bio, city, or country.</p>
        </div>

        <form onSubmit={submitSearch} className="relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400" size={20} />
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search people..."
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

        <div className="mt-6">
          {loading && <p className="py-8 text-center text-sm text-slate-500">Searching...</p>}
          {error && !loading && <p className="rounded-2xl bg-red-50 p-4 text-sm text-red-600">{error}</p>}
          {!loading && !error && searched && users.length === 0 && (
            <div className="py-12 text-center">
              <UserRound className="mx-auto text-slate-300" size={42} />
              <p className="mt-3 font-semibold text-slate-700">No people found</p>
              <p className="mt-1 text-sm text-slate-500">Try a different search term.</p>
            </div>
          )}

          {!loading && !error && users.length > 0 && (
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
                      <p className="truncate text-xs text-slate-500">{[user.city, user.country].filter(Boolean).join(", ")}</p>
                    )}
                    {user.bio && <p className="mt-1 truncate text-sm text-slate-600">{user.bio}</p>}
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      </div>
    </section>
  );
}
