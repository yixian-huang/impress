import { describe, expect, it } from "vitest";
import { parseHtmlOutline, parseMarkdownOutline } from "./outline";

describe("parseMarkdownOutline", () => {
  it("collects ATX headings with 1-based lines", () => {
    const md = "Intro\n\n# One\n\npara\n\n## Two\n\n### Three\n";
    const items = parseMarkdownOutline(md);
    expect(items.map((i) => [i.level, i.text, i.line])).toEqual([
      [1, "One", 3],
      [2, "Two", 7],
      [3, "Three", 9],
    ]);
  });

  it("ignores headings inside fenced code", () => {
    const md = "# Real\n\n```md\n# Fake\n```\n\n## Also real\n";
    const items = parseMarkdownOutline(md);
    expect(items.map((i) => i.text)).toEqual(["Real", "Also real"]);
  });
});

describe("parseHtmlOutline", () => {
  it("strips tags inside headings", () => {
    const html = "<h2>Hello <strong>World</strong></h2><p>x</p><h3>Sub</h3>";
    expect(parseHtmlOutline(html).map((i) => [i.level, i.text])).toEqual([
      [2, "Hello World"],
      [3, "Sub"],
    ]);
  });
});
