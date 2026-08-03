import { useState, useEffect, useCallback } from "react";
import { User, Mail } from "lucide-react";

import Header from "./Header";
import Stepper from "./Stepper";
import Navigation from "./Navigation";
import StatusToast from "./StatusToast";

import FormSection from "./common/FormSection";
import FloatingInput from "./common/FloatingInput";
import FloatingTextarea from "./common/FloatingTextarea";
import UploadZone from "./common/UploadZone";

import { useUser } from "../../hooks/useUser";


const steps = [
  {
    id: 1,
    title: "Profile",
    icon: User,
  },
  {
    id: 2,
    title: "Contact & Location",
    icon: Mail,
  },
];


const initialForm = {
  username: "",
  email: "",
  phone: "",
  bio: "",
  city: "",
  country: "",
};



export default function UserManage() {

  const {
    user,
    loading,
    updateProfile,
    updateProfilePicture,
  } = useUser();



  const [currentStep, setCurrentStep] = useState(1);

  const [form, setForm] = useState(initialForm);

  const [avatar, setAvatar] = useState(null);

  const [toast, setToast] = useState(null);

  const [saving, setSaving] = useState(false);



  /*
    Load existing user data
  */
  useEffect(() => {

    if (!user) return;


    setForm((previous) => ({
      ...previous,

      username: user.username ?? "",

      email: user.email ?? "",

      bio: user.bio ?? "",

      city: user.city ?? "",

      country: user.country ?? "",
    }));


    setAvatar(
      user.profilePicture ??
      user.profile_picture ??
      null
    );


  }, [user]);





  const updateField = useCallback(
    (field, value) => {

      setForm((previous)=>({
        ...previous,
        [field]:value,
      }));

    },
    []
  );





  const handleSave = async()=>{

    if(saving) return;


    setSaving(true);

    setToast(null);



    try {


      /*
        Upload avatar if changed
      */
      if(avatar instanceof File){

        await updateProfilePicture(avatar);

      }




      /*
        Update profile data
      */
      await updateProfile({

        username:form.username,

        email:form.email,

        bio:form.bio,

        city:form.city,

        country:form.country,

      });




      setToast({

        type:"success",

        message:"Profile updated successfully",

      });



    }catch(error){


      setToast({

        type:"error",

        message:
          error.response?.data?.message ??
          error.message ??
          "Unable to update profile",

      });



    }finally{

      setSaving(false);

    }

  };





  if(loading){

    return (

      <div
        className="
          min-h-screen
          flex
          items-center
          justify-center
          bg-linear-to-br
          from-slate-100
          via-white
          to-indigo-100
        "
      >

        <div
          className="
            text-slate-600
            font-medium
            animate-pulse
          "
        >
          Loading profile settings...
        </div>

      </div>

    );

  }





  const renderStep = ()=>{


    switch(currentStep){


      case 1:

        return (

          <FormSection

            icon={User}

            title="Personal Information"

            description="Update your profile details and avatar."

          >


            <FloatingInput

              label="Username"

              value={form.username}

              onChange={(e)=>
                updateField(
                  "username",
                  e.target.value
                )
              }

            />



            <FloatingTextarea

              label="Bio"

              value={form.bio}

              onChange={(e)=>
                updateField(
                  "bio",
                  e.target.value
                )
              }

            />



            <UploadZone

              avatar

              value={avatar}

              onChange={setAvatar}

              label="Profile Picture"

            />


          </FormSection>

        );





      case 2:

        return (

          <FormSection

            icon={Mail}

            title="Contact & Location"

            description="Manage your contact information."

          >



            <FloatingInput

              label="Email Address"

              value={form.email}

              onChange={(e)=>
                updateField(
                  "email",
                  e.target.value
                )
              }

            />



            <FloatingInput

              label="City"

              value={form.city}

              onChange={(e)=>
                updateField(
                  "city",
                  e.target.value
                )
              }

            />



            <FloatingInput

              label="Country"

              value={form.country}

              onChange={(e)=>
                updateField(
                  "country",
                  e.target.value
                )
              }

            />


          </FormSection>

        );



      default:

        return null;

    }

  };





  return (

    <div
      className="
        min-h-screen
        bg-linear-to-br
        from-slate-100
        via-white
        to-indigo-100
        p-6
      "
    >


      <Header />



      <main
        className="
          max-w-5xl
          mx-auto
          mt-8
        "
      >


        <Stepper

          steps={steps}

          currentStep={currentStep}

          setCurrentStep={setCurrentStep}

        />



        <section className="mt-8">

          {renderStep()}

        </section>



        <Navigation

          currentStep={currentStep}

          setCurrentStep={setCurrentStep}

          totalSteps={steps.length}

          onSave={handleSave}

          saving={saving}

        />


      </main>




      {toast && (

        <StatusToast

          {...toast}

          onClose={() => setToast(null)}

        />

      )}



    </div>

  );

}