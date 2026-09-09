const DEFAULT_WINDOW = 10;

/**
 * Keeps a lightweight queue around feed videos.
 * The queue does not force downloads; it coordinates which media items
 * are candidates for metadata preparation and preload.
 */
export class VideoPreloadQueue {
  constructor(windowSize = DEFAULT_WINDOW) {
    this.windowSize = windowSize;
    this.items = new Map();
  }

  register(id, element) {
    if (!id || !element) return;
    this.items.set(id, { element, state: "registered" });
  }

  prepare(ids = []) {
    const active = new Set(ids.slice(0, this.windowSize));

    for (const [id, item] of this.items.entries()) {
      if (!active.has(id)) {
        item.element.preload = "metadata";
        item.state = "idle";
        continue;
      }

      item.element.preload = "auto";
      item.state = "preparing";
    }
  }

  release(id) {
    const item = this.items.get(id);
    if (!item) return;
    item.element.pause();
    item.element.preload = "metadata";
    item.state = "released";
  }

  clear() {
    this.items.clear();
  }
}

export const videoPreloadQueue = new VideoPreloadQueue();
