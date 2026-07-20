import type { Editor } from "@tiptap/react";

export type TextMatch = { from: number; to: number };

/**
 * Collect all plain-text matches in a TipTap/ProseMirror document.
 * Matches do not cross text-node boundaries (standard for editor find).
 */
export function collectTextMatches(
  editor: Editor,
  query: string,
  opts?: { caseSensitive?: boolean },
): TextMatch[] {
  if (!query) return [];
  const caseSensitive = !!opts?.caseSensitive;
  const q = caseSensitive ? query : query.toLowerCase();
  const matches: TextMatch[] = [];

  editor.state.doc.descendants((node, pos) => {
    if (!node.isText || !node.text) return;
    const hay = caseSensitive ? node.text : node.text.toLowerCase();
    let from = 0;
    while (from < hay.length) {
      const idx = hay.indexOf(q, from);
      if (idx < 0) break;
      matches.push({ from: pos + idx, to: pos + idx + query.length });
      from = idx + Math.max(1, query.length);
    }
  });

  return matches;
}

export function selectMatch(editor: Editor, match: TextMatch): void {
  editor.chain().focus().setTextSelection({ from: match.from, to: match.to }).run();
  try {
    const start = editor.view.coordsAtPos(match.from);
    const editorDom = editor.view.dom;
    const scroller = editorDom.closest(".overflow-y-auto") || editorDom.parentElement;
    if (scroller && start) {
      const rect = (scroller as HTMLElement).getBoundingClientRect();
      if (start.top < rect.top + 40 || start.top > rect.bottom - 40) {
        const dom = editor.view.domAtPos(match.from);
        const el = dom.node instanceof HTMLElement ? dom.node : dom.node.parentElement;
        el?.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }
  } catch {
    /* ignore scroll errors */
  }
}

export function replaceMatch(editor: Editor, match: TextMatch, replacement: string): TextMatch {
  editor
    .chain()
    .focus()
    .setTextSelection({ from: match.from, to: match.to })
    .insertContent(replacement)
    .run();
  return { from: match.from, to: match.from + replacement.length };
}

/** Replace all; returns number of replacements. */
export function replaceAllMatches(
  editor: Editor,
  query: string,
  replacement: string,
  opts?: { caseSensitive?: boolean },
): number {
  const matches = collectTextMatches(editor, query, opts);
  if (matches.length === 0) return 0;
  // Apply from end so earlier positions stay valid
  const sorted = [...matches].sort((a, b) => b.from - a.from);
  const { tr } = editor.state;
  let chain = tr;
  for (const m of sorted) {
    chain = chain.insertText(replacement, m.from, m.to);
  }
  editor.view.dispatch(chain);
  editor.commands.focus();
  return matches.length;
}
