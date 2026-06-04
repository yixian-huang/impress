import type { ReactNode } from "react";
import CommentSection, { type CommentSectionProps } from "./CommentSection";
import { useCommentsEnabled } from "../useCommentsEnabled";

export interface CommentSlotProps extends CommentSectionProps {
  /** Per-content override (e.g. article `allowComments`). Defaults to true. */
  contentAllowed?: boolean;
  fallback?: ReactNode;
}

/**
 * Public mount point: site features + theme setting + per-content allowComments.
 */
export default function CommentSlot({
  contentAllowed = true,
  fallback = null,
  ...sectionProps
}: CommentSlotProps) {
  const enabled = useCommentsEnabled(contentAllowed);
  if (!enabled) return fallback ? <>{fallback}</> : null;

  return <CommentSection {...sectionProps} />;
}
