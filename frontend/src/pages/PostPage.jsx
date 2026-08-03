import React from "react";

import { usePosts } from "../hooks/usePosts";

import { Headerposts } from "../components/PostHeader";
import { PostCard } from "../components/PostCard";
import { Sidebar } from "../components/Sidebar";

const Posts = () => {
  const {
    posts,
    loading,
    error,
    refetch
  } = usePosts();
  return (
    <div
      className="
      h-screen
      w-screen
      bg-neutral-950
      text-neutral-100
      flex
      overflow-hidden
      "
    >
     <main
        className="
        flex-1
        flex
        flex-col
        h-full
        max-w-2xl
        mx-auto
        border-x
        border-neutral-800
        "
      >
        <Headerposts
          title="Feed"
        />

        <section
          className="
          flex-1
          overflow-y-auto
          no-scrollbar
          "
        >
          <div
            className="
            px-4
            divide-y
            divide-neutral-800
            "
          >
            {loading && (

              <div
                className="
                h-64
                flex
                items-center
                justify-center
                text-neutral-500
                text-sm
                "
              >

                Loading posts...

              </div>

            )}

            {!loading && error && (

              <div
                className="
                h-64
                flex
                flex-col
                items-center
                justify-center
                gap-4
                text-sm
                text-red-400
                "
              >
                <p>
                  {error}
                </p>
                <button
                  onClick={refetch}
                  className="
                  px-4
                  py-2
                  rounded-lg
                  bg-neutral-800
                  hover:bg-neutral-700
                  text-neutral-200
                  "
                >
                  Retry
                </button>
              </div>
            )}

            {!loading &&
            !error &&
            posts.length === 0 && (

              <div
                className="py-20 text-center  text-neutral-500  text-sm "
              >
                No posts yet.
              </div>
            )}
            {!loading &&  !error &&  posts.map(post=>(
              <PostCard  key={post.id}  post={post} />
            ))}
          </div>
        </section>
      </main>
    </div>
  );
};
export default Posts;