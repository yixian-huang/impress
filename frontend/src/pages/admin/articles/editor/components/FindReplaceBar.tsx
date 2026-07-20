import { useCallback, useEffect, useMemo, useState } from "react";
import type { Editor } from "@tiptap/react";
import {
  collectTextMatches,
  replaceAllMatches,
  replaceMatch,
  selectMatch,
  type TextMatch,
} from "../utils/findReplace";

/**
 * Floating find/replace bar for TipTap.
 * Markdown mode relies on CodeMirror's built-in search (Mod-f).
 */
export function FindReplaceBar({
  open,
  editor,
  onClose,
}: {
  open: boolean;
  editor: Editor | null;
  onClose: () => void;
}) {
  const [query, setQuery] = useState("");
  const [replacement, setReplacement] = useState("");
  const [caseSensitive, setCaseSensitive] = useState(false);
  const [index, setIndex] = useState(0);
  const [tick, setTick] = useState(0);

  const matches: TextMatch[] = useMemo(() => {
    if (!open || !editor || !query) return [];
    void tick;
    return collectTextMatches(editor, query, { caseSensitive });
  }, [open, editor, query, caseSensitive, tick]);

  const goTo = useCallback(
    (i: number) => {
      if (!editor || matches.length === 0) return;
      const next = ((i % matches.length) + matches.length) % matches.length;
      setIndex(next);
      selectMatch(editor, matches[next]);
    },
    [editor, matches],
  );

  useEffect(() => {
    if (!open) return;
    if (matches.length > 0) {
      const i = Math.min(index, matches.length - 1);
      setIndex(i);
      if (editor) selectMatch(editor, matches[i]);
    }
    // only when query/matches set changes
    // eslint-disable-next-line react-hooks/exhaustive-deps -- intentional: re-select on new search
  }, [open, query, caseSensitive, matches.length]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        onClose();
        return;
      }
      if ((e.metaKey || e.ctrlKey) && e.key === "g") {
        e.preventDefault();
        if (e.shiftKey) goTo(index - 1);
        else goTo(index + 1);
      }
      if (e.key === "Enter" && (e.target as HTMLElement)?.tagName === "INPUT") {
        e.preventDefault();
        if (e.shiftKey) goTo(index - 1);
        else goTo(index + 1);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose, goTo, index]);

  if (!open) return null;

  const status =
    !query
      ? "输入查找内容"
      : matches.length === 0
        ? "无匹配"
        : `${index + 1} / ${matches.length}`;

  return (
    <div className="flex flex-wrap items-center gap-2 px-3 py-2 border-b border-slate-200 bg-white shadow-sm">
      <input
        autoFocus
        type="search"
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          setIndex(0);
        }}
        placeholder="查找"
        className="w-40 px-2 py-1 text-sm border border-slate-200 rounded-md focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
      />
      <input
        type="text"
        value={replacement}
        onChange={(e) => setReplacement(e.target.value)}
        placeholder="替换为"
        className="w-40 px-2 py-1 text-sm border border-slate-200 rounded-md focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
      />
      <label className="flex items-center gap-1 text-xs text-slate-600 select-none">
        <input
          type="checkbox"
          checked={caseSensitive}
          onChange={(e) => setCaseSensitive(e.target.checked)}
        />
        区分大小写
      </label>
      <span className="text-xs text-slate-500 tabular-nums min-w-[4.5rem]">{status}</span>
      <button
        type="button"
        disabled={matches.length === 0}
        onClick={() => goTo(index - 1)}
        className="px-2 py-1 text-xs border border-slate-200 rounded-md disabled:opacity-40 hover:bg-slate-50"
      >
        上一个
      </button>
      <button
        type="button"
        disabled={matches.length === 0}
        onClick={() => goTo(index + 1)}
        className="px-2 py-1 text-xs border border-slate-200 rounded-md disabled:opacity-40 hover:bg-slate-50"
      >
        下一个
      </button>
      <button
        type="button"
        disabled={!editor || matches.length === 0}
        onClick={() => {
          if (!editor || matches.length === 0) return;
          const m = matches[index];
          replaceMatch(editor, m, replacement);
          setTick((t) => t + 1);
        }}
        className="px-2 py-1 text-xs border border-slate-200 rounded-md disabled:opacity-40 hover:bg-slate-50"
      >
        替换
      </button>
      <button
        type="button"
        disabled={!editor || !query || matches.length === 0}
        onClick={() => {
          if (!editor || !query) return;
          replaceAllMatches(editor, query, replacement, { caseSensitive });
          setTick((t) => t + 1);
          setIndex(0);
        }}
        className="px-2 py-1 text-xs border border-slate-200 rounded-md disabled:opacity-40 hover:bg-slate-50"
      >
        全部替换
      </button>
      <button
        type="button"
        onClick={onClose}
        className="ml-auto px-2 py-1 text-xs text-slate-500 hover:text-slate-800"
        title="关闭 (Esc)"
      >
        关闭
      </button>
    </div>
  );
}
