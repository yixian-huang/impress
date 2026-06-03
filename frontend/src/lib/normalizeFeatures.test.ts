import { describe, expect, it } from "vitest";
import { normalizeFeatures, resolveSiteMode } from "./normalizeFeatures";

describe("normalizeFeatures", () => {
  it("returns undefined for empty object (no site_configs row)", () => {
    expect(normalizeFeatures({})).toBeUndefined();
  });

  it("parses explicit siteMode blog", () => {
    const f = normalizeFeatures({
      siteMode: "blog",
      publicPages: { home: true, blog: true, contact: true, about: false, experts: false, coreServices: false, advantages: false, cases: false },
      blog: { comments: true, rss: true },
    });
    expect(f?.siteMode).toBe("blog");
    expect(resolveSiteMode(f)).toBe("blog");
  });

  it("infers blog from BlankSiteSeed-style publicPages when siteMode omitted", () => {
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
    expect(f?.siteMode).toBe("blog");
  });
});
