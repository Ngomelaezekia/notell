import { Loader2 } from "lucide-react";


const GoogleIcon = () => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 48 48"
    className="h-4 w-4 shrink-0"
    aria-hidden="true"
  >
    <path
      fill="#FFC107"
      d="M43.611 20.083H42V20H24v8h11.303C33.651 32.657 29.204 36 24 36c-6.627 0-12-5.373-12-12S17.373 12 24 12c3.059 0 5.842 1.154 7.955 3.045l5.657-5.657C34.046 6.053 29.277 4 24 4 12.954 4 4 12.954 4 24s8.954 20 20 20 20-8.954 20-20c0-1.341-.138-2.65-.389-3.917z"
    />

    <path
      fill="#FF3D00"
      d="M6.306 14.691l6.571 4.819C14.655 16.108 18.961 13 24 13c3.059 0 5.842 1.154 7.955 3.045l5.657-5.657C34.046 6.053 29.277 4 24 4c-7.732 0-14.41 4.388-17.694 10.691z"
    />

    <path
      fill="#4CAF50"
      d="M24 44c5.177 0 9.862-1.977 13.409-5.192l-6.191-5.238C29.149 35.091 26.671 36 24 36c-5.183 0-9.619-3.317-11.284-7.946l-6.522 5.025C9.435 39.556 16.227 44 24 44z"
    />

    <path
      fill="#1976D2"
      d="M43.611 20.083H42V20H24v8h11.303c-1.102 3.116-3.466 5.576-6.591 6.932l6.191 5.238C38.472 37.004 44 31.055 44 24c0-1.341-.138-2.65-.389-3.917z"
    />

  </svg>
);



export default function GoogleButton({
  onClick,
  loading = false,
  disabled = false,
  text = "Continue with Google",
  className = "",
}) {


  const isDisabled = disabled || loading;



  return (

    <button
      type="button"
      onClick={onClick}
      disabled={isDisabled}
      aria-label={text}
      className={`
        group
        relative
        w-full
        h-11
        rounded-xl

        border
        border-white/20

        bg-white/10

        backdrop-blur-xl

        text-white

        text-sm
        font-medium

        flex
        items-center
        justify-center
        gap-3

        overflow-hidden

        transition-all
        duration-300

        hover:bg-white/20
        hover:border-white/40

        active:scale-[0.98]

        disabled:
        opacity-60

        disabled:
        cursor-not-allowed

        ${className}
      `}
    >


      {
        loading ? (

          <>

            <Loader2
              className="
                h-4
                w-4
                animate-spin
              "
            />

            <span>
              Connecting...
            </span>

          </>


        ) : (

          <>

            <GoogleIcon />

            <span>
              {text}
            </span>

          </>

        )

      }



    </button>

  );

}