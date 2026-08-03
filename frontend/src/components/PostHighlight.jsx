import React from "react";

export const HighlighPost = () => {
  const dummyHighlights = [1, 2, 3, 4, 5, 6, 7];

  return (
    <div className="border-b border-neutral-800/80 py-4 px-2 w-full">
      <div className="flex gap-4 overflow-x-auto no-scrollbar px-2">
        {dummyHighlights.map((item) => (
          <div
            key={item}
            className="flex flex-col items-center gap-1.5 flex-shrink-0 cursor-pointer group"
          >
            <div className="p-[2px] rounded-full bg-gradient-to-tr from-amber-500 via-rose-500 to-indigo-500 group-hover:scale-105 transition-transform">
              <div className="w-14 h-14 rounded-full border-2 border-neutral-950 bg-neutral-800 overflow-hidden flex items-center justify-center">
                <span className="text-xs text-neutral-400">Story</span>
              </div>
            </div>
            <span className="text-[11px] text-neutral-400 truncate max-w-[64px]">
              user_{item}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};






