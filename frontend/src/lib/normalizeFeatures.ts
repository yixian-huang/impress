import {
  SITE_CONFIG_FEATURES_DEFAULT,
  type SiteConfigFeatures,
  type SiteMode,
} from "@/types/siteConfig";

function isRecord(v: unknown): v is Record<string, unknown> {
  return !!v && typeof v === "object" && !Array.isArray(v);
}

/** Personal-blog feature preset without explicit siteMode (BlankSiteSeed pattern). */
function looksLikeBlogFirstPublicPages(
  publicPages: SiteConfigFeatures["publicPages"] | undefined,
): boolean {
  if (!publicPages) return false;
  return (
    publicPages.blog === true &&
    publicPages.home === true &&
    publicPages.about === false &&
    publicPages.experts === false &&
    publicPages.coreServices === false &&
    publicPages.advantages === false &&
    publicPages.cases === false
  );
}

function inferSiteMode(f: SiteConfigFeatures): SiteMode {
  if (f.siteMode === "blog" || f.siteMode === "corporate") {
    return f.siteMode;
  }
  if (looksLikeBlogFirstPublicPages(f.publicPages)) {
    return "blog";
  }
  return "corporate";
}

/**
 * Merge API features with defaults and resolve siteMode.
 * Empty object `{}` from bootstrap (no site_configs row) → undefined (FeatureGate old-deploy compat).
 */
export function normalizeFeatures(raw: unknown): SiteConfigFeatures | undefined {
  if (raw === undefined || raw === null) {
    return undefined;
  }
  if (!isRecord(raw) || Object.keys(raw).length === 0) {
    return undefined;
  }

  const publicPagesRaw = isRecord(raw.publicPages) ? raw.publicPages : {};
  const blogRaw = isRecord(raw.blog) ? raw.blog : {};

  const merged: SiteConfigFeatures = {
    ...SITE_CONFIG_FEATURES_DEFAULT,
    siteMode: raw.siteMode === "blog" ? "blog" : raw.siteMode === "corporate" ? "corporate" : undefined,
    publicPages: {
      ...SITE_CONFIG_FEATURES_DEFAULT.publicPages,
      ...publicPagesRaw,
    } as SiteConfigFeatures["publicPages"],
    blog: {
      ...SITE_CONFIG_FEATURES_DEFAULT.blog,
      comments: blogRaw.comments !== false,
      rss: blogRaw.rss !== false,
    },
  };

  merged.siteMode = inferSiteMode(merged);
  return merged;
}

export function resolveSiteMode(features: SiteConfigFeatures | undefined): SiteMode {
  if (!features) return "corporate";
  return inferSiteMode(features);
}
