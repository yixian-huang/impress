import { afterEach, describe, expect, it, vi } from "vitest";
import { emitMediaUpload, subscribeMediaUpload } from "./mediaUploadTracked";

describe("mediaUpload bus", () => {
  afterEach(() => {
    // no global reset needed — each test unsubscribes
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
});

// silence unused import if tree-shaken in some runners
void vi;
