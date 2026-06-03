import { useMemo } from "react";
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { useThemePages } from "@/contexts/ThemePagesContext";
import { isFeatureEnabled, routeFeatureMap } from "@/router/featureMap";
import type { NavItem } from "@/theme/layouts/types";

export interface SiteNavItem {
  label?: string;
  path?: string;
  children?: SiteNavItem[];
}

function filterByFeatures(
  items: SiteNavItem[],
  features: ReturnType<typeof useGlobalConfig>["features"],
): SiteNavItem[] {
  const result: SiteNavItem[] = [];
  for (const item of items) {
    const children = item.children?.length
      ? filterByFeatures(item.children, features)
      : undefined;
    const path = item.path || "/";
    const featureKey = routeFeatureMap[path];
    if (featureKey && !isFeatureEnabled(features, featureKey)) {
      continue;
    }
    result.push({
      label: item.label,
      path: item.path,
      children: children?.length ? children : undefined,
    });
  }
  return result;
}

/** Resolve public header navigation: menu > theme pages > layout override > legacy global nav. */
export function useSiteNavigation(configNavigation?: NavItem[]): SiteNavItem[] {
  const { config: globalConfig, features } = useGlobalConfig();
  const { headerNavItems, menuNavItems } = useThemePages();

  return useMemo(() => {
    let navigation: SiteNavItem[];

    if (configNavigation?.length) {
      navigation = configNavigation.map((item) => ({
        label: item.label,
        path: item.path,
        children: item.children?.map((c) => ({
          label: c.label,
          path: c.path,
          children: c.children,
        })),
      }));
    } else if (menuNavItems.length > 0) {
      navigation = menuNavItems.map((item) => ({
        label: item.label,
        path: item.path,
        children: item.children,
      }));
    } else if (headerNavItems.length > 0) {
      navigation = headerNavItems.map((item) => ({
        label: item.label,
        path: item.path,
      }));
    } else {
      navigation = (globalConfig.nav?.items || []).map((item) => ({
        label: item.label,
        path: item.href,
      }));
    }

    return filterByFeatures(navigation, features);
  }, [configNavigation, menuNavItems, headerNavItems, globalConfig.nav?.items, features]);
}
