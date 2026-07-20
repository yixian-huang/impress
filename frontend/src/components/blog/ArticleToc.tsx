import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import type { TocHeading, TocLayout } from "@/utils/articleToc";

function TocList({
  headings,
  activeId,
  onSelect,
  compact = false,
}: {
  headings: TocHeading[];
  activeId: string | null;
  onSelect: (id: string) => void;
  compact?: boolean;
}) {
  return (
    <ul className={compact ? "space-y-1" : "space-y-1.5"}>
      {headings.map((h) => (
        <li key={h.id} className={h.level === 3 ? "pl-2.5" : ""}>
          <button
            type="button"
            onClick={() => onSelect(h.id)}
            className={`text-left leading-snug transition-colors w-full rounded-sm ${
              compact ? "text-xs py-0.5" : "text-sm"
            } ${
              activeId === h.id
                ? "text-primary border-l-2 border-primary pl-2 font-medium"
                : "text-on-surface-muted hover:text-primary pl-2 border-l-2 border-transparent"
            }`}
          >
            {h.text}
          </button>
        </li>
      ))}
    </ul>
  );
}

export function ArticleTocInline({
  headings,
  layout,
  activeId,
  onSelect,
}: {
  headings: TocHeading[];
  layout: TocLayout;
  activeId: string | null;
  onSelect: (id: string) => void;
}) {
  const { t } = useTranslation("common");

  if (layout === "none" || headings.length === 0) return null;

  // Floating panel covers lg+; keep inline collapsible on smaller screens for full layout.
  const hideOnDesktop = layout === "full";

  return (
    <details
      className={`article-page-ui font-sans mb-8 ${hideOnDesktop ? "lg:hidden" : ""}`}
      open={layout === "inline"}
    >
      <summary className="cursor-pointer list-none text-xs uppercase tracking-[0.15em] text-on-surface-muted hover:text-primary transition-colors [&::-webkit-details-marker]:hidden">
        <span className="inline-flex items-center gap-2">
          {t("blog.tocTitle")}
          <span className="text-on-surface-muted/60 normal-case tracking-normal">({headings.length})</span>
        </span>
      </summary>
      <nav className="mt-3 pt-3 border-t border-border/60" aria-label={t("blog.tocTitle")}>
        <TocList headings={headings} activeId={activeId} onSelect={onSelect} />
      </nav>
    </details>
  );
}

/**
 * Viewport-fixed floating TOC — vertically centered on the right.
 * Does not participate in content flow; body keeps full reading width.
 */
export function ArticleTocSidebar({
  headings,
  layout,
  activeId,
  onSelect,
}: {
  headings: TocHeading[];
  layout: TocLayout;
  activeId: string | null;
  onSelect: (id: string) => void;
}) {
  const { t } = useTranslation("common");

  if (layout !== "full" || headings.length === 0) return null;

  return (
    <aside
      className="article-toc-float hidden lg:flex fixed z-30 top-1/2 right-4 xl:right-6 -translate-y-1/2 flex-col article-page-ui font-sans pointer-events-auto"
      aria-label={t("blog.tocTitle")}
    >
      <div className="w-44 xl:w-48 max-h-[min(70vh,32rem)] overflow-y-auto rounded-xl border border-border/70 bg-surface/90 backdrop-blur-md shadow-lg shadow-black/5 px-3 py-3">
        <p className="text-[11px] uppercase tracking-[0.16em] text-on-surface-muted mb-2.5 px-0.5">
          {t("blog.tocTitle")}
        </p>
        <TocList headings={headings} activeId={activeId} onSelect={onSelect} compact />
      </div>
    </aside>
  );
}

/** Full layout: body full-width; floating TOC is portaled via fixed positioning. */
export function ArticleTocLayout({
  layout,
  sidebar,
  children,
}: {
  layout: TocLayout;
  sidebar: ReactNode;
  children: ReactNode;
}) {
  if (layout !== "full") {
    return <>{children}</>;
  }

  return (
    <>
      <div className="min-w-0 w-full">{children}</div>
      {sidebar}
    </>
  );
}
