import type { HeaderBrandMode } from "@/types/siteConfig";
import type { BrandingView } from "@/hooks/useBranding";

/** On theme home, keep chrome minimal so AuthorIntro owns the site title. */
export function resolveBlogHomeBrandMode(
  brandMode: HeaderBrandMode,
  branding: BrandingView,
  compactHome: boolean,
): HeaderBrandMode {
  if (!compactHome || brandMode === "none") return brandMode;
  if (brandMode === "logo" && branding.logo.light?.trim()) return "logo";
  if (branding.author.avatar?.trim()) return "avatar";
  if (brandMode === "logo") return "logo";
  return "none";
}
