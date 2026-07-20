import { useEffect, useMemo, useRef, useState, useCallback } from "react";
import { markdownToHtml } from "@/lib/markdown";
import MarkdownHtmlPreview from "./MermaidPreview";
import type { MarkdownSelectionApi } from "./MarkdownToolbar";

interface MarkdownModeProps {
  value: string;
  onChange: (value: string) => void;
  onImageUpload?: (file: File) => Promise<string>;
  /** Expose selection/wrap API for the external Markdown toolbar. */
  onApiReady?: (api: MarkdownSelectionApi | null) => void;
  /** Optional key identity (e.g. active locale) — forces debounce reset when it changes. */
  contentKey?: string;
}

export default function MarkdownMode({
  value,
  onChange,
  onImageUpload,
  onApiReady,
  contentKey,
}: MarkdownModeProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const lineGutterRef = useRef<HTMLDivElement>(null);
  const editorScrollRef = useRef<HTMLDivElement>(null);
  const previewScrollRef = useRef<HTMLDivElement>(null);
  const syncingRef = useRef<"editor" | "preview" | null>(null);
  const [debounced, setDebounced] = useState(value);

  // Live preview: short debounce while typing; immediate when contentKey/locale switches
  useEffect(() => {
    setDebounced(value);
  }, [contentKey]); // eslint-disable-line react-hooks/exhaustive-deps -- intentional: locale switch

  useEffect(() => {
    const t = window.setTimeout(() => setDebounced(value), 150);
    return () => window.clearTimeout(t);
  }, [value]);

  const previewHtml = useMemo(() => markdownToHtml(debounced), [debounced]);

  const lineCount = useMemo(() => {
    // Always at least 1 line for empty content
    if (!value) return 1;
    let n = 1;
    for (let i = 0; i < value.length; i++) {
      if (value.charCodeAt(i) === 10) n++;
    }
    return n;
  }, [value]);

  const lineNumbers = useMemo(() => {
    const lines: number[] = new Array(lineCount);
    for (let i = 0; i < lineCount; i++) lines[i] = i + 1;
    return lines;
  }, [lineCount]);

  // Keep latest value/onChange for the selection API without re-creating every keystroke
  const valueRef = useRef(value);
  valueRef.current = value;
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  useEffect(() => {
    if (!onApiReady) return;
    const api: MarkdownSelectionApi = {
      getValue: () => valueRef.current,
      setValue: (next, cursor) => {
        onChangeRef.current(next);
        // Restore selection after React re-render
        requestAnimationFrame(() => {
          const el = textareaRef.current;
          if (!el || !cursor) return;
          el.focus();
          el.setSelectionRange(cursor.start, cursor.end);
        });
      },
      getSelection: () => {
        const el = textareaRef.current;
        if (!el) return { start: 0, end: 0 };
        return { start: el.selectionStart, end: el.selectionEnd };
      },
      focus: () => textareaRef.current?.focus(),
    };
    onApiReady(api);
    return () => onApiReady(null);
  }, [onApiReady]);

  const syncScroll = useCallback((source: "editor" | "preview") => {
    if (syncingRef.current && syncingRef.current !== source) return;
    const editorEl = editorScrollRef.current;
    const previewEl = previewScrollRef.current;
    if (!editorEl || !previewEl) return;

    const from = source === "editor" ? editorEl : previewEl;
    const to = source === "editor" ? previewEl : editorEl;
    const fromMax = from.scrollHeight - from.clientHeight;
    const toMax = to.scrollHeight - to.clientHeight;
    if (fromMax <= 0 || toMax <= 0) {
      // Still sync line gutter with editor scroll
      if (source === "editor" && lineGutterRef.current) {
        lineGutterRef.current.scrollTop = editorEl.scrollTop;
      }
      return;
    }

    const ratio = from.scrollTop / fromMax;
    syncingRef.current = source;
    to.scrollTop = ratio * toMax;
    if (source === "editor" && lineGutterRef.current) {
      lineGutterRef.current.scrollTop = editorEl.scrollTop;
    } else if (source === "preview" && lineGutterRef.current && editorEl) {
      lineGutterRef.current.scrollTop = editorEl.scrollTop;
    }
    // Release lock on next frame so mutual scroll events don't loop
    requestAnimationFrame(() => {
      syncingRef.current = null;
    });
  }, []);

  const handleEditorScroll = useCallback(() => {
    if (lineGutterRef.current && editorScrollRef.current) {
      lineGutterRef.current.scrollTop = editorScrollRef.current.scrollTop;
    }
    syncScroll("editor");
  }, [syncScroll]);

  const handlePreviewScroll = useCallback(() => {
    syncScroll("preview");
  }, [syncScroll]);

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    if (!onImageUpload) return;
    const files = Array.from(e.dataTransfer.files).filter((f) => f.type.startsWith("image/"));
    for (const file of files) {
      const url = await onImageUpload(file);
      const md = `![${file.name}](${url})`;
      onChange(value + "\n" + md);
    }
  };

  const handlePaste = async (e: React.ClipboardEvent) => {
    if (!onImageUpload) return;
    const items = Array.from(e.clipboardData.items);
    for (const item of items) {
      if (item.type.startsWith("image/")) {
        e.preventDefault();
        const file = item.getAsFile();
        if (file) {
          const url = await onImageUpload(file);
          const md = `![image](${url})`;
          onChange(value + "\n" + md);
        }
      }
    }
  };

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
      const el = e.currentTarget;
      const { selectionStart, selectionEnd } = el;
      const selected = value.substring(selectionStart, selectionEnd);

      if (e.ctrlKey || e.metaKey) {
        if (e.key === "b") {
          e.preventDefault();
          onChange(
            value.substring(0, selectionStart) +
              `**${selected || "粗体"}**` +
              value.substring(selectionEnd),
          );
        } else if (e.key === "i") {
          e.preventDefault();
          onChange(
            value.substring(0, selectionStart) +
              `*${selected || "斜体"}*` +
              value.substring(selectionEnd),
          );
        } else if (e.key === "k") {
          e.preventDefault();
          onChange(
            value.substring(0, selectionStart) +
              `[${selected || "链接文字"}](url)` +
              value.substring(selectionEnd),
          );
        }
      }

      if (e.key === "Tab") {
        e.preventDefault();
        const start = selectionStart;
        const end = selectionEnd;
        const next = value.substring(0, start) + "  " + value.substring(end);
        onChange(next);
        requestAnimationFrame(() => {
          el.selectionStart = el.selectionEnd = start + 2;
        });
      }
    },
    [onChange, value],
  );

  return (
    <div className="flex h-full min-h-0 gap-0 border border-gray-200 rounded-lg overflow-hidden bg-white">
      {/* Editor */}
      <div className="flex-1 min-w-0 min-h-0 flex flex-col border-r border-gray-200">
        <div className="flex-shrink-0 px-3 py-1.5 text-xs font-medium text-gray-500 bg-gray-50 border-b border-gray-200">
          Markdown
        </div>
        <div className="flex-1 min-h-0 flex">
          {/* Line numbers */}
          <div
            ref={lineGutterRef}
            aria-hidden
            className="flex-shrink-0 w-10 overflow-hidden select-none bg-gray-50 border-r border-gray-100 text-right font-mono text-xs leading-6 text-gray-400 py-4"
          >
            <div className="px-2">
              {lineNumbers.map((n) => (
                <div key={n}>{n}</div>
              ))}
            </div>
          </div>
          <div
            ref={editorScrollRef}
            onScroll={handleEditorScroll}
            className="flex-1 min-w-0 min-h-0 overflow-auto"
          >
            <textarea
              ref={textareaRef}
              value={value}
              onChange={(e) => onChange(e.target.value)}
              onDrop={handleDrop}
              onPaste={handlePaste}
              onKeyDown={handleKeyDown}
              className="block w-full min-h-full font-mono text-sm leading-6 p-4 resize-none focus:ring-0 focus:outline-none bg-white border-0"
              style={{ minHeight: "100%" }}
              placeholder={"# 标题\n\n支持 **Markdown**、表格与 ```mermaid 图表```…"}
              spellCheck={false}
              rows={Math.max(lineCount, 20)}
            />
          </div>
        </div>
      </div>

      {/* Live preview — always renders currently active (debounced) content */}
      <div className="flex-1 min-w-0 min-h-0 flex flex-col bg-white">
        <div className="flex-shrink-0 px-3 py-1.5 text-xs font-medium text-gray-500 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
          <span>实时预览</span>
          {contentKey ? (
            <span className="text-[10px] uppercase tracking-wide text-gray-400">
              {contentKey}
            </span>
          ) : null}
        </div>
        <div
          ref={previewScrollRef}
          onScroll={handlePreviewScroll}
          className="flex-1 min-h-0 overflow-auto p-4"
        >
          <MarkdownHtmlPreview
            html={previewHtml}
            className="markdown-preview article-typography max-w-none text-sm leading-relaxed
              [&_h1]:text-2xl [&_h1]:font-bold [&_h1]:mb-3
              [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:mb-2 [&_h2]:mt-4
              [&_h3]:text-lg [&_h3]:font-semibold [&_h3]:mb-2 [&_h3]:mt-3
              [&_p]:mb-3 [&_p]:leading-relaxed
              [&_ul]:list-disc [&_ul]:pl-6 [&_ul]:mb-3
              [&_ol]:list-decimal [&_ol]:pl-6 [&_ol]:mb-3
              [&_blockquote]:border-l-4 [&_blockquote]:border-gray-300 [&_blockquote]:pl-3 [&_blockquote]:text-gray-600 [&_blockquote]:italic
              [&_a]:text-blue-600 [&_a]:underline
              [&_code]:bg-gray-100 [&_code]:px-1 [&_code]:rounded [&_code]:text-[0.9em]
              [&_pre]:bg-gray-900 [&_pre]:text-gray-100 [&_pre]:p-3 [&_pre]:rounded-lg [&_pre]:overflow-x-auto [&_pre]:mb-3
              [&_pre_code]:bg-transparent [&_pre_code]:p-0
              [&_img]:max-w-full [&_img]:rounded
              [&_table]:w-full [&_table]:border-collapse [&_th]:border [&_td]:border [&_th]:px-2 [&_td]:px-2 [&_th]:bg-gray-50
              [&_.mermaid]:my-4 [&_.mermaid]:flex [&_.mermaid]:justify-center [&_.mermaid]:overflow-x-auto"
          />
        </div>
      </div>
    </div>
  );
}
