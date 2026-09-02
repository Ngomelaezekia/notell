import React, { useEffect, useRef } from "react";
import { Loader2 } from "lucide-react";
import { usePosts } from "../hooks/usePosts";
import { Headerposts } from "../components/PostHeader";
import { PostCard } from "../components/PostCard";

const Posts = () => {
  const { posts, loading, loadingMore, hasMore, error, refetch, loadMore } = usePosts();
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
    <div className="h-screen w-screen overflow-hidden bg-neutral-950 text-neutral-100 flex">
      <main className="mx-auto flex h-full w-full max-w-2xl flex-col border-x border-neutral-800">
        <Headerposts title="Feed" />
        <section ref={scrollContainerRef} className="no-scrollbar flex-1 overflow-y-auto">
          <div className="divide-y divide-neutral-800 px-4">
            {loading && (
              <div className="flex h-64 items-center justify-center text-sm text-neutral-500">
                Loading posts...
              </div>
            )}

            {!loading && error && posts.length === 0 && (
              <div className="flex h-64 flex-col items-center justify-center gap-4 text-sm text-red-400">
                <p>{error}</p>
                <button
                  type="button"
                  onClick={refetch}
                  className="rounded-lg bg-neutral-800 px-4 py-2 text-neutral-200 hover:bg-neutral-700"
                >
                  Retry
                </button>
              </div>
            )}

            {!loading && !error && posts.length === 0 && (
              <div className="py-20 text-center text-sm text-neutral-500">
                No posts yet.
              </div>
            )}

            {!loading && posts.map((post) => (
              <PostCard key={post.postId} post={post} />
            ))}

            {error && posts.length > 0 && (
              <div className="flex items-center justify-center gap-3 py-6 text-sm text-red-400">
                <span>{error}</span>
                <button
                  type="button"
                  onClick={loadMore}
                  className="rounded-lg bg-neutral-800 px-3 py-1.5 text-neutral-200 hover:bg-neutral-700"
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
