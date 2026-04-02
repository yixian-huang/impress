import { createContext, useContext, useEffect, useState, useMemo, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { fetchBootstrap, type BootstrapData } from "@/api/bootstrap";
import { resolveLocale } from "@/utils/locale";

interface BootstrapContextValue {
  data: BootstrapData | null;
  isLoading: boolean;
  locale: string;
}

const BootstrapContext = createContext<BootstrapContextValue>({
  data: null,
  isLoading: true,
  locale: "zh",
});

// eslint-disable-next-line react-refresh/only-export-components
export function useBootstrap() {
  return useContext(BootstrapContext);
}

export function BootstrapProvider({ children }: { children: ReactNode }) {
  const { i18n } = useTranslation("common");
  const locale = resolveLocale(i18n.language);
  const [data, setData] = useState<BootstrapData | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    setIsLoading(true);
    fetchBootstrap(locale)
      .then((result) => {
        if (!cancelled) setData(result);
      })
      .catch(() => {
        // Keep null on error — individual providers will use defaults
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => { cancelled = true; };
  }, [locale]);

  const value = useMemo(() => ({
    data,
    isLoading,
    locale,
  }), [data, isLoading, locale]);

  return (
    <BootstrapContext.Provider value={value}>
      {children}
    </BootstrapContext.Provider>
  );
}
