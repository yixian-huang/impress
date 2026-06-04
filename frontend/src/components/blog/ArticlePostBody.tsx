import type { RefObject } from "react";

interface ArticlePostBodyProps {
  html: string;
  contentRef?: RefObject<HTMLElement | null>;
  onClick?: (e: React.MouseEvent) => void;
}

/** Article body HTML — must sit inside `ArticleTypographyRoot`. */
export default function ArticlePostBody({ html, contentRef, onClick }: ArticlePostBodyProps) {
  return (
    <article
      ref={contentRef}
      className="tiptap ProseMirror max-w-none article-public-view"
      dangerouslySetInnerHTML={{ __html: html }}
      onClick={onClick}
    />
  );
}
