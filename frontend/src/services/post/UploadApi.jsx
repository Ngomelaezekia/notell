import API from "../../utils/api"


export const uploadAPI = {
  uploadMedia: async(file)=>{
    const formData = new FormData();


    formData.append("file",  file  );
    const response =
      await API.post(
        "/upload",
        formData,
        {
          headers:{
            "Content-Type":"multipart/form-data",
          },
        }
      );
    return response.data;
  }
};
