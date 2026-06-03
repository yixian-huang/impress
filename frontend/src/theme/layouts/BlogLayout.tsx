import type { ReactNode } from "react";
import SiteLayout from "./PublicLayout";
import RssHeadLink from "@/components/feature/RssHeadLink";

/** Blog public pages shell; chrome comes from active blog-first theme. */
export default function BlogLayout({ children }: { children: ReactNode }) {
  return (
    <SiteLayout>
      <RssHeadLink />
      {children}
    </SiteLayout>
  );
}
