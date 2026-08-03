import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  User,
  Eye,
  EyeOff,
  Mail,
  Globe,
  MapPin,
  AlertCircle,
  Loader2,
} from "lucide-react";

import { useAuth } from "../context/AuthContext";
import GoogleButton from "../components/GoogleButton";


const initialForm = {
  username: "",
  email: "",
  password: "",
  country: "",
  city: "",
};



export default function AuthPage() {


  const {
    login,
    register,
    loginWithGoogle,
    error: authError
  } = useAuth();



  const [isLogin,setIsLogin] = useState(true);

  const [showPassword,setShowPassword] = useState(false);

  const [loading,setLoading] = useState(false);

  const [localError,setLocalError] = useState("");

  const [formData,setFormData] =
    useState(initialForm);





  const switchMode = ()=>{

    setIsLogin(prev=>!prev);

    setLocalError("");

    setFormData(initialForm);

  };






  const handleChange = (e)=>{

    setFormData(prev=>({
      ...prev,
      [e.target.name]:e.target.value
    }));

  };







  const handleSubmit = async(e)=>{

    e.preventDefault();

    setLoading(true);

    setLocalError("");



    try{


      if(isLogin){


        await login({

          email:formData.email,

          password:formData.password

        });



      }else{


        await register({

          username:formData.username,

          email:formData.email,

          password:formData.password,

          country:formData.country || null,

          city:formData.city || null,

        });


      }



    }catch(err){


      setLocalError(
        err.message ||
        "Authentication failed"
      );


    }finally{


      setLoading(false);

    }


  };







  const error =
    localError ||
    authError;









  return (

    <main
      className="
      relative
      min-h-[100dvh]
      w-full
      flex
      items-center
      justify-center
      overflow-y-auto
      p-4
      "
    >



      <div
        className="
        absolute
        inset-0
        bg-cover
        bg-center
        "
        style={{
          backgroundImage:
          "url('https://images.unsplash.com/photo-1500382017468-9049fed747ef?q=80&w=2000')"
        }}
      />



      <div
        className="
        absolute
        inset-0
        bg-black/40
        "
      />






      <section
        className="
        relative
        z-10
        w-full
        max-w-md
        "
      >



        <div
          className="
          rounded-[32px]
          border
          border-white/30
          bg-white/10
          backdrop-blur-xl
          shadow-2xl
          p-6
          text-white
          "
        >





          <header className="mb-6">


            <h1
              className="
              text-3xl
              font-bold
              "
            >

              {isLogin
              ?"Welcome Back"
              :"Create Account"}

            </h1>



            <p
              className="
              text-sm
              text-white/70
              mt-1
              "
            >

              {isLogin
              ?"Login to continue"
              :"Join the community"}

            </p>


          </header>








          {error && (

            <div
              className="
              mb-4
              flex
              gap-2
              items-center
              rounded-xl
              bg-red-500/20
              border
              border-red-400/30
              p-3
              text-sm
              text-red-200
              "
            >

              <AlertCircle size={16}/>

              {error}

            </div>

          )}








          <AnimatePresence mode="wait">


            <motion.form

              key={isLogin}

              initial={{
                opacity:0,
                y:15
              }}

              animate={{
                opacity:1,
                y:0
              }}

              exit={{
                opacity:0,
                y:-15
              }}

              transition={{
                duration:.2
              }}

              onSubmit={handleSubmit}

              className="space-y-3"

            >








              {!isLogin && (

                <>

                <Input
                  icon={User}
                  name="username"
                  placeholder="Username"
                  value={formData.username}
                  onChange={handleChange}
                  required
                />


                </>

              )}







              <Input
                icon={Mail}
                name="email"
                type="email"
                placeholder="Email address"
                value={formData.email}
                onChange={handleChange}
                required
              />








              {!isLogin && (

                <div
                className="
                grid
                grid-cols-2
                gap-3
                "
                >

                <Input
                  icon={Globe}
                  name="country"
                  placeholder="Country"
                  value={formData.country}
                  onChange={handleChange}
                />


                <Input
                  icon={MapPin}
                  name="city"
                  placeholder="City"
                  value={formData.city}
                  onChange={handleChange}
                />

                </div>

              )}








              <div className="relative">


                <input

                  name="password"

                  type={
                    showPassword
                    ?"text"
                    :"password"
                  }

                  placeholder="Password"

                  value={formData.password}

                  onChange={handleChange}

                  required

                  className="
                  w-full
                  h-11
                  rounded-xl
                  bg-white/10
                  border
                  border-white/30
                  px-4
                  pr-12
                  text-sm
                  placeholder:text-white/60
                  outline-none
                  focus:border-white
                  "

                />



                <button

                  type="button"

                  onClick={()=>
                    setShowPassword(!showPassword)
                  }

                  className="
                  absolute
                  right-4
                  top-1/2
                  -translate-y-1/2
                  text-white/70
                  "

                >

                  {
                    showPassword
                    ?
                    <EyeOff size={17}/>
                    :
                    <Eye size={17}/>
                  }

                </button>


              </div>









              <button

                disabled={loading}

                className="
                mt-2
                w-full
                h-11
                rounded-xl
                bg-gradient-to-r
                from-lime-400
                to-green-600
                text-black
                font-semibold
                flex
                justify-center
                items-center
                gap-2
                disabled:opacity-50
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
                  isLogin
                  ?
                  "Login"
                  :
                  "Create Account"
                }


              </button>









              <div
                className="
                flex
                items-center
                gap-3
                py-3
                "
              >

                <span className="flex-1 h-px bg-white/20"/>

                <span className="
                text-xs
                text-white/60
                ">
                  OR
                </span>

                <span className="flex-1 h-px bg-white/20"/>


              </div>








              <GoogleButton
                onClick={loginWithGoogle}
              />






            </motion.form>


          </AnimatePresence>









          <p
            className="
            mt-6
            text-center
            text-sm
            text-white/80
            "
          >

            {isLogin
            ?
            "Don't have an account?"
            :
            "Already registered?"
            }


            <button

              onClick={switchMode}

              className="
              ml-2
              font-semibold
              underline
              "

            >

              {
                isLogin
                ?
                "Sign Up"
                :
                "Login"
              }

            </button>


          </p>



        </div>


      </section>


    </main>

  );

}








function Input({
  icon:Icon,
  ...props
}){


return (

<div className="relative">


<input

{...props}

className="
w-full
h-11
rounded-xl
bg-white/10
border
border-white/30
px-4
pr-11
text-sm
placeholder:text-white/60
outline-none
focus:border-white
"

/>


<Icon

className="
absolute
right-4
top-1/2
-translate-y-1/2
text-white/60
"

size={16}

/>


</div>

)

}