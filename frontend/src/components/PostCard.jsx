import {
  MoreHorizontal,
  Trash2,
  Loader2,
  Heart,
  MessageSquare,
} from "lucide-react";

import { useState } from "react";

import { usePostActions } from "../hooks/usePosts";
import { useAuth } from "../context/AuthContext";



export const PostCard = ({
  post,
  onPostDeleted,
}) => {


  const { user: currentUser } = useAuth();

  const {
    deletePost,
    toggleLike,
    loading,
  } = usePostActions();



  const [showMenu,setShowMenu] = useState(false);

  const [liked,setLiked] = useState(
    post.liked ?? false
  );


  const postId =
    post.id ||
    post.postId ||
    post.ID;



  /*
    Backend sends User
    but keep fallback for old responses
  */
  const author =
    post.User ||
    post.user ||
    {};




  const isOwner =
    currentUser &&
    currentUser.id === post.userId;



  const avatar =
    author.profile_picture ||
    author.avatar ||
    "/default-avatar.png";





  const handleDelete = async()=>{


    const confirmDelete =
      window.confirm(
        "Delete this post?"
      );


    if(!confirmDelete)
      return;



    try{


      await deletePost(postId);



      onPostDeleted?.(
        postId
      );


    }catch(error){


      console.error(error);


    }
    finally{

      setShowMenu(false);

    }

  };







  const handleLike = async()=>{


    try{


      const response =
        await toggleLike(postId);



      setLiked(
        response.liked
      );


    }catch(error){

      console.error(error);

    }


  };






  return (

    <article
      className="
        bg-white
        rounded-2xl
        border
        border-slate-200
        overflow-hidden
        shadow-sm
        hover:shadow-md
        transition
        mb-6
      "
    >





      {/* Header */}

      <header
        className="
          flex
          items-center
          justify-between
          px-5
          py-4
        "
      >


        <div className="
          flex
          items-center
          gap-3
        ">


          <img

            src={avatar}

            alt={author.username}

            className="
              w-11
              h-11
              rounded-full
              object-cover
              border
            "

          />



          <div>

            <h3
              className="
                font-semibold
                text-slate-900
                text-sm
              "
            >
              {
                author.username ||
                "Anonymous"
              }
            </h3>


            <p
              className="
                text-xs
                text-slate-500
              "
            >

              {
                new Date(
                  post.createdAt
                )
                .toLocaleDateString()
              }

            </p>


          </div>


        </div>





        {
          isOwner && (

            <div className="
              relative
            ">


              <button

                onClick={()=>
                  setShowMenu(
                    previous =>
                    !previous
                  )
                }

                className="
                  p-2
                  rounded-full
                  hover:bg-slate-100
                "
              >

                {
                  loading
                  ?
                  <Loader2
                    size={18}
                    className="animate-spin"
                  />

                  :

                  <MoreHorizontal
                    size={18}
                  />
                }


              </button>





              {
                showMenu && (

                  <div
                    className="
                      absolute
                      right-0
                      mt-2
                      w-40
                      bg-white
                      border
                      rounded-xl
                      shadow-lg
                      z-20
                    "
                  >


                    <button

                      onClick={handleDelete}

                      disabled={loading}

                      className="
                        w-full
                        flex
                        items-center
                        gap-2
                        px-4
                        py-3
                        text-sm
                        text-red-600
                        hover:bg-red-50
                      "
                    >

                      <Trash2 size={15}/>

                      Delete


                    </button>


                  </div>

                )
              }


            </div>

          )
        }


      </header>









      {/* Media */}


      {
        post.contentURL && (

          <div
            className="
              bg-black
              flex
              justify-center
              max-h-[600px]
              overflow-hidden
            "
          >

            {
              post.contentType === "video"

              ?

              <video

                src={post.contentURL}

                controls

                className="
                  w-full
                  object-contain
                "

              />


              :

              <img

                src={post.contentURL}

                alt="post"

                className="
                  w-full
                  object-cover
                "

              />

            }


          </div>

        )
      }







      {/* Actions */}


      <section
        className="
          px-5
          py-4
        "
      >


        <div
          className="
            flex
            gap-5
          "
        >



          <button

            onClick={handleLike}

            className={`
              flex
              items-center
              gap-2
              text-sm
              transition
              ${
                liked
                ?
                "text-red-500"
                :
                "text-slate-600 hover:text-red-500"
              }
            `}

          >

            <Heart
              size={19}
              fill={
                liked
                ?
                "currentColor"
                :
                "none"
              }
            />


            Like


          </button>





          <button

            className="
              flex
              items-center
              gap-2
              text-sm
              text-slate-600
              hover:text-indigo-600
            "

          >

            <MessageSquare size={19}/>

            Comment


          </button>


        </div>






        {
          post.caption && (

            <p
              className="
                mt-4
                text-sm
                text-slate-800
              "
            >

              <span
                className="
                  font-semibold
                  mr-2
                "
              >

                {author.username}

              </span>


              {post.caption}


            </p>

          )
        }



      </section>




    </article>

  );

};