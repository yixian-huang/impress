import { afterEach, describe, expect, it, vi } from "vitest";
import {
  emitMediaUpload,
  formatUploadError,
  subscribeMediaUpload,
} from "./mediaUploadTracked";

describe("mediaUpload bus", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("delivers events to subscribers and supports unsubscribe", () => {
    const seen: string[] = [];
    const unsub = subscribeMediaUpload((e) => seen.push(e.type));
    emitMediaUpload({ type: "start", id: "1", name: "a.png", size: 10 });
    emitMediaUpload({ type: "progress", id: "1", percent: 50 });
    unsub();
    emitMediaUpload({ type: "success", id: "1" });
    expect(seen).toEqual(["start", "progress"]);
  });

  it("isolates subscriber errors", () => {
    const good: string[] = [];
    const unsubBad = subscribeMediaUpload(() => {
      throw new Error("boom");
    });
    const unsubGood = subscribeMediaUpload((e) => good.push(e.type));
    emitMediaUpload({ type: "success", id: "x" });
    expect(good).toEqual(["success"]);
    unsubBad();
    unsubGood();
  });
});

describe("formatUploadError", () => {
  it("prefers API error messages", () => {
    expect(
      formatUploadError({
        response: { data: { error: { message: "quota exceeded" } } },
      }),
    ).toBe("quota exceeded");
  });

  it("falls back to Error.message", () => {
    expect(formatUploadError(new Error("network"))).toBe("network");
  });

  it("uses fallback for empty values", () => {
    expect(formatUploadError(null, "上传失败")).toBe("上传失败");
  });
});
