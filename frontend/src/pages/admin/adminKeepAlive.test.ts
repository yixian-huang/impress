import { describe, expect, it } from "vitest";
import {
  resolveKeepAliveKey,
  touchKeepAliveLru,
  ADMIN_KEEP_ALIVE_MAX,
} from "./adminKeepAlive";

describe("adminKeepAlive", () => {
  it("keeps list paths and rejects editors", () => {
    expect(resolveKeepAliveKey("/admin")).toBe("/admin");
    expect(resolveKeepAliveKey("/admin/articles")).toBe("/admin/articles");
    expect(resolveKeepAliveKey("/admin/pages")).toBe("/admin/pages");
    expect(resolveKeepAliveKey("/admin/media")).toBe("/admin/media");
    expect(resolveKeepAliveKey("/admin/articles/edit/12")).toBeNull();
    expect(resolveKeepAliveKey("/admin/articles/new")).toBeNull();
    expect(resolveKeepAliveKey("/admin/pages/edit/3")).toBeNull();
    expect(resolveKeepAliveKey("/admin/login")).toBeNull();
  });

  it("normalizes trailing slash", () => {
    expect(resolveKeepAliveKey("/admin/media/")).toBe("/admin/media");
  });

  it("applies LRU eviction", () => {
    let order: string[] = [];
    const keys = Array.from({ length: ADMIN_KEEP_ALIVE_MAX + 2 }, (_, i) => `/admin/k${i}`);
    let evicted: string[] = [];
    for (const key of keys) {
      const result = touchKeepAliveLru(order, key, ADMIN_KEEP_ALIVE_MAX);
      order = result.order;
      evicted = result.evicted;
    }
    expect(order).toHaveLength(ADMIN_KEEP_ALIVE_MAX);
    expect(order[0]).toBe(keys[keys.length - 1]);
    expect(evicted.length).toBeGreaterThan(0);
    expect(order).not.toContain(keys[0]);
  });

  it("moves revisited key to front without growing", () => {
    const first = touchKeepAliveLru(["a", "b", "c"], "b", 8);
    expect(first.order).toEqual(["b", "a", "c"]);
    expect(first.evicted).toEqual([]);
  });
});
