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

  it("maps legacy impress branding/header/footer into siteConfig fields", () => {
    const cfg = normalizeSiteConfig({
      branding: {
        companyName: { zh: "印迹安合法规咨询", en: "Blotting Consultancy" },
        logo: { url: "/uploads/logo.png", alt: { zh: "", en: "" } },
      },
      header: { logo: { url: "/images/logo.svg" } },
      footer: {
        copyright: { zh: "版权所有 © 2018-2025 印迹法规 京ICP备12345678号" },
        phone: { zh: "+86 159 1076 9614" },
      },
    });
    expect(cfg.identity.name.zh).toBe("印迹安合法规咨询");
    expect(cfg.identity.name.en).toBe("Blotting Consultancy");
    expect(cfg.brand.logo.light).toBe("/uploads/logo.png");
    expect(cfg.header?.brandMode).toBe("logo");
    expect(cfg.footer.copyright?.zh).toContain("版权所有");
    expect(cfg.footer.icp).toBe("京ICP备12345678号");
    expect(cfg.identity.name.zh).not.toMatch(/My Site/i);
  });
});
