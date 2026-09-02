import React from "react";
import { usePosts } from "../hooks/usePosts";
import { Headerposts } from "../components/PostHeader";
import { PostCard } from "../components/PostCard";

const Posts = () => {
  const { posts, loading, error, refetch } = usePosts();

  return (
    <div className="h-screen w-screen overflow-hidden bg-neutral-950 text-neutral-100 flex">
      <main className="mx-auto flex h-full w-full max-w-2xl flex-col border-x border-neutral-800">
        <Headerposts title="Feed" />
        <section className="no-scrollbar flex-1 overflow-y-auto">
          <div className="divide-y divide-neutral-800 px-4">
            {loading && <div className="flex h-64 items-center justify-center text-sm text-neutral-500">Loading posts...</div>}

            {!loading && error && (
              <div className="flex h-64 flex-col items-center justify-center gap-4 text-sm text-red-400">
                <p>{error}</p>
                <button type="button" onClick={refetch} className="rounded-lg bg-neutral-800 px-4 py-2 text-neutral-200 hover:bg-neutral-700">Retry</button>
              </div>
            )}

            {!loading && !error && posts.length === 0 && (
              <div className="py-20 text-center text-sm text-neutral-500">No posts yet.</div>
            )}

            {!loading && !error && posts.map((post) => (
              <PostCard key={post.postId} post={post} />
            ))}
          </div>
        </section>
      </main>
    </div>
  );
};

export default Posts;
