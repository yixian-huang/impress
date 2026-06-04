import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { useThemeSettings } from "@/plugins/hooks";
import { SITE_CONFIG_FEATURES_DEFAULT } from "@/types/siteConfig";

/**
 * Site features.blog.comments + theme article.showComments + per-content allowComments.
 */
export function useCommentsEnabled(contentAllowed = true): boolean {
  const { features } = useGlobalConfig();
  const themeSettings = useThemeSettings();

  const blog = features?.blog ?? SITE_CONFIG_FEATURES_DEFAULT.blog;
  const siteOn = blog.comments !== false;
  const themeOn = themeSettings["article.showComments"] !== false;

  return contentAllowed && siteOn && themeOn;
}
