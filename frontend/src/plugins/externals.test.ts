import { describe, expect, it } from "vitest";

describe("external theme shared globals", () => {
  it("exposes Inkless shared dependencies with a legacy read alias", async () => {
    await import("./externals");

    expect((window as typeof window & { __INKLESS_SHARED__?: unknown }).__INKLESS_SHARED__).toBeTruthy();
    expect((window as typeof window & { __IMPRESS_SHARED__?: unknown }).__IMPRESS_SHARED__).toBe(
      (window as typeof window & { __INKLESS_SHARED__?: unknown }).__INKLESS_SHARED__,
    );
  });
});
