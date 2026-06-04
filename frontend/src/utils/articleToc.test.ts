import { describe, expect, it } from "vitest";
import { resolveTocLayout } from "@/utils/articleToc";

describe("resolveTocLayout", () => {
  it("returns none when fewer than 2 headings", () => {
    expect(resolveTocLayout(0, 10000)).toBe("none");
    expect(resolveTocLayout(1, 10000)).toBe("none");
  });

  it("returns inline for medium articles", () => {
    expect(resolveTocLayout(2, 1000)).toBe("inline");
    expect(resolveTocLayout(3, 5000)).toBe("inline");
  });

  it("returns full for long or heavily structured articles", () => {
    expect(resolveTocLayout(4, 1000)).toBe("full");
    expect(resolveTocLayout(2, 8000)).toBe("full");
  });
});
