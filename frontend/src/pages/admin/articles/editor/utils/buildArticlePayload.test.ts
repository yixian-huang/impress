import { describe, expect, it } from "vitest";
import { buildArticlePayload, type ArticlePayloadFields } from "./buildArticlePayload";

const base: ArticlePayloadFields = {
  zhTitle: "你好",
  enTitle: "Hello",
  slug: "",
  coverImage: "",
  zhSeoTitle: "",
  enSeoTitle: "",
  zhMetaDescription: "",
  enMetaDescription: "",
  ogImage: "",
  selectedCategoryIds: [1],
  selectedTagIds: [2],
  author: "a",
  autoSummary: false,
  allowComments: true,
  pinned: false,
  visibility: "public",
  metadata: {},
};

describe("buildArticlePayload", () => {
  it("slugifies empty slug preferring English title as Latin path", () => {
    const p = buildArticlePayload(base, { zhBody: "<p>x</p>", enBody: "" }, "draft");
    expect(p.slug).toBe("hello");
    expect(p.status).toBe("draft");
    expect(p.categoryIds).toEqual([1]);
  });

  it("slugifies Chinese-only title to pinyin letters", () => {
    const p = buildArticlePayload(
      { ...base, enTitle: "", zhTitle: "你好世界" },
      { zhBody: "<p>x</p>", enBody: "" },
      "draft",
    );
    expect(p.slug).toBe("ni-hao-shi-jie");
  });

  it("sets publishedAt when publishing", () => {
    const fixed = "2026-01-01T00:00:00.000Z";
    const p = buildArticlePayload(base, { zhBody: "", enBody: "" }, "published", fixed);
    expect(p.status).toBe("published");
    expect(p.publishedAt).toBe(fixed);
  });
});
