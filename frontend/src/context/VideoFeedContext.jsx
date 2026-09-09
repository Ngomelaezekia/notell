import { createContext, useCallback, useContext, useMemo, useRef, useState } from "react";

const VideoFeedContext = createContext(null);

export function VideoFeedProvider({ children }) {
  const videos = useRef(new Map());
  const [activeId, setActiveId] = useState(null);

  const registerVideo = useCallback((id, element) => {
    if (!id || !element) return;
    videos.current.set(id, element);
    return () => videos.current.delete(id);
  }, []);

  const activateVideo = useCallback((id) => {
    setActiveId(id);

    videos.current.forEach((video, videoId) => {
      if (videoId === id) {
        video.muted = true;
        video.preload = "auto";
        video.play().catch(() => {});
      } else {
        video.pause();
      }
    });
  }, []);

  const prepareVideo = useCallback((id) => {
    const video = videos.current.get(id);
    if (video) video.preload = "auto";
  }, []);

  const value = useMemo(() => ({
    activeId,
    registerVideo,
    activateVideo,
    prepareVideo,
  }), [activeId, registerVideo, activateVideo, prepareVideo]);

  return <VideoFeedContext.Provider value={value}>{children}</VideoFeedContext.Provider>;
}

export function useVideoFeed() {
  const context = useContext(VideoFeedContext);
  if (!context) throw new Error("useVideoFeed must be used inside VideoFeedProvider");
  return context;
}
