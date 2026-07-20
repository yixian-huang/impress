import { describe, expect, it } from "vitest";

describe("external theme shared globals", () => {
  it("exposes Inkless shared dependencies with a legacy read alias", async () => {
    await import("./externals");

    const shared = (window as typeof window & {
      __INKLESS_SHARED__?: {
        React?: unknown;
        host?: unknown;
      };
      __IMPRESS_SHARED__?: unknown;
      InklessThemeHost?: unknown;
    });

    expect(shared.__INKLESS_SHARED__).toBeTruthy();
    expect(shared.__INKLESS_SHARED__?.React).toBeTruthy();
    expect(shared.__INKLESS_SHARED__?.host).toBeTruthy();
    expect(shared.InklessThemeHost).toBe(shared.__INKLESS_SHARED__?.host);
    expect(shared.__IMPRESS_SHARED__).toBe(shared.__INKLESS_SHARED__);
    expect((window as typeof window & { React?: unknown }).React).toBeTruthy();
    expect((window as typeof window & { ReactRouterDOM?: unknown }).ReactRouterDOM).toBeTruthy();
  });
});
