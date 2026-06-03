import { Link } from "react-router-dom";
import { useBranding } from "@/hooks/useBranding";
import type { HeaderBrandMode } from "@/types/siteConfig";

interface BrandMarkProps {
  brandMode: HeaderBrandMode;
  /** When true, omit default placeholder logo image */
  hideDefaultLogo?: boolean;
  className?: string;
  textClassName?: string;
  logoClassName?: string;
  avatarClassName?: string;
}

export default function BrandMark({
  brandMode,
  hideDefaultLogo = false,
  className = "",
  textClassName = "text-lg font-heading font-semibold text-on-surface",
  logoClassName = "h-8 w-auto",
  avatarClassName = "h-8 w-8 rounded-full object-cover border border-border",
}: BrandMarkProps) {
  const branding = useBranding();
  const displayName = branding.author.name?.trim() || branding.siteName;
  const logoSrc = branding.logo.light?.trim();
  const avatarSrc = branding.author.avatar?.trim();

  if (brandMode === "none") {
    return null;
  }

  return (
    <Link to="/" className={`shrink-0 inline-flex items-center gap-2 ${className}`}>
      {brandMode === "logo" && logoSrc && (
        <img src={logoSrc} alt={displayName} className={logoClassName} />
      )}
      {brandMode === "logo" && !logoSrc && !hideDefaultLogo && (
        <img src="/images/logo.png" alt={displayName} className={logoClassName} />
      )}
      {brandMode === "avatar" && avatarSrc && (
        <img src={avatarSrc} alt={displayName} className={avatarClassName} />
      )}
      {brandMode === "avatar" && (
        <span className={textClassName}>{displayName}</span>
      )}
      {brandMode === "text" && (
        <span className={textClassName}>{displayName}</span>
      )}
    </Link>
  );
}
