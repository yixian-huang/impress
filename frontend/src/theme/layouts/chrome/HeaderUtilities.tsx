import AuthorSocialLinks from "@/components/blog/AuthorSocialLinks";
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { useHeaderSettings } from "./useHeaderSettings";

interface HeaderUtilitiesProps {
  /** Hide social links (e.g. theme home — socials live on author page). */
  hideSocials?: boolean;
  /** Hide RSS link. */
  hideRss?: boolean;
}

/** RSS + social links for blog-style headers (theme settingSchema + Site Config Header). */
export default function HeaderUtilities({
  hideSocials = false,
  hideRss = false,
}: HeaderUtilitiesProps = {}) {
  const { showRssLink, showSocials } = useHeaderSettings();
  const { features } = useGlobalConfig();
  const rssEnabled = features?.blog?.rss === true;

  const showRss = !hideRss && showRssLink && rssEnabled;
  const showSocial = !hideSocials && showSocials;

  if (!showRss && !showSocial) return null;

  return (
    <div className="hidden lg:flex items-center gap-3 ml-4">
      {showRss && (
        <a
          href="/feed.xml"
          className="text-xs text-on-surface-muted hover:text-primary transition-colors"
          aria-label="RSS feed"
        >
          RSS
        </a>
      )}
      {showSocial && (
        <AuthorSocialLinks
          className="flex items-center gap-3"
          linkClassName="text-xs text-on-surface-muted hover:text-primary transition-colors"
        />
      )}
    </div>
  );
}
