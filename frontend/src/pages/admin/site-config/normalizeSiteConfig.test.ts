import { describe, expect, it } from "vitest";
import { normalizeSiteConfig } from "./normalizeSiteConfig";

describe("normalizeSiteConfig", () => {
  it("fills defaults for empty input", () => {
    const cfg = normalizeSiteConfig(null);
    expect(cfg.identity.name.zh).toBeTruthy();
    expect(cfg.brand.logo.light).toBe("");
    expect(cfg.author.socials).toEqual([]);
  });

  it("preserves tagline and defaultTitle when present", () => {
    const cfg = normalizeSiteConfig({
      identity: {
        name: { zh: "一弦" },
        tagline: { zh: "记录" },
        localeMode: "mono-zh",
        defaultLocale: "zh",
      },
      seo: {
        defaultTitle: { zh: "一弦 · 首页" },
        titleTemplate: "{page} | {site}",
      },
    });
    expect(cfg.identity.tagline?.zh).toBe("记录");
    expect(cfg.seo.defaultTitle?.zh).toBe("一弦 · 首页");
    expect(cfg.seo.titleTemplate).toBe("{page} | {site}");
  });
});
