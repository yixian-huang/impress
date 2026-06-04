import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { useThemeSettings } from "@/plugins/hooks";
import { SITE_CONFIG_FEATURES_DEFAULT } from "@/types/siteConfig";

/** Site features + active theme settings for article reading meta. */
export function useArticleReadingMeta() {
  const { features } = useGlobalConfig();
  const themeSettings = useThemeSettings();

  const blog = features?.blog ?? SITE_CONFIG_FEATURES_DEFAULT.blog;
  const siteOn = blog.readingMeta !== false;
  const themeOn = themeSettings["article.showReadingMeta"] !== false;
  const themeWpm = themeSettings["article.wordsPerMinute"];
  const siteWpm = blog.wordsPerMinute;

  const wordsPerMinute =
    (typeof themeWpm === "number" && themeWpm > 0 ? themeWpm : undefined) ??
    (typeof siteWpm === "number" && siteWpm > 0 ? siteWpm : undefined) ??
    280;

  return {
    enabled: siteOn && themeOn,
    wordsPerMinute,
  };
}
