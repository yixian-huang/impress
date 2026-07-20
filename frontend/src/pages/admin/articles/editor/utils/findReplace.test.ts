import { describe, expect, it, vi } from "vitest";
import { collectTextMatches } from "./findReplace";
import type { Editor } from "@tiptap/react";

function fakeEditor(segments: { text: string; pos: number }[]): Editor {
  return {
    state: {
      doc: {
        descendants: (fn: (node: { isText: boolean; text?: string }, pos: number) => void) => {
          for (const s of segments) {
            fn({ isText: true, text: s.text }, s.pos);
          }
        },
      },
    },
  } as unknown as Editor;
}

describe("collectTextMatches", () => {
  it("finds case-insensitive matches by default", () => {
    const editor = fakeEditor([
      { text: "Hello world", pos: 1 },
      { text: "hello again", pos: 20 },
    ]);
    const m = collectTextMatches(editor, "hello");
    expect(m).toEqual([
      { from: 1, to: 6 },
      { from: 20, to: 25 },
    ]);
  });

  it("respects case sensitivity", () => {
    const editor = fakeEditor([{ text: "Hello hello", pos: 1 }]);
    expect(collectTextMatches(editor, "Hello", { caseSensitive: true })).toEqual([
      { from: 1, to: 6 },
    ]);
  });

  it("returns empty for blank query", () => {
    const editor = fakeEditor([{ text: "x", pos: 1 }]);
    expect(collectTextMatches(editor, "")).toEqual([]);
  });
});

// silence unused in case tree-shaking tools complain about vi in some configs
void vi;
