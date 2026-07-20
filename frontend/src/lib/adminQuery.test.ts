import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  __adminQueryCacheSizeForTests,
  clearAdminQueryCache,
  getAdminQueryData,
  invalidateAdminQueryPrefix,
  setAdminQueryData,
  useAdminQuery,
} from "./adminQuery";

describe("adminQuery cache", () => {
  beforeEach(() => {
    clearAdminQueryCache();
  });

  it("serves cached data without loading on remount", async () => {
    const fetcher = vi.fn().mockResolvedValue({ items: [1], total: 1 });

    const first = renderHook(() =>
      useAdminQuery(["admin", "articles", 1, ""], fetcher, { staleTime: 60_000 }),
    );

    await waitFor(() => expect(first.result.current.loading).toBe(false));
    expect(first.result.current.data).toEqual({ items: [1], total: 1 });
    expect(fetcher).toHaveBeenCalledTimes(1);
    first.unmount();

    const second = renderHook(() =>
      useAdminQuery(["admin", "articles", 1, ""], fetcher, { staleTime: 60_000 }),
    );

    // Cached: no loading flash
    expect(second.result.current.loading).toBe(false);
    expect(second.result.current.data).toEqual({ items: [1], total: 1 });
    // Still fresh — no immediate refetch required within staleTime for display
    expect(second.result.current.data).toBeDefined();
  });

  it("dedupes concurrent fetches for the same key", async () => {
    let resolve!: (value: string) => void;
    const fetcher = vi.fn(
      () =>
        new Promise<string>((r) => {
          resolve = r;
        }),
    );

    const a = renderHook(() => useAdminQuery("shared-key", fetcher));
    const b = renderHook(() => useAdminQuery("shared-key", fetcher));

    expect(fetcher).toHaveBeenCalledTimes(1);
    await act(async () => {
      resolve("ok");
    });
    await waitFor(() => expect(a.result.current.data).toBe("ok"));
    expect(b.result.current.data).toBe("ok");
  });

  it("invalidates by prefix and refetches", async () => {
    const fetcher = vi
      .fn()
      .mockResolvedValueOnce({ v: 1 })
      .mockResolvedValueOnce({ v: 2 });

    const { result } = renderHook(() =>
      useAdminQuery(["admin", "media", 1], fetcher, { staleTime: 60_000 }),
    );
    await waitFor(() => expect(result.current.data).toEqual({ v: 1 }));

    act(() => {
      invalidateAdminQueryPrefix(["admin", "media"]);
    });

    await waitFor(() => expect(result.current.data).toEqual({ v: 2 }));
    expect(fetcher).toHaveBeenCalledTimes(2);
  });

  it("setAdminQueryData updates subscribers", async () => {
    setAdminQueryData(["admin", "x"], { n: 1 });
    expect(getAdminQueryData(["admin", "x"])).toEqual({ n: 1 });

    const fetcher = vi.fn().mockResolvedValue({ n: 9 });
    const { result } = renderHook(() =>
      useAdminQuery(["admin", "x"], fetcher, { staleTime: 60_000 }),
    );
    expect(result.current.data).toEqual({ n: 1 });
    expect(result.current.loading).toBe(false);
  });

  it("clearAdminQueryCache empties store", () => {
    setAdminQueryData("a", 1);
    expect(__adminQueryCacheSizeForTests()).toBeGreaterThan(0);
    clearAdminQueryCache();
    expect(__adminQueryCacheSizeForTests()).toBe(0);
  });
});
