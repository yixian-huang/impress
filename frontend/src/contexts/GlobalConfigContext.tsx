import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import {
  fetchPublicContent,
  normalizeConfigForLocale,
  type Locale,
} from "@/api/publicContent";
import { useBootstrap } from "@/contexts/BootstrapContext";
import { resolveLocale } from "@/utils/locale";

interface MediaRef {
  url?: string;
  alt?: string;
}

interface NavItem {
  label?: string;
  href?: string;
}

interface LinkItem {
  label?: string;
  href?: string;
}

export interface GlobalConfig {
  branding?: {
    logo?: MediaRef;
    companyName?: string;
  };
  nav?: {
    items?: NavItem[];
  };
  footer?: {
    address?: string;
    phone?: string;
    links?: LinkItem[];
    copyright?: string;
  };
}

interface GlobalConfigContextValue {
  config: GlobalConfig;
  loading: boolean;
  locale: Locale;
  features: Record<string, { enabled?: boolean }>;
  refetch: () => Promise<void>;
}

const GlobalConfigContext = createContext<GlobalConfigContextValue>({
  config: {},
  loading: true,
  locale: "zh",
  features: {},
  refetch: async () => {},
});

export function GlobalConfigProvider({ children }: { children: ReactNode }) {
  const { i18n } = useTranslation("common");
  const locale = resolveLocale(i18n.language);

  const { data: bootstrapData, isLoading: bootstrapLoading } = useBootstrap();
  const [config, setConfig] = useState<GlobalConfig>({});
  const [loading, setLoading] = useState(true);
  const features = bootstrapData?.features ?? {};

  // Use bootstrap data for initial load
  useEffect(() => {
    if (bootstrapLoading) return;

    const globalData = bootstrapData?.globalConfig;
    if (globalData?.config) {
      const normalized = normalizeConfigForLocale(
        globalData.config as Record<string, unknown>,
        locale
      );
      setConfig(normalized as GlobalConfig);
    }
    setLoading(false);
  }, [bootstrapData, bootstrapLoading, locale]);

  // refetch still uses the direct API for manual refresh scenarios (e.g. admin edits)
  const doFetch = useCallback(async () => {
    try {
      const data = await fetchPublicContent("global", locale);
      const normalized = normalizeConfigForLocale(data.config, locale);
      setConfig(normalized as GlobalConfig);
    } catch {
      // Keep previous config on error
    } finally {
      setLoading(false);
    }
  }, [locale]);

  return (
    <GlobalConfigContext.Provider value={{ config, loading, locale, features, refetch: doFetch }}>
      {children}
    </GlobalConfigContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useGlobalConfig(): GlobalConfigContextValue {
  return useContext(GlobalConfigContext);
}
