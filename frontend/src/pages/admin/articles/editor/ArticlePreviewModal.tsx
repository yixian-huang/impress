import { useEffect, useRef, useState } from "react";
import ArticleTypographyRoot from "@/components/blog/ArticleTypographyRoot";
import ArticlePostBody from "@/components/blog/ArticlePostBody";

export interface ArticlePreviewData {
  title: string;
  bodyHtml: string;
  coverImage?: string;
  author?: string;
  langLabel: string;
  statusLabel: string;
  /** When set, show "open live page" action */
  publicPath?: string | null;
  metadata?: Record<string, unknown>;
}

/**
 * Full-screen reading preview of the current editor state (works for drafts
 * and unsaved content — no server round-trip required).
 */
export default function ArticlePreviewModal({
  open,
  data,
  onClose,
}: {
  open: boolean;
  data: ArticlePreviewData | null;
  onClose: () => void;
}) {
  const contentRef = useRef<HTMLElement>(null);
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        if (lightboxSrc) setLightboxSrc(null);
        else onClose();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, lightboxSrc, onClose]);

  if (!open || !data) return null;

  return (
    <div className="fixed inset-0 z-[60] flex flex-col bg-black/40" role="dialog" aria-modal="true">
      <div className="flex-shrink-0 flex items-center justify-between gap-3 px-4 py-2.5 bg-white border-b border-gray-200 shadow-sm">
        <div className="flex items-center gap-2 min-w-0">
          <span className="text-sm font-semibold text-gray-900">预览</span>
          <span className="text-[10px] px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-100">
            {data.langLabel}
          </span>
          <span className="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 text-gray-600">
            {data.statusLabel}
          </span>
          <span className="text-xs text-gray-400 truncate hidden sm:inline">
            基于当前编辑器内容（含未保存更改）
          </span>
        </div>
        <div className="flex items-center gap-2 flex-shrink-0">
          {data.publicPath && (
            <a
              href={data.publicPath}
              target="_blank"
              rel="noopener noreferrer"
              className="px-3 py-1.5 text-xs border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700"
            >
              打开线上页 ↗
            </a>
          )}
          <button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-900 border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            关闭 (Esc)
          </button>
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto bg-slate-50">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 py-8">
          <ArticleTypographyRoot
            mode="reading"
            articleMetadata={data.metadata}
            className="article-public-view bg-white rounded-xl shadow-sm border border-gray-100 px-6 sm:px-10 py-8"
          >
            {data.coverImage ? (
              <img
                src={data.coverImage}
                alt=""
                className="w-full max-h-72 object-cover rounded-lg mb-6 border border-gray-100"
              />
            ) : null}
            <h1 className="text-3xl font-bold text-gray-900 mb-3 leading-tight">
              {data.title || <span className="text-gray-300 italic">（无标题）</span>}
            </h1>
            {(data.author || data.statusLabel) && (
              <div className="flex flex-wrap items-center gap-3 text-sm text-gray-500 mb-8 pb-4 border-b border-gray-100">
                {data.author ? <span>{data.author}</span> : null}
                <span className="text-gray-300">·</span>
                <span>{data.statusLabel}</span>
              </div>
            )}
            {data.bodyHtml && data.bodyHtml !== "<p></p>" ? (
              <ArticlePostBody
                html={data.bodyHtml}
                contentRef={contentRef}
                onClick={(e) => {
                  const t = e.target as HTMLElement;
                  if (t.tagName === "IMG") {
                    const src = (t as HTMLImageElement).src;
                    if (src) setLightboxSrc(src);
                  }
                }}
              />
            ) : (
              <p className="text-gray-400 italic text-sm">正文为空</p>
            )}
          </ArticleTypographyRoot>
        </div>
      </div>

      {lightboxSrc && (
        <div
          className="fixed inset-0 z-[70] flex items-center justify-center bg-black/70 cursor-zoom-out"
          onClick={() => setLightboxSrc(null)}
          role="presentation"
        >
          <img
            src={lightboxSrc}
            alt=""
            className="max-w-[90vw] max-h-[90vh] object-contain rounded-lg shadow-2xl"
          />
        </div>
      )}
    </div>
  );
}
