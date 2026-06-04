import { useLocation } from "react-router-dom";
import { useIsReadingLayout } from "@/plugins/hooks";
import type { CommentContentType } from "./api";

/**
 * Unified comment typography applies when the active theme is blog/reading,
 * or when comments are shown on a public blog article URL.
 */
export function useCommentReadingUi(contentType?: CommentContentType): boolean {
  const isReadingTheme = useIsReadingLayout();
  const { pathname } = useLocation();
  const isBlogArticle =
    contentType === "article" && /^\/blog\/[^/]+\/?$/.test(pathname.replace(/\/+$/, "") || pathname);

  return isReadingTheme || isBlogArticle;
}
