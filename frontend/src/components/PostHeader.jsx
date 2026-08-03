import React from "react";
import { useNavigate } from "react-router-dom";
import { PlusIcon } from "lucide-react";

export const Headerposts = ({ title = "Feed" }) => {
  const navigate = useNavigate();

  return (
    <header className="sticky top-0 z-20 flex items-center justify-between px-4 py-3 bg-neutral-950/80 backdrop-blur-md border-b border-neutral-800 w-full">
      <button 
        onClick={() => navigate("/create-post")}
        className="p-2 rounded-full hover:bg-neutral-900 transition text-neutral-300 cursor-pointer"
        title="Create Post"
      >
        <PlusIcon size={20} />
      </button>
      </header>
  );
};