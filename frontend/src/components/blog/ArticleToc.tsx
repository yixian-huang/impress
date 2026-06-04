import { useEffect, useState, useCallback } from "react";
import { useTranslation } from "react-i18next";
import type { RefObject } from "react";
import type { TocHeading, TocLayout } from "@/utils/articleToc";

export function useActiveHeading(contentRef: RefObject<HTMLElement | null>, headings: TocHeading[]) {
  const [activeId, setActiveId] = useState<string | null>(headings[0]?.id ?? null);

  useEffect(() => {
    const root = contentRef.current;
    if (!root || headings.length === 0) return;

    const elements = headings
      .map((h) => document.getElementById(h.id))
      .filter((el): el is HTMLElement => el != null);

    if (elements.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);
        if (visible.length > 0 && visible[0].target.id) {
          setActiveId(visible[0].target.id);
        }
      },
      { rootMargin: "-20% 0px -55% 0px", threshold: 0 },
    );

    elements.forEach((el) => observer.observe(el));
    return () => observer.disconnect();
  }, [contentRef, headings]);

  return activeId;
}

export function useTocScroll() {
  return useCallback((id: string) => {
    const el = document.getElementById(id);
    if (!el) return;
    el.scrollIntoView({ behavior: "smooth", block: "start" });
    history.replaceState(null, "", `#${id}`);
  }, []);
}

function TocList({
  headings,
  activeId,
  onSelect,
}: {
  headings: TocHeading[];
  activeId: string | null;
  onSelect: (id: string) => void;
}) {
  return (
    <ul className="space-y-1.5">
      {headings.map((h) => (
        <li key={h.id} className={h.level === 3 ? "pl-3" : ""}>
          <button
            type="button"
            onClick={() => onSelect(h.id)}
            className={`text-left text-sm leading-snug transition-colors w-full ${
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
    <aside className="hidden lg:block article-page-ui font-sans" aria-label={t("blog.tocTitle")}>
      <div className="sticky top-8 max-h-[calc(100vh-4rem)] overflow-y-auto">
        <p className="text-xs uppercase tracking-[0.15em] text-on-surface-muted mb-3">{t("blog.tocTitle")}</p>
        <TocList headings={headings} activeId={activeId} onSelect={onSelect} />
      </div>
    </aside>
  );
}

export function ArticleTocLayout({
  layout,
  sidebar,
  children,
}: {
  layout: TocLayout;
  sidebar: React.ReactNode;
  children: React.ReactNode;
}) {
  if (layout !== "full") {
    return <>{children}</>;
  }

  return (
    <div className="lg:grid lg:grid-cols-[minmax(0,1fr)_11rem] lg:gap-x-12 xl:grid-cols-[minmax(0,1fr)_12rem]">
      <div className="min-w-0">{children}</div>
      {sidebar}
    </div>
  );
}
