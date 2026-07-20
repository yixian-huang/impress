import { describe, expect, it } from "vitest";
import { slugifyTitle, titleToLatinSlug } from "./slugify";

describe("titleToLatinSlug", () => {
  it("slugifies English titles", () => {
    expect(titleToLatinSlug("Hello World!")).toBe("hello-world");
    expect(titleToLatinSlug("  Inkless CMS  ")).toBe("inkless-cms");
  });

  it("romanizes Chinese titles to pinyin letters", () => {
    const slug = titleToLatinSlug("产品设计与增长");
    expect(slug).toMatch(/^[a-z0-9-]+$/);
    expect(slug).not.toMatch(/[\u4e00-\u9fff]/);
    expect(slug).toContain("chan");
    expect(slug).toContain("pin");
  });

  it("handles mixed Chinese + English", () => {
    const slug = titleToLatinSlug("Inkless 介绍");
    expect(slug).toMatch(/^inkless-/);
    expect(slug).not.toMatch(/[\u4e00-\u9fff]/);
  });
});

describe("slugifyTitle", () => {
  it("prefers English title when present", () => {
    expect(slugifyTitle("中文标题", "My English Title")).toBe("my-english-title");
  });

  it("falls back to pinyin of Chinese title", () => {
    const slug = slugifyTitle("你好世界");
    expect(slug).toBe("ni-hao-shi-jie");
  });

  it("falls back to article-* when empty", () => {
    expect(slugifyTitle("")).toMatch(/^article-\d+$/);
    expect(slugifyTitle("   ", "   ")).toMatch(/^article-\d+$/);
  });
});
