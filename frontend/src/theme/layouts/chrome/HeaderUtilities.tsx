import AuthorSocialLinks from "@/components/blog/AuthorSocialLinks";
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { useHeaderSettings } from "./useHeaderSettings";

/** RSS + social links for blog-style headers (theme settingSchema + Site Config Header). */
export default function HeaderUtilities() {
  const { showRssLink, showSocials } = useHeaderSettings();
  const { features } = useGlobalConfig();
  const rssEnabled = features?.blog?.rss === true;

  if (!showRssLink && !showSocials) return null;

  return (
    <div className="hidden lg:flex items-center gap-3 ml-4">
      {showRssLink && rssEnabled && (
        <a
          href="/feed.xml"
          className="text-xs text-on-surface-muted hover:text-primary transition-colors"
          aria-label="RSS feed"
        >
          RSS
        </a>
      )}
      {showSocials && (
        <AuthorSocialLinks
          className="flex items-center gap-3"
          linkClassName="text-xs text-on-surface-muted hover:text-primary transition-colors"
        />
      )}
    </div>
  );
}
