import { useEffect, useRef, useState } from "react";

const WATCH_MARKS = [25, 50, 75, 100];

/**
 * TikTok/Reels-style lifecycle for one feed video.
 * A video gets progressively promoted as it approaches the viewport,
 * plays only when it is the dominant visible item, and emits lightweight
 * watch milestones for the recommendation system.
 */
export function useVideoLoadingManager({
  videoRef,
  containerRef,
  enabled = true,
  preloadDistance = "700px 0px",
  onEvent,
}) {
  const [active, setActive] = useState(false);
  const [ready, setReady] = useState(false);
  const [nearViewport, setNearViewport] = useState(false);
  const startedRef = useRef(false);
  const marksRef = useRef(new Set());
  const observerRef = useRef(null);

  useEffect(() => {
    const element = containerRef?.current;
    if (!element || !enabled) return undefined;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry) return;
        const ratio = entry.intersectionRatio;
        const near = entry.isIntersecting && ratio > 0;
        const shouldPlay = ratio >= 0.6;

        setNearViewport(near);
        setActive(shouldPlay);

        const video = videoRef?.current;
        if (!video) return;

        if (near) {
          video.preload = shouldPlay ? "auto" : "metadata";
        }

        if (shouldPlay) {
          video.muted = true;
          if (!startedRef.current) {
            startedRef.current = true;
            onEvent?.("video_start");
          }
          video.play().catch(() => {});
        } else if (!entry.isIntersecting) {
          video.pause();
        }
      },
      { threshold: [0, 0.25, 0.5, 0.6, 0.75], rootMargin: preloadDistance },
    );

    observerRef.current = observer;
    observer.observe(element);

    return () => observer.disconnect();
  }, [containerRef, enabled, onEvent, preloadDistance, videoRef]);

  useEffect(() => {
    const video = videoRef?.current;
    if (!video || !enabled) return undefined;

    const handleLoaded = () => {
      setReady(true);
      onEvent?.("video_ready");
    };

    const handleTimeUpdate = () => {
      if (!Number.isFinite(video.duration) || video.duration <= 0) return;
      const progress = (video.currentTime / video.duration) * 100;
      for (const mark of WATCH_MARKS) {
        if (progress >= mark && !marksRef.current.has(mark)) {
          marksRef.current.add(mark);
          onEvent?.(`watch_${mark}`);
        }
      }
    };

    const handleEnded = () => onEvent?.("video_complete");
    const handleVisibility = () => {
      if (document.visibilityState === "hidden") video.pause();
      else if (active) video.play().catch(() => {});
    };

    video.addEventListener("loadeddata", handleLoaded);
    video.addEventListener("timeupdate", handleTimeUpdate);
    video.addEventListener("ended", handleEnded);
    document.addEventListener("visibilitychange", handleVisibility);

    return () => {
      video.removeEventListener("loadeddata", handleLoaded);
      video.removeEventListener("timeupdate", handleTimeUpdate);
      video.removeEventListener("ended", handleEnded);
      document.removeEventListener("visibilitychange", handleVisibility);
    };
  }, [active, enabled, onEvent, videoRef]);

  return {
    active,
    ready,
    nearViewport,
    markReady: () => setReady(true),
  };
}
