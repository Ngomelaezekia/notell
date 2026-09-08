import React, { useEffect, useRef } from "react";
import { Loader2, MessageCircle, Users, Compass, Plus } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { usePosts } from "../hooks/usePosts";
import { Headerposts } from "../components/PostHeader";
import { PostCard } from "../components/PostCard";

const EmptyFeed = () => {
  const navigate = useNavigate();

  return (
    <div className="flex min-h-[calc(100dvh-9rem)] flex-col items-center justify-center px-5 py-16 text-center sm:min-h-[28rem]">
      <div className="relative mb-6">
        <div className="flex h-20 w-20 items-center justify-center rounded-full border border-neutral-800 bg-neutral-900 shadow-xl shadow-black/20">
          <MessageCircle size={34} strokeWidth={1.7} className="text-neutral-400" />
        </div>
        <div className="absolute -bottom-1 -right-1 flex h-8 w-8 items-center justify-center rounded-full border-4 border-neutral-950 bg-neutral-100 text-neutral-950">
          <Plus size={16} strokeWidth={2.5} />
        </div>
      </div>

      <h2 className="text-xl font-semibold tracking-tight text-neutral-100">
        Your feed is quiet
      </h2>
      <p className="mt-2 max-w-sm text-sm leading-6 text-neutral-500">
        Follow people and communities to see their posts here, or share
        something yourself and start the conversation.
      </p>

      <div className="mt-6 flex w-full max-w-xs flex-col gap-2.5 sm:flex-row sm:max-w-none sm:justify-center">
        <button
          type="button"
          onClick={() => navigate("/search")}
          className="inline-flex items-center justify-center gap-2 rounded-full bg-neutral-100 px-5 py-2.5 text-sm font-semibold text-neutral-950 transition hover:bg-white active:scale-[0.98]"
        >
          <Users size={16} /> Find people
        </button>
        <button
          type="button"
          onClick={() => navigate("/create-post")}
          className="inline-flex items-center justify-center gap-2 rounded-full border border-neutral-800 bg-neutral-900 px-5 py-2.5 text-sm font-semibold text-neutral-200 transition hover:bg-neutral-800 active:scale-[0.98]"
        >
          <Plus size={16} /> Create a post
        </button>
      </div>

      <button
        type="button"
        onClick={() => navigate("/search")}
        className="mt-5 inline-flex items-center gap-1.5 text-xs font-medium text-neutral-600 transition hover:text-neutral-300"
      >
        <Compass size={13} /> Explore Notell
      </button>
    </div>
  );
};

const Posts = () => {
  const {
    posts,
    loading,
    loadingMore,
    hasMore,
    error,
    refetch,
    loadMore,
    removePost,
  } = usePosts();
  const scrollContainerRef = useRef(null);
  const loadMoreRef = useRef(null);

  useEffect(() => {
    const sentinel = loadMoreRef.current;
    const scrollContainer = scrollContainerRef.current;
    if (!sentinel || !scrollContainer || !hasMore) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) loadMore();
      },
      {
        root: scrollContainer,
        rootMargin: "500px 0px",
      }
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasMore, loadMore]);

  return (
    <div className="flex h-[calc(100dvh-5rem)] w-full overflow-hidden bg-neutral-950 text-neutral-100 md:h-screen">
      <main className="mx-auto flex h-full w-full min-w-0 max-w-3xl flex-col border-x border-neutral-800">
        <Headerposts title="Feed" />

        <section
          ref={scrollContainerRef}
          className="no-scrollbar min-h-0 flex-1 overflow-y-auto"
        >
          <div className="mx-auto w-full max-w-2xl px-2 py-2 sm:px-4 sm:py-4 md:px-5 md:py-5">
            {loading && (
              <div className="flex h-64 items-center justify-center text-sm text-neutral-500">
                Loading posts...
              </div>
            )}

            {!loading && error && posts.length === 0 && (
              <div className="flex h-64 flex-col items-center justify-center gap-4 px-4 text-center text-sm text-red-400">
                <p>{error}</p>
                <button
                  type="button"
                  onClick={refetch}
                  className="rounded-full bg-neutral-800 px-5 py-2 text-neutral-200 transition hover:bg-neutral-700"
                >
                  Try again
                </button>
              </div>
            )}

            {!loading && !error && posts.length === 0 && <EmptyFeed />}

            {!loading && posts.map((post) => (
              <PostCard key={post.postId} post={post} onPostDeleted={removePost} />
            ))}

            {error && posts.length > 0 && (
              <div className="flex items-center justify-center gap-3 py-6 text-sm text-red-400">
                <span>{error}</span>
                <button
                  type="button"
                  onClick={loadMore}
                  className="rounded-full bg-neutral-800 px-3 py-1.5 text-neutral-200 transition hover:bg-neutral-700"
                >
                  Retry
                </button>
              </div>
            )}

            <div ref={loadMoreRef} className="flex min-h-16 items-center justify-center py-5">
              {loadingMore && (
                <div className="flex items-center gap-2 text-sm text-neutral-500">
                  <Loader2 size={16} className="animate-spin" />
                  Loading more posts...
                </div>
              )}
              {!loadingMore && !hasMore && posts.length > 0 && (
                <span className="text-xs text-neutral-600">You&apos;re all caught up.</span>
              )}
            </div>
          </div>
        </section>
      </main>
    </div>
  );
};

export default Posts;
