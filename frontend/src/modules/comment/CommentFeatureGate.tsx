import type { ReactNode } from "react";
import { BlogFeatureGate } from "@/components/feature/BlogFeatureGate";

interface CommentFeatureGateProps {
  children: ReactNode;
  fallback?: ReactNode;
}

/** Site-level `features.blog.comments` gate for the comment module. */
export function CommentFeatureGate({ children, fallback = null }: CommentFeatureGateProps) {
  return (
    <BlogFeatureGate feature="comments" fallback={fallback}>
      {children}
    </BlogFeatureGate>
  );
}
