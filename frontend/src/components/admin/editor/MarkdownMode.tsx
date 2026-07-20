import { useEffect, useMemo, useRef, useState, useCallback } from "react";
import { EditorView } from "@codemirror/view";
import { EditorState } from "@codemirror/state";
import { markdownToHtml } from "@/lib/markdown";
import MarkdownHtmlPreview from "./MermaidPreview";
import type { MarkdownSelectionApi } from "./MarkdownToolbar";
import {
  createMarkdownBaseExtensions,
  PREVIEW_TYPOGRAPHY_CLASS,
  syncScrollRatio,
} from "./markdownCmSetup";

interface MarkdownModeProps {
  value: string;
  onChange: (value: string) => void;
  onImageUpload?: (file: File) => Promise<string>;
  onApiReady?: (api: MarkdownSelectionApi | null) => void;
  contentKey?: string;
  showPreview?: boolean;
  label?: string;
}

export default function MarkdownMode({
  value,
  onChange,
  onImageUpload,
  onApiReady,
  contentKey,
  showPreview = true,
  label = "Markdown",
}: MarkdownModeProps) {
  const hostRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const previewScrollRef = useRef<HTMLDivElement>(null);
  const syncingRef = useRef<"editor" | "preview" | null>(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;
  const onImageUploadRef = useRef(onImageUpload);
  onImageUploadRef.current = onImageUpload;
  const valueRef = useRef(value);
  valueRef.current = value;
  const applyingExternalRef = useRef(false);

  const [debounced, setDebounced] = useState(value);
  const [cursorLine, setCursorLine] = useState(1);

  useEffect(() => {
    setDebounced(value);
  }, [contentKey]); // eslint-disable-line react-hooks/exhaustive-deps -- locale switch

  useEffect(() => {
    const t = window.setTimeout(() => setDebounced(value), 150);
    return () => window.clearTimeout(t);
  }, [value]);

  const previewHtml = useMemo(() => markdownToHtml(debounced), [debounced]);

  useEffect(() => {
    if (!hostRef.current) return;

    const state = EditorState.create({
      doc: valueRef.current,
      extensions: [
        ...createMarkdownBaseExtensions(),
        EditorView.updateListener.of((update) => {
          if (update.docChanged && !applyingExternalRef.current) {
            const next = update.state.doc.toString();
            valueRef.current = next;
            onChangeRef.current(next);
          }
          if (update.selectionSet || update.docChanged) {
            setCursorLine(update.state.doc.lineAt(update.state.selection.main.head).number);
          }
        }),
        EditorView.domEventHandlers({
          drop: (event, view) => {
            const upload = onImageUploadRef.current;
            if (!upload || !event.dataTransfer) return false;
            const files = Array.from(event.dataTransfer.files).filter((f) =>
              f.type.startsWith("image/"),
            );
            if (files.length === 0) return false;
            event.preventDefault();
            void (async () => {
              let insert = "";
              for (const file of files) {
                const url = await upload(file);
                insert += `\n![${file.name}](${url})\n`;
              }
              const pos = view.state.selection.main.head;
              view.dispatch({
                changes: { from: pos, insert },
                selection: { anchor: pos + insert.length },
              });
            })();
            return true;
          },
          paste: (event, view) => {
            const upload = onImageUploadRef.current;
            if (!upload || !event.clipboardData) return false;
            const images = Array.from(event.clipboardData.items).filter((i) =>
              i.type.startsWith("image/"),
            );
            if (images.length === 0) return false;
            event.preventDefault();
            void (async () => {
              for (const item of images) {
                const file = item.getAsFile();
                if (!file) continue;
                const url = await upload(file);
                const md = `![image](${url})`;
                const pos = view.state.selection.main.head;
                view.dispatch({
                  changes: { from: pos, insert: md },
                  selection: { anchor: pos + md.length },
                });
              }
            })();
            return true;
          },
        }),
      ],
    });

    const view = new EditorView({ state, parent: hostRef.current });
    viewRef.current = view;

    const scroller = view.scrollDOM;
    const onEditorScroll = () => {
      const previewEl = previewScrollRef.current;
      if (!previewEl) return;
      syncScrollRatio(scroller, previewEl, syncingRef, "editor");
    };
    scroller.addEventListener("scroll", onEditorScroll, { passive: true });

    return () => {
      scroller.removeEventListener("scroll", onEditorScroll);
      view.destroy();
      viewRef.current = null;
    };
  }, []);

  useEffect(() => {
    const view = viewRef.current;
    if (!view) return;
    const current = view.state.doc.toString();
    if (current === value) return;
    applyingExternalRef.current = true;
    view.dispatch({ changes: { from: 0, to: current.length, insert: value } });
    applyingExternalRef.current = false;
    valueRef.current = value;
  }, [value, contentKey]);

  useEffect(() => {
    if (!onApiReady) return;
    const api: MarkdownSelectionApi = {
      getValue: () => viewRef.current?.state.doc.toString() ?? valueRef.current,
      setValue: (next, cursor) => {
        const view = viewRef.current;
        if (!view) {
          onChangeRef.current(next);
          return;
        }
        const cur = view.state.doc.toString();
        view.dispatch({
          changes: { from: 0, to: cur.length, insert: next },
          selection: cursor
            ? { anchor: cursor.start, head: cursor.end }
            : { anchor: next.length },
        });
        view.focus();
      },
      getSelection: () => {
        const view = viewRef.current;
        if (!view) return { start: 0, end: 0 };
        const { from, to } = view.state.selection.main;
        return { start: from, end: to };
      },
      focus: () => viewRef.current?.focus(),
    };
    onApiReady(api);
    return () => onApiReady(null);
  }, [onApiReady]);

  useEffect(() => {
    if (!showPreview) return;
    const previewEl = previewScrollRef.current;
    const view = viewRef.current;
    if (!previewEl || !view) return;
    const total = Math.max(1, view.state.doc.lines);
    const ratio = (cursorLine - 1) / total;
    const toMax = previewEl.scrollHeight - previewEl.clientHeight;
    if (toMax <= 0 || syncingRef.current === "editor") return;
    syncingRef.current = "editor";
    previewEl.scrollTop = ratio * toMax;
    requestAnimationFrame(() => {
      syncingRef.current = null;
    });
  }, [cursorLine, showPreview, debounced]);

  const handlePreviewScroll = useCallback(() => {
    const view = viewRef.current;
    const previewEl = previewScrollRef.current;
    if (!view || !previewEl) return;
    syncScrollRatio(previewEl, view.scrollDOM, syncingRef, "preview");
  }, []);

  return (
    <div className="flex h-full min-h-0 gap-0 border border-gray-200 rounded-lg overflow-hidden bg-white">
      <div
        className={`min-w-0 min-h-0 flex flex-col ${showPreview ? "flex-1 border-r border-gray-200" : "flex-1"}`}
      >
        <div className="flex-shrink-0 px-3 py-1.5 text-xs font-medium text-gray-500 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
          <span>{label}</span>
          <span className="text-[10px] text-gray-400 tabular-nums">L{cursorLine}</span>
        </div>
        <div ref={hostRef} className="flex-1 min-h-0 min-w-0 overflow-hidden" />
      </div>

      {showPreview && (
        <div className="flex-1 min-w-0 min-h-0 flex flex-col bg-white">
          <div className="flex-shrink-0 px-3 py-1.5 text-xs font-medium text-gray-500 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
            <span>实时预览</span>
            {contentKey ? (
              <span className="text-[10px] uppercase tracking-wide text-gray-400">{contentKey}</span>
            ) : null}
          </div>
          <div
            ref={previewScrollRef}
            onScroll={handlePreviewScroll}
            className="flex-1 min-h-0 overflow-auto p-4"
          >
            <MarkdownHtmlPreview html={previewHtml} className={PREVIEW_TYPOGRAPHY_CLASS} />
          </div>
        </div>
      )}
    </div>
  );
}
