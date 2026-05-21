import type { SiteConfigFeatures } from "@/types/siteConfig";

/** Routes (kebab-case URLs) that are gated by a feature key. */
export const routeFeatureMap: Record<string, keyof SiteConfigFeatures["publicPages"]> = {
  "/about": "about",
  "/experts": "experts",
  "/core-services": "coreServices",
  "/advantages": "advantages",
  "/cases": "cases",
};

/** Returns true if the given camelCase feature key is enabled.
 *
 * Backwards-compat rule (spec §4.2): when the features record is missing
 * entirely (old deployment, no migration), every key defaults to TRUE so
 * the existing site keeps behaving as before. Only when a publishedConfig
 * record EXISTS but the specific key is unset does it fall to false.
 * New deployments must rely on BlankSiteSeed to explicitly seed a record
 * with personal-blog defaults.
 */
export function isFeatureEnabled(
  features: SiteConfigFeatures | undefined,
  key: keyof SiteConfigFeatures["publicPages"],
): boolean {
  if (!features || !features.publicPages) return true;
  return features.publicPages[key] === true;
}
