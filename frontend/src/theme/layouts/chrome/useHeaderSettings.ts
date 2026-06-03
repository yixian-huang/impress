import { useMemo } from "react";
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { useThemeSettings } from "@/plugins/hooks";
import type { HeaderBrandMode } from "@/types/siteConfig";
import { SITE_CONFIG_GLOBAL_DEFAULT } from "@/types/siteConfig";

export interface ResolvedHeaderSettings {
  brandMode: HeaderBrandMode;
  showRssLink: boolean;
  showSocials: boolean;
}

/** Merge theme settingSchema defaults with Site Config header overrides. */
export function useHeaderSettings(): ResolvedHeaderSettings {
  const { config } = useGlobalConfig();
  const themeSettings = useThemeSettings();
  const siteHeader = config.siteConfig?.header;

  return useMemo(() => {
    const themeBrandMode = themeSettings["header.brandMode"] as HeaderBrandMode | undefined;
    const themeShowRss = themeSettings["header.showRssLink"] as boolean | undefined;
    const themeShowSocials = themeSettings["header.showSocials"] as boolean | undefined;

    return {
      brandMode: siteHeader?.brandMode ?? themeBrandMode ?? "logo",
      showRssLink: siteHeader?.showRssLink ?? themeShowRss ?? false,
      showSocials: siteHeader?.showSocials ?? themeShowSocials ?? false,
    };
  }, [siteHeader, themeSettings]);
}

export function useSiteConfigGlobal() {
  const { config } = useGlobalConfig();
  return config.siteConfig ?? SITE_CONFIG_GLOBAL_DEFAULT;
}
