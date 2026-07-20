import { useEffect, useState } from "react";
import type { Editor } from "@tiptap/react";
import { countWords, htmlToPlainText } from "../bilingualUtils";
import { WORD_STATS_DEBOUNCE_MS } from "../utils/constants";

export function useWordStats(opts: {
  editorMode: "richtext" | "markdown";
  markdownContent: Record<string, string>;
  zhBody: string;
  enBody: string;
  zhEditor: Editor | null;
  enEditor: Editor | null;
  /** Recompute when these flip (typing / save). */
  tick: unknown;
}) {
  const { editorMode, markdownContent, zhBody, enBody, zhEditor, enEditor, tick } = opts;
  const [wordStats, setWordStats] = useState({
    zh: { chars: 0, words: 0 },
    en: { chars: 0, words: 0 },
  });

  useEffect(() => {
    const t = window.setTimeout(() => {
      const zhText =
        editorMode === "markdown"
          ? (markdownContent.zh ?? "")
          : (zhEditor?.getText() || htmlToPlainText(zhBody));
      const enText =
        editorMode === "markdown"
          ? (markdownContent.en ?? "")
          : (enEditor?.getText() || htmlToPlainText(enBody));
      setWordStats({
        zh: countWords(zhText),
        en: countWords(enText),
      });
    }, WORD_STATS_DEBOUNCE_MS);
    return () => window.clearTimeout(t);
  }, [editorMode, markdownContent, zhBody, enBody, zhEditor, enEditor, tick]);

  return wordStats;
}
