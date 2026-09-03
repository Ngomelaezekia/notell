import { useCallback } from "react";
import { useAuth } from "../context/AuthContext";
import api from "../utils/api";

export const useUser = () => {
  const {
    user,
    loading,
    error,
    fetchCurrentUser,
    updateUser,
  } = useAuth();

  const updateProfile = useCallback(async (updatedFields) => {
    const response = await api.put("/users/profile", updatedFields);
    const updatedUser = response.data?.data?.user ?? response.data?.user ?? response.data;
    updateUser(updatedUser);
    return updatedUser;
  }, [updateUser]);

  const updateProfilePicture = useCallback(async (file) => {
    if (!file) throw new Error("Profile picture is required");

    const formData = new FormData();
    formData.append("file", file);

    const uploadResponse = await api.post("/upload", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });

    const profilePicture = uploadResponse.data?.url;
    if (!profilePicture) throw new Error("Upload did not return a URL");

    return updateProfile({ profilePicture });
  }, [updateProfile]);

  return {
    user,
    loading,
    error,
    refetch: fetchCurrentUser,
    updateProfile,
    updateProfilePicture,
  };
};
