import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { X, UploadCloud, Loader2,} from "lucide-react";
import { usePostActions } from "../hooks/usePosts";
import { uploadAPI } from "../services/post/uploadApi";

const MAX_FILE_SIZE = 50 * 1024 * 1024;

export const CreatePost = () => {
  const navigate = useNavigate();
  const {createPost, loading,  error,} = usePostActions();
  const [selectedFile,setSelectedFile] = useState(null);
  const [previewURL,setPreviewURL] = useState("");
  const [contentType,setContentType] =  useState(null);
  const [caption,setCaption] = useState("");
  const [localError,setLocalError] =    useState("");
  const [dragActive,setDragActive] =  useState(false);
  const [uploading,setUploading] =    useState(false);

    useEffect(()=>{
    return ()=>{   
       if(
        previewURL &&  previewURL.startsWith("blob:")
      ){
        URL.revokeObjectURL(previewURL);
    }
    };
  },[previewURL]);
  const validateFile = (file)=>{
    if(
      !file.type.startsWith("image/") &&
      !file.type.startsWith("video/")
    ){
      throw new Error(
        "Only images and videos are allowed"
      );
    }
    if(file.size > MAX_FILE_SIZE){

      throw new Error(
        "File size must be below 50MB"
      );
    }
  };

  const processFile = useCallback(
    (file)=>{
      try{
        validateFile(file);
        if(
          previewURL &&
          previewURL.startsWith("blob:")
        ){
          URL.revokeObjectURL(previewURL);
        }
        const isVideo =
          file.type.startsWith("video/");
        setContentType(
          isVideo
          ? "video"
          : "image"
        );
        setSelectedFile(file);
        setPreviewURL(
          URL.createObjectURL(file)
        );
        setLocalError("");

      }catch(error){

        setLocalError(
          error.message
        );
      }
    },
    [previewURL]
  );

  const handleFileSelect=(event)=>{
    const file = event.target.files?.[0];
    if(file){
      processFile(file);
    }
  };
  const handleDrop=(event)=>{

    event.preventDefault();
    setDragActive(false);
    const file =
      event.dataTransfer.files?.[0];
    if(file){
      processFile(file);
    }
  };

  const clearMedia=()=>{

    if(previewURL && previewURL.startsWith("blob:")
    ){
      URL.revokeObjectURL(previewURL);
    }
    setSelectedFile(null);
    setPreviewURL("");
    setContentType(null);
  };

  const handleSubmit = async(event)=>{
    event.preventDefault();
    setLocalError("");
    if(!selectedFile){
      setLocalError(
        "Please select an image or video."
      );
      return;
    }
    try{
       setUploading(true);
         
       const uploadResponse =
        await uploadAPI.uploadMedia(
          selectedFile
        );
      if(
        !uploadResponse?.url
      ){
        throw new Error(
          "Media upload failed"
        );
      }

  const mediaURL = uploadResponse.url.startswith("http") ? uploadResponse.url : `${import.meta.env.REACT_API_URL}${uploadResponse.url}`;
      await createPost({
        contentType,
        contentUrl:
          mediaURL,
        caption:
          caption.trim(),

      });
      navigate("/");
    }catch(error){
      setLocalError(
        error.message ||
        "Failed to publish post"
      );

    }finally{
      setUploading(false);
    }
  };

  const isSubmitting =
    loading || uploading;
  return (
    <div
      className="
        max-w-2xl        mx-auto    mt-10    p-6   rounded-2xl bg-white/80   backdrop-blur-xl   border  border-slate-200   shadow-xl
      "
    >


      <div className="mb-6">


        <h2
          className="
            text-2xl
            font-bold
            text-slate-900
          "
        >
          Create New Post
        </h2>



        <p
          className="
            text-sm
            text-slate-500
            mt-1
          "
        >
          Share images or videos with your community.
        </p>


      </div>









      {(localError || error) && (

        <div
          className="
            mb-5
            p-3
            rounded-xl
            bg-red-50
            border
            border-red-200
            text-red-700
            text-sm
          "
        >

          {localError || error}

        </div>

      )}









      <form
        onSubmit={handleSubmit}
        className="space-y-6"
      >







        <div>


          <label
            className="
              text-sm
              font-medium
              text-slate-700
            "
          >
            Media
          </label>





          {
            previewURL ? (

              <div
                className="
                  mt-3
                  relative
                  rounded-xl
                  overflow-hidden
                  bg-black
                  border
                  border-slate-200
                "
              >


                {
                  contentType === "video"
                  ?

                  (
                    <video
                      src={previewURL}
                      controls
                      className="
                        w-full
                        max-h-96
                        object-contain
                      "
                    />
                  )

                  :

                  (
                    <img
                      src={previewURL}
                      alt="Preview"
                      className="
                        w-full
                        max-h-96
                        object-contain
                      "
                    />
                  )

                }





                <button

                  type="button"

                  onClick={clearMedia}

                  className="
                    absolute
                    top-3
                    right-3
                    p-2
                    rounded-full
                    bg-black/70
                    text-white
                    hover:bg-black
                  "

                >

                  <X size={18}/>

                </button>



              </div>


            )

            :

            (

              <label

                onDragOver={(e)=>{

                  e.preventDefault();

                  setDragActive(true);

                }}


                onDragLeave={()=>{

                  setDragActive(false);

                }}


                onDrop={handleDrop}



                className={`
                  mt-3
                  h-56
                  flex
                  flex-col
                  items-center
                  justify-center
                  rounded-xl
                  border-2
                  border-dashed
                  cursor-pointer
                  transition

                  ${
                    dragActive
                    ?
                    "border-indigo-500 bg-indigo-50"
                    :
                    "border-slate-300 bg-slate-50 hover:bg-slate-100"
                  }

                `}

              >


                <UploadCloud
                  className="
                    w-12
                    h-12
                    text-slate-400
                    mb-3
                  "
                />



                <p
                  className="
                    text-sm
                    text-slate-600
                  "
                >

                  Drop file here or click upload

                </p>



                <p
                  className="
                    text-xs
                    text-slate-400
                    mt-1
                  "
                >

                  PNG JPG MP4 MOV • Max 50MB

                </p>




                <input

                  type="file"

                  accept="image/*,video/*"

                  onChange={handleFileSelect}

                  hidden

                />



              </label>

            )

          }


        </div>









        <div>


          <label
            className="
              text-sm
              font-medium
              text-slate-700
            "
          >

            Caption

          </label>



          <textarea

            rows="4"

            value={caption}

            onChange={(e)=>
              setCaption(
                e.target.value
              )
            }


            placeholder="Write something..."


            className="
              mt-2
              w-full
              rounded-xl
              border
              border-slate-300
              p-3
              text-sm
              outline-none
              focus:ring-2
              focus:ring-indigo-500
            "

          />


        </div>









        <div
          className="
            flex
            justify-end
            gap-3
          "
        >


          <button

            type="button"

            onClick={()=>
              navigate("/")
            }

            className="
              px-5
              py-2.5
              rounded-xl
              text-sm
              text-slate-600
              hover:bg-slate-100
            "

          >

            Cancel

          </button>





          <button

            type="submit"

            disabled={
              isSubmitting ||
              !selectedFile
            }


            className="
              flex
              items-center
              gap-2
              px-6
              py-2.5
              rounded-xl
              bg-indigo-600
              text-white
              font-medium
              hover:bg-indigo-700
              disabled:opacity-50
              disabled:cursor-not-allowed
            "

          >


            {
              isSubmitting &&
              (
                <Loader2
                  size={18}
                  className="animate-spin"
                />
              )
            }



            {
              uploading
              ?
              "Uploading..."
              :
              loading
              ?
              "Publishing..."
              :
              "Publish Post"
            }


          </button>



        </div>




      </form>



    </div>

  );

};