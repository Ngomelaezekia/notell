import { useEffect, useRef, useState } from "react";

/**
 * Coordinates short-form video loading behavior.
 * Keeps only the visible video active while allowing nearby videos
 * to prepare without exhausting bandwidth.
 */
export function useVideoLoadingManager({
  videoRef,
  containerRef,
  enabled = true,
  preloadDistance = "360px 0px",
}) {
  const [active, setActive] = useState(false);
  const [ready, setReady] = useState(false);
  const observerRef = useRef(null);

  useEffect(() => {
    const element = containerRef?.current;
    if (!element || !enabled) return undefined;

    observerRef.current = new IntersectionObserver(
      ([entry]) => {
        const visible = entry.intersectionRatio;

        if (visible >= 0.25) {
          setActive(true);
        }

        const video = videoRef?.current;
        if (!video) return;

        if (visible >= 0.6) {
          video.muted = true;
          video.preload = "auto";
          video.play().catch(() => {});
        } else if (visible === 0) {
          video.pause();
        }
      },
      {
        threshold: [0, 0.25, 0.6],
        rootMargin: preloadDistance,
      },
    );

    observerRef.current.observe(element);

    return () => {
      observerRef.current?.disconnect();
    };
  }, [containerRef, enabled, preloadDistance, videoRef]);

  return {
    active,
    ready,
    markReady: () => setReady(true),
  };
}
