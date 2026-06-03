import { useEffect } from "react";
import { BlogFeatureGate } from "@/components/feature/BlogFeatureGate";

/** Injects RSS alternate link into document head when blog.rss is enabled. */
export default function RssHeadLink() {
  return (
    <BlogFeatureGate feature="rss">
      <RssLinkInjector />
    </BlogFeatureGate>
  );
}

function RssLinkInjector() {
  useEffect(() => {
    let link = document.querySelector<HTMLLinkElement>('link[rel="alternate"][type="application/rss+xml"]');
    if (!link) {
      link = document.createElement("link");
      link.rel = "alternate";
      link.type = "application/rss+xml";
      document.head.appendChild(link);
    }
    link.title = "RSS Feed";
    link.href = "/feed.xml";

    return () => {
      link?.remove();
    };
  }, []);

  return null;
}
