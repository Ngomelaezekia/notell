import React from "react";
import { useNavigate } from "react-router-dom";
import { PlusIcon } from "lucide-react";

export const Headerposts = ({ title = "Feed" }) => {
  const navigate = useNavigate();

  return (
    <header className="sticky top-0 z-20 flex w-full items-center justify-between border-b border-neutral-800 bg-neutral-950/80 px-4 py-3 backdrop-blur-md">
      <h1 className="text-sm font-semibold text-neutral-200">{title}</h1>
      <button
        type="button"
        onClick={() => navigate("/create-post")}
        className="cursor-pointer rounded-full p-2 text-neutral-300 transition hover:bg-neutral-900"
        title="Create Post"
        aria-label="Create Post"
      >
        <PlusIcon size={20} />
      </button>
    </header>
  );
};
