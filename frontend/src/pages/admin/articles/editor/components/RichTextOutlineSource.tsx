import type { ReactNode } from "react";
import { useEditorState, type Editor } from "@tiptap/react";
import type { OutlineItem } from "../utils/outline";

export type RichTextOutlineStats = {
  headings: OutlineItem[];
  charCount: number;
  wordCount: number;
};

/** TipTap-only reactive outline + counts (lazy chunk). */
function useRichTextOutlineStats(editor: Editor | null): RichTextOutlineStats {
  return useEditorState({
    editor,
    selector: ({ editor: e }) => {
      if (!e) return { headings: [] as OutlineItem[], charCount: 0, wordCount: 0 };
      const h: OutlineItem[] = [];
      e.state.doc.descendants((node, pos) => {
        if (node.type.name === "heading") {
          h.push({ level: node.attrs.level as number, text: node.textContent, pos });
        }
      });
      const text = e.state.doc.textContent;
      const chars = text.length;
      const cjk = (text.match(/[\u4e00-\u9fff\u3400-\u4dbf\uf900-\ufaff]/g) || []).length;
      const latin = text
        .replace(/[\u4e00-\u9fff\u3400-\u4dbf\uf900-\ufaff]/g, " ")
        .trim()
        .split(/\s+/)
        .filter(Boolean).length;
      return { headings: h, charCount: chars, wordCount: cjk + latin };
    },
    equalityFn: (a, b) => {
      if (a.charCount !== b.charCount || a.wordCount !== b.wordCount) return false;
      if (a.headings.length !== b.headings.length) return false;
      return a.headings.every(
        (h, i) =>
          h.level === b.headings[i].level
          && h.text === b.headings[i].text
          && h.pos === b.headings[i].pos,
      );
    },
  });
}

/**
 * Renders children with live TipTap outline stats.
 * Parent passes a render prop so the shell stays TipTap-free.
 */
export function RichTextOutlineSource({
  editor,
  children,
}: {
  editor: Editor | null;
  children: (stats: RichTextOutlineStats) => ReactNode;
}) {
  const stats = useRichTextOutlineStats(editor);
  return <>{children(stats)}</>;
}

export default RichTextOutlineSource;
