import type { ReactNode } from "react";
import PublicLayout from "./PublicLayout";
import RssHeadLink from "@/components/feature/RssHeadLink";

/** Blog-first public pages: same chrome as PublicLayout without corporate PageHero blocks. */
export default function BlogLayout({ children }: { children: ReactNode }) {
  return (
    <PublicLayout>
      <RssHeadLink />
      {children}
    </PublicLayout>
  );
}
