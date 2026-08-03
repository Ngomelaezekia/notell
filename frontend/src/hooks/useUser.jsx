import { useState, useEffect, useCallback } from "react";
import api from "../utils/api";


export const useUser = ({ autoFetch = true } = {}) => {

  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);



  // ===============================
  // Fetch current authenticated user
  // GET /auth/me
  // ===============================

  const fetchProfile = useCallback(async () => {

    setLoading(true);
    setError(null);


    try {

      const response = await api.get("/auth/me");


      const currentUser =
        response.data?.data?.user ??
        response.data?.user ??
        response.data;


      setUser(currentUser);


      return currentUser;


    } catch (err) {

      const message =
        err.response?.data?.message ||
        "Failed to fetch profile";


      setError(message);

      throw err;


    } finally {

      setLoading(false);

    }


  }, []);




  useEffect(() => {

    if(autoFetch){
      fetchProfile();
    }

  },[autoFetch,fetchProfile]);





  // ===============================
  // Update profile
  // PUT /users/profile
  // ===============================

  const updateProfile = async (updatedFields)=>{

    setLoading(true);
    setError(null);


    try {


      const response = await api.put(
        "/users/profile",
        updatedFields
      );


      const updatedUser =
        response.data?.data?.user ??
        response.data?.user ??
        response.data;



      setUser(updatedUser);


      return updatedUser;



    } catch(err){


      const message =
        err.response?.data?.message ||
        "Failed to update profile";


      setError(message);

      throw err;


    } finally {

      setLoading(false);

    }

  };





  // ===============================
  // Profile picture
  // Currently NOT IMPLEMENTED BACKEND
  // ===============================

  const updateProfilePicture = async(file)=>{

    /*
      Backend endpoint does not exist yet.

      Keep this placeholder until we add:
      
      POST /users/profile/avatar

    */


    const localPreview = URL.createObjectURL(file);


    setUser(previous=>({
      ...previous,
      profilePicture:localPreview
    }));


    return {
      message:"avatar preview updated locally"
    };

  };






  // ===============================
  // Change password
  // Currently NOT IMPLEMENTED
  // ===============================

  const changePassword = async(
    currentPassword,
    newPassword
  )=>{


    throw new Error(
      "Password update endpoint is not implemented yet"
    );


  };





  // ===============================
  // Delete account
  // Currently NOT IMPLEMENTED
  // ===============================

  const deleteAccount = async()=>{

    throw new Error(
      "Delete account endpoint is not implemented yet"
    );

  };




  return {

    user,

    loading,

    error,

    refetch:fetchProfile,

    updateProfile,

    updateProfilePicture,

    changePassword,

    deleteAccount,

  };

};