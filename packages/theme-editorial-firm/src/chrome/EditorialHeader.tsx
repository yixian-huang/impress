import { Link } from "react-router-dom";
import {
  BaseSiteHeader,
  useBranding,
  useHeaderScroll,
  type HeaderChromeProps,
} from "@inkless/theme-host";

/**
 * Paper/editorial sticky header.
 * Solidifies surface on scroll; logo or uppercase wordmark; host nav + locale via BaseSiteHeader.
 */
export default function EditorialHeader({ config }: HeaderChromeProps) {
  const branding = useBranding();
  const style = config?.style ?? "sticky";
  const showScrollEffect = style !== "static";
  const scrolled = useHeaderScroll(showScrollEffect);
  const logoSrc = config?.logo ?? branding.logo.light?.trim();
  const siteName =
    typeof branding.siteName === "string" && branding.siteName.trim()
      ? branding.siteName
      : "Site";
  const isSticky = style === "sticky" || style === "transparent";
  const solid = !showScrollEffect || scrolled;

  return (
    <BaseSiteHeader
      config={config}
      variant="blog"
      scrolled={solid}
      sticky={isSticky}
      languagePlacement="inline"
      showMobileLanguagePanel
      headerClassName={`transition-all duration-300 ${
        solid
          ? "bg-surface/95 backdrop-blur border-b border-border"
          : "bg-transparent border-b border-transparent"
      }`}
      navPaddingClassName="py-5"
      brand={(
        <Link to="/" className="inline-flex items-center min-h-[2.25rem]">
          {logoSrc ? (
            <img src={logoSrc} alt={siteName} className="h-9 w-auto" />
          ) : (
            <span className="font-heading tracking-wide uppercase text-sm md:text-base font-semibold text-on-surface">
              {siteName}
            </span>
          )}
        </Link>
      )}
    />
  );
}
