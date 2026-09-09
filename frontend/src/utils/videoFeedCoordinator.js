const DEFAULT_WINDOW = {
  preload: 2,
  metadata: 10,
};

/**
 * Coordinates which videos should be promoted inside a scrolling feed.
 * The feed can keep many posts mounted while only a small media window
 * consumes bandwidth.
 */
export function createVideoFeedCoordinator(options = {}) {
  const windowSize = { ...DEFAULT_WINDOW, ...options };
  const registry = new Map();

  const register = (id, element) => {
    if (!id || !element) return;
    registry.set(id, element);
  };

  const unregister = (id) => {
    registry.delete(id);
  };

  const update = (activeIndex, items = []) => {
    items.forEach((item, index) => {
      const video = registry.get(item.id);
      if (!video) return;

      const distance = Math.abs(index - activeIndex);

      if (distance === 0) {
        video.preload = "auto";
      } else if (distance <= windowSize.preload) {
        video.preload = "auto";
      } else if (distance <= windowSize.metadata) {
        video.preload = "metadata";
      } else {
        video.pause();
        video.preload = "none";
      }
    });
  };

  return { register, unregister, update };
}
