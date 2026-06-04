import { describe, expect, it } from "vitest";
import { articleReadingStatsFromHtml, countWordsFromText } from "./articleReadingStats";

describe("articleReadingStats", () => {
  it("counts CJK and latin words", () => {
    expect(countWordsFromText("hello 世界")).toBe(3);
  });

  it("strips HTML before counting", () => {
    const stats = articleReadingStatsFromHtml("<p>你好</p><p>world test</p>");
    expect(stats.wordCount).toBeGreaterThanOrEqual(3);
    expect(stats.readingMinutes).toBeGreaterThanOrEqual(1);
  });
});
