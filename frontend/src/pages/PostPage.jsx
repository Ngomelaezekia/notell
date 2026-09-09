import React, { useEffect, useRef, useState } from "react";
import { Compass, Flame, Loader2, MapPin, MessageCircle, Plus, Users, UsersRound } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { usePosts } from "../hooks/usePosts";
import { Headerposts } from "../components/PostHeader";
import { PostCard } from "../components/PostCard";

const FEED_CATEGORIES = [
  { id: "all", label: "All", icon: Compass },
  { id: "popular", label: "Popular", icon: Flame },
  { id: "local", label: "Local", icon: MapPin },
  { id: "following", label: "My friends", icon: UsersRound },
];

const EmptyFeed = ({ category }) => {
  const navigate = useNavigate();
  const isFollowing = category === "following";
  const isLocal = category === "local";
  return (
    <div className="flex min-h-[calc(100dvh-13rem)] flex-col items-center justify-center px-4 py-12 text-center sm:min-h-[28rem] sm:px-5 sm:py-16">
      <div className="relative mb-5 sm:mb-6"><div className="flex h-20 w-20 items-center justify-center rounded-full border border-neutral-800 bg-neutral-900 shadow-xl shadow-black/20">{isFollowing ? <UsersRound size={34} strokeWidth={1.7} className="text-neutral-400" /> : isLocal ? <MapPin size={34} strokeWidth={1.7} className="text-neutral-400" /> : <MessageCircle size={34} strokeWidth={1.7} className="text-neutral-400" />}</div><div className="absolute -bottom-1 -right-1 flex h-8 w-8 items-center justify-center rounded-full border-4 border-neutral-950 bg-neutral-100 text-neutral-950"><Plus size={16} strokeWidth={2.5} /></div></div>
      <h2 className="text-xl font-semibold tracking-tight text-neutral-100">{isFollowing ? "No friends posts yet" : isLocal ? "No local posts yet" : "Your feed is quiet"}</h2>
      <p className="mt-2 max-w-sm text-sm leading-6 text-neutral-500">{isFollowing ? "Follow more people to see their posts in this feed." : isLocal ? "Posts from people in your city or country will appear here." : "Explore posts from across Notell, or share something yourself and start the conversation."}</p>
      <div className="mt-5 flex w-full max-w-xs flex-col gap-2 sm:mt-6 sm:max-w-none sm:flex-row sm:justify-center sm:gap-2.5"><button type="button" onClick={() => navigate("/search")} className="inline-flex items-center justify-center gap-2 rounded-full bg-neutral-100 px-5 py-2.5 text-sm font-semibold text-neutral-950 transition hover:bg-white active:scale-[0.98]"><Users size={16} /> Find people</button><button type="button" onClick={() => navigate("/create-post")} className="inline-flex items-center justify-center gap-2 rounded-full border border-neutral-800 bg-neutral-900 px-5 py-2.5 text-sm font-semibold text-neutral-200 transition hover:bg-neutral-800 active:scale-[0.98]"><Plus size={16} /> Create a post</button></div>
      <button type="button" onClick={() => navigate("/")} className="mt-4 inline-flex items-center gap-1.5 text-xs font-medium text-neutral-600 transition hover:text-neutral-300 sm:mt-5"><Compass size={13} /> View all posts</button>
    </div>
  );
};

const FeedCategories = ({ category, onChange }) => (
  <nav className="border-b border-neutral-900 bg-neutral-950 px-2 py-2 sm:px-4" aria-label="Feed categories"><div className="no-scrollbar flex gap-1 overflow-x-auto">{FEED_CATEGORIES.map(({ id, label, icon: Icon }) => { const active = category === id; return <button key={id} type="button" onClick={() => onChange(id)} aria-current={active ? "page" : undefined} className={`inline-flex min-h-9 shrink-0 items-center gap-1.5 rounded-full px-3.5 text-xs font-semibold transition active:scale-[0.98] ${active ? "bg-neutral-100 text-neutral-950" : "text-neutral-500 hover:bg-neutral-900 hover:text-neutral-200"}`}><Icon size={14} /> {label}</button>; })}</div></nav>
);

const Posts = () => {
  const [category, setCategory] = useState("all");
  const { posts, loading, loadingMore, hasMore, error, refetch, loadMore, removePost } = usePosts(1, 20, category);
  const scrollContainerRef = useRef(null);
  const loadMoreRef = useRef(null);

  useEffect(() => {
    const sentinel = loadMoreRef.current;
    const scrollContainer = scrollContainerRef.current;
    if (!sentinel || !scrollContainer || !hasMore) return undefined;
    const observer = new IntersectionObserver((entries) => { if (entries[0]?.isIntersecting) void loadMore(); }, { root: scrollContainer, rootMargin: "800px 0px" });
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasMore, loadMore]);

  return (
    <div className="flex h-[calc(100dvh-5rem)] w-full overflow-hidden bg-neutral-950 text-neutral-100 md:h-screen">
      <main className="mx-auto flex h-full w-full min-w-0 max-w-4xl flex-col border-x border-neutral-800/80">
        <section ref={scrollContainerRef} className="no-scrollbar min-h-0 flex-1 overflow-y-auto overscroll-contain">
          <Headerposts />
          <FeedCategories category={category} onChange={setCategory} />
          <div className="relative mx-auto w-full max-w-3xl">
            {loading && posts.length > 0 && <div className="pointer-events-none sticky top-0 z-20 h-0.5 overflow-hidden bg-neutral-900"><div className="h-full w-1/3 animate-pulse bg-neutral-500" /></div>}
            {!loading && !error && posts.length > 0 && <div className="flex items-center justify-between px-3 pb-1 pt-3 sm:px-5 sm:pt-4 md:px-6"><span className="text-xs font-semibold uppercase tracking-[0.16em] text-neutral-600">{FEED_CATEGORIES.find((item) => item.id === category)?.label} feed</span><span className="text-[11px] text-neutral-700">{posts.length} loaded</span></div>}
            {loading && posts.length === 0 && <div className="min-h-40" aria-label="Loading feed" />}
            {!loading && error && posts.length === 0 && <div className="mx-3 my-5 flex min-h-56 flex-col items-center justify-center rounded-2xl border border-neutral-900 bg-neutral-950/70 px-5 text-center sm:mx-5 md:mx-6"><p className="max-w-sm text-sm leading-6 text-red-400">{error}</p><button type="button" onClick={() => void refetch()} className="mt-4 rounded-full bg-neutral-800 px-5 py-2 text-sm font-medium text-neutral-200 transition hover:bg-neutral-700 active:scale-[0.98]">Try again</button></div>}
            {!loading && !error && posts.length === 0 && <EmptyFeed category={category} />}
            {posts.length > 0 && <div className="space-y-2 px-2 pb-2 pt-2 sm:space-y-3 sm:px-4 sm:pb-3 sm:pt-3 md:px-6 lg:px-8">{posts.map((post, index) => <PostCard key={post.postId} post={post} priority={index < 3} onPostDeleted={removePost} />)}</div>}
            {error && posts.length > 0 && <div className="mx-2 my-2 flex items-center justify-center gap-3 rounded-xl border border-red-950/60 bg-red-950/10 px-3 py-3 text-xs text-red-400 sm:mx-4 md:mx-6 lg:mx-8"><span>{error}</span><button type="button" onClick={() => void loadMore()} className="shrink-0 rounded-full bg-neutral-800 px-3 py-1.5 font-medium text-neutral-200 transition hover:bg-neutral-700">Retry</button></div>}
            <div ref={loadMoreRef} className="flex min-h-20 items-center justify-center px-4 py-6">{loadingMore && <div className="flex items-center gap-2 rounded-full border border-neutral-900 bg-neutral-950 px-4 py-2 text-xs text-neutral-500"><Loader2 size={14} className="animate-spin" /> Loading more posts...</div>}{!loadingMore && !hasMore && posts.length > 0 && <div className="flex items-center gap-2 text-[11px] text-neutral-700"><span className="h-px w-8 bg-neutral-900" /> You&apos;re all caught up. <span className="h-px w-8 bg-neutral-900" /></div>}</div>
          </div>
        </section>
      </main>
    </div>
  );
};

export default Posts;
