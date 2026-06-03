import type { ReactNode } from "react";

interface BlogPageShellProps {
  children: ReactNode;
}

/** Narrow reading column shared by blog-first public pages. */
export default function BlogPageShell({ children }: BlogPageShellProps) {
  return (
    <div className="max-w-3xl mx-auto px-4 md:px-content py-section-sm flex-1 w-full">
      {children}
    </div>
  );
}
