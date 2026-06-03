import { describe, expect, it } from "vitest";
import { normalizeFeatures } from "./normalizeFeatures";

describe("normalizeFeatures", () => {
  it("returns undefined for empty object (no site_configs row)", () => {
    expect(normalizeFeatures({})).toBeUndefined();
  });

  it("merges publicPages and blog flags", () => {
    const f = normalizeFeatures({
      publicPages: {
        home: true,
        blog: true,
        contact: true,
        about: false,
        experts: false,
        coreServices: false,
        advantages: false,
        cases: false,
      },
      blog: { comments: true, rss: true },
    });
    expect(f?.publicPages.blog).toBe(true);
    expect(f?.blog.rss).toBe(true);
  });

  it("ignores legacy siteMode field", () => {
    const f = normalizeFeatures({
      siteMode: "blog",
      publicPages: { home: true, blog: false, contact: false, about: false, experts: false, coreServices: false, advantages: false, cases: false },
      blog: { comments: false, rss: false },
    });
    expect(f).not.toHaveProperty("siteMode");
    expect(f?.publicPages.blog).toBe(false);
  });
});
