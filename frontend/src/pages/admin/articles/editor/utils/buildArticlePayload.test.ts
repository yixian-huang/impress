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
  it("slugifies empty slug from zhTitle", () => {
    const p = buildArticlePayload(base, { zhBody: "<p>x</p>", enBody: "" }, "draft");
    expect(p.slug).toBeTruthy();
    expect(p.status).toBe("draft");
    expect(p.categoryIds).toEqual([1]);
  });

  it("sets publishedAt when publishing", () => {
    const fixed = "2026-01-01T00:00:00.000Z";
    const p = buildArticlePayload(base, { zhBody: "", enBody: "" }, "published", fixed);
    expect(p.status).toBe("published");
    expect(p.publishedAt).toBe(fixed);
  });
});
