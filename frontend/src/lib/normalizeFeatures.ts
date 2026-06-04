import {
  SITE_CONFIG_FEATURES_DEFAULT,
  type SiteConfigFeatures,
} from "@/types/siteConfig";

function isRecord(v: unknown): v is Record<string, unknown> {
  return !!v && typeof v === "object" && !Array.isArray(v);
}

/**
 * Merge API features with defaults.
 * Empty object `{}` from bootstrap (no site_configs row) → undefined (FeatureGate old-deploy compat).
 * Legacy `siteMode` in stored config is ignored; active theme is the single source of truth.
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

  return {
    ...SITE_CONFIG_FEATURES_DEFAULT,
    publicPages: {
      ...SITE_CONFIG_FEATURES_DEFAULT.publicPages,
      ...publicPagesRaw,
    } as SiteConfigFeatures["publicPages"],
    blog: {
      ...SITE_CONFIG_FEATURES_DEFAULT.blog,
      comments: blogRaw.comments !== false,
      rss: blogRaw.rss !== false,
      readingMeta: blogRaw.readingMeta !== false,
      wordsPerMinute:
        typeof blogRaw.wordsPerMinute === "number" && blogRaw.wordsPerMinute > 0
          ? blogRaw.wordsPerMinute
          : SITE_CONFIG_FEATURES_DEFAULT.blog.wordsPerMinute,
    },
  };
}
