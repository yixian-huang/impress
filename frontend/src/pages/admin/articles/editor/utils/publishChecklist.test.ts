import { describe, expect, it } from "vitest";
import {
  checklistHasBlocks,
  evaluatePublishChecklist,
  isBodyEmpty,
} from "./publishChecklist";

const base = {
  zhTitle: "标题",
  enTitle: "",
  slug: "hello",
  coverImage: "https://x/y.jpg",
  zhBody: "<p>正文内容</p>",
  enBody: "",
  zhMetaDescription:
    "这是一段足够长的中文描述用来通过最短长度检查的文字，需要再写一些内容确保超过五十个字符的要求。继续补充几个字即可。",
  enMetaDescription: "",
  zhSeoTitle: "SEO",
  enSeoTitle: "",
  enabledLangs: ["zh"],
  author: "Ada",
};

describe("isBodyEmpty", () => {
  it("treats empty shells as empty", () => {
    expect(isBodyEmpty("<p></p>")).toBe(true);
    expect(isBodyEmpty("<p><br></p>")).toBe(true);
    expect(isBodyEmpty("<p>hi</p>")).toBe(false);
  });
});

describe("evaluatePublishChecklist", () => {
  it("returns no issues for a complete article", () => {
    expect(evaluatePublishChecklist(base)).toEqual([]);
  });

  it("blocks empty title and body", () => {
    const items = evaluatePublishChecklist({
      ...base,
      zhTitle: "",
      zhBody: "<p></p>",
    });
    expect(checklistHasBlocks(items)).toBe(true);
    expect(items.map((i) => i.id)).toEqual(expect.arrayContaining(["zh-title", "zh-body"]));
  });

  it("warns on missing cover and meta", () => {
    const items = evaluatePublishChecklist({
      ...base,
      coverImage: "",
      zhMetaDescription: "",
      author: "",
    });
    expect(checklistHasBlocks(items)).toBe(false);
    expect(items.map((i) => i.id)).toEqual(
      expect.arrayContaining(["cover", "zh-meta", "author"]),
    );
  });

  it("warns when English is enabled but empty", () => {
    const items = evaluatePublishChecklist({
      ...base,
      enabledLangs: ["zh", "en"],
    });
    expect(items.some((i) => i.id === "en-empty")).toBe(true);
  });
});
