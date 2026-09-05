import { useCallback } from "react";
import { useAuth } from "../context/AuthContext";
import api from "../utils/api";

export const useUser = () => {
  const { user, loading, error, fetchCurrentUser, updateUser } = useAuth();

  const updateProfile = useCallback(async (updatedFields) => {
    const response = await api.put("/users/profile", updatedFields);
    const updatedUser = response.data?.data?.user ?? response.data?.user ?? response.data;
    updateUser(updatedUser);
    return updatedUser;
  }, [updateUser]);

  const uploadImage = useCallback(async (file) => {
    if (!file) throw new Error("Image is required");
    const formData = new FormData();
    formData.append("file", file);
    const uploadResponse = await api.post("/upload", formData, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    const url = uploadResponse.data?.url;
    if (!url) throw new Error("Upload did not return a URL");
    return url;
  }, []);

  const updateProfilePicture = useCallback(async (file) => {
    const profilePicture = await uploadImage(file);
    return updateProfile({ profilePicture });
  }, [uploadImage, updateProfile]);

  const updateCoverPicture = useCallback(async (file) => {
    const coverPicture = await uploadImage(file);
    return updateProfile({ coverPicture });
  }, [uploadImage, updateProfile]);

  return {
    user,
    loading,
    error,
    refetch: fetchCurrentUser,
    updateProfile,
    updateProfilePicture,
    updateCoverPicture,
  };
};
