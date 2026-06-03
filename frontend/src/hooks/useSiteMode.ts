import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { resolveSiteMode } from "@/lib/normalizeFeatures";
import type { SiteConfigFeatures, SiteMode } from "@/types/siteConfig";

/** Resolved site mode; missing features record → corporate. */
export function useSiteMode(): SiteMode {
  const { features } = useGlobalConfig();
  return resolveSiteMode(features);
}

export function isBlogSiteMode(features: SiteConfigFeatures | undefined): boolean {
  return resolveSiteMode(features) === "blog";
}

export function isHomePageDef(pageDef: { slug?: string; contentKey?: string }): boolean {
  return pageDef.slug === "home" || pageDef.contentKey === "home";
}
