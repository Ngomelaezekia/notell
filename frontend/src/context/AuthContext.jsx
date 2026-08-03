import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
} from "react";

import { useNavigate } from "react-router-dom";

import API, { API_BASE_URL } from "../utils/api";


const AuthContext = createContext(null);



export function AuthProvider({ children }) {


  const [user,setUser] = useState(null);

  const [loading,setLoading] = useState(true);

  const [error,setError] = useState(null);


  const navigate = useNavigate();





  const extractErrorMessage = (
    err,
    fallback
  ) => {

    return (
      err.response?.data?.message ||
      err.response?.data?.error ||
      err.message ||
      fallback
    );

  };


  const fetchCurrentUser = useCallback(async()=>{


    try {


      const response =
        await API.get(
          "/auth/me"
        );



      const currentUser =
        response.data?.data?.user ||
        response.data?.user ||
        response.data;



      setUser(currentUser);

      setError(null);



    } catch(err){


      console.warn(
        "No active session",
        err.message
      );


      setUser(null);


    } finally {


      setLoading(false);

    }



  },[]);









  /*
     Initial authentication check

     Handles:
     - normal refresh
     - Google OAuth redirect
  */

  useEffect(()=>{


    const init = async()=>{


      const params =
        new URLSearchParams(
          window.location.search
        );


      const oauthError =
        params.get("error");



      if(oauthError){


        setError(
          `OAuth Error: ${oauthError}`
        );


        window.history.replaceState(
          {},
          document.title,
          window.location.pathname
        );

      }



      await fetchCurrentUser();


    };



    init();



  },[fetchCurrentUser]);









  /*
      Google OAuth

      Backend:
      GET /api/auth/google

      Redirects to Google

      Callback sets cookie

  */

  const loginWithGoogle = ()=>{


    window.location.href =
      `${API_BASE_URL}/auth/google`;


  };









  /*
      Email Login

      Backend sets cookie
  */

  const login = async(credentials)=>{


    setError(null);



    try{


      const response =
        await API.post(
          "/auth/login",
          credentials
        );



      await fetchCurrentUser();



      navigate(
        "/",
        {
          replace:true
        }
      );



      return response.data;



    }catch(err){


      const message =
        extractErrorMessage(
          err,
          "Login failed"
        );


      setError(message);


      throw new Error(message);

    }


  };









  /*
      Register

      Backend sets cookie
  */


  const register = async(formData)=>{


    setError(null);



    try{


      const response =
        await API.post(
          "/auth/register",
          formData
        );



      await fetchCurrentUser();



      navigate(
        "/",
        {
          replace:true
        }
      );



      return response.data;



    }catch(err){


      const message =
        extractErrorMessage(
          err,
          "Registration failed"
        );


      setError(message);


      throw new Error(message);

    }


  };









  const logout = async()=>{


    try{


      await API.post(
        "/auth/logout"
      );


    }catch(err){


      console.warn(
        "Logout warning:",
        err.message
      );


    }finally{


      setUser(null);

      setError(null);


      navigate(
        "/auth",
        {
          replace:true
        }
      );


    }


  };









  const updateUser = (
    updatedFields
  )=>{


    setUser(
      prev =>
        prev
        ? {
            ...prev,
            ...updatedFields
          }
        : null
    );


  };








  return (

    <AuthContext.Provider

      value={{

        user,

        authenticated:
          Boolean(user),

        loading,

        error,

        setError,

        login,

        register,

        loginWithGoogle,

        logout,

        fetchCurrentUser,

        updateUser,

      }}

    >

      {children}

    </AuthContext.Provider>

  );

}






export function useAuth(){

  const context =
    useContext(AuthContext);


  if(!context){

    throw new Error(
      "useAuth must be used inside AuthProvider"
    );

  }


  return context;

}