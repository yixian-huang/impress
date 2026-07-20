import { describe, expect, it } from "vitest";
import { ARTICLE_TEMPLATES, getArticleTemplate } from "./articleTemplates";

describe("articleTemplates", () => {
  it("includes tutorial and news", () => {
    expect(ARTICLE_TEMPLATES.map((t) => t.id)).toEqual(
      expect.arrayContaining(["blank", "tutorial", "news", "case-study"]),
    );
  });

  it("tutorial has structured body", () => {
    const t = getArticleTemplate("tutorial");
    expect(t?.zhBody).toContain("步骤");
    expect(t?.enBody).toContain("Steps");
  });
});
