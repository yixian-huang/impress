import type { ReactNode } from "react";
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { SITE_CONFIG_FEATURES_DEFAULT, type SiteConfigFeatures } from "@/types/siteConfig";

export interface BlogFeatureGateProps {
  feature: keyof SiteConfigFeatures["blog"];
  children: ReactNode;
  fallback?: ReactNode;
}

/** Gates blog-specific features (comments, rss). Missing record → enabled (old-deploy compat). */
export function BlogFeatureGate({ feature, children, fallback = null }: BlogFeatureGateProps) {
  const { features } = useGlobalConfig();
  const blog = features?.blog ?? SITE_CONFIG_FEATURES_DEFAULT.blog;
  const enabled = blog[feature] !== false;
  if (!enabled) return <>{fallback}</>;
  return <>{children}</>;
}
