import { describe, expect, it } from "vitest";
import { diffLines, htmlToPlainText } from "./textDiff";

describe("diffLines", () => {
  it("marks equal lines", () => {
    const lines = diffLines("a\nb", "a\nb");
    expect(lines.every((l) => l.op === "equal")).toBe(true);
    expect(lines).toHaveLength(2);
  });

  it("detects additions and removals", () => {
    const lines = diffLines("hello\nworld", "hello\nthere\nworld");
    const ops = lines.map((l) => l.op);
    expect(ops).toContain("add");
    expect(ops.filter((o) => o === "equal").length).toBe(2);
  });

  it("handles empty inputs", () => {
    expect(diffLines("", "")).toEqual([{ op: "equal", text: "", leftLine: 1, rightLine: 1 }]);
  });
});

describe("htmlToPlainText", () => {
  it("strips tags and decodes entities", () => {
    expect(htmlToPlainText("<p>Hello&nbsp;<strong>世界</strong></p>")).toContain("Hello");
    expect(htmlToPlainText("<p>Hello&nbsp;<strong>世界</strong></p>")).toContain("世界");
  });
});
