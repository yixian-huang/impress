import { useState, useEffect } from "react";
import { Link, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { resolveLocale } from "@/utils/locale";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { useBranding } from "@/hooks/useBranding";
import type { HeaderChromeProps } from "@/plugins/types";
import {
  BrandMark,
  DesktopNavLinks,
  MobileNavPanel,
  useHeaderSettings,
  useSiteNavigation,
} from "@/theme/layouts/chrome";

function HeaderUtilities() {
  const { showRssLink, showSocials } = useHeaderSettings();
  const { features } = useGlobalConfig();
  const branding = useBranding();
  const rssEnabled = features?.blog?.rss === true;

  if (!showRssLink && !showSocials) return null;

  return (
    <div className="hidden lg:flex items-center gap-3 ml-4">
      {showRssLink && rssEnabled && (
        <a
          href="/feed.xml"
          className="text-xs text-on-surface-muted hover:text-primary transition-colors"
          aria-label="RSS feed"
        >
          RSS
        </a>
      )}
      {showSocials && branding.author.socials.map((s) => (
        <a
          key={`${s.kind}-${s.url}`}
          href={s.url}
          className="text-xs text-on-surface-muted hover:text-primary transition-colors"
          target="_blank"
          rel="noopener noreferrer"
        >
          {s.label || s.kind}
        </a>
      ))}
    </div>
  );
}

export default function BlogHeader({ config }: HeaderChromeProps) {
  const { i18n } = useTranslation("common");
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const location = useLocation();
  const { isMono, currentLocale } = useLocaleMode();
  const { brandMode } = useHeaderSettings();
  const navigation = useSiteNavigation(config?.navigation);
  const showLanguageToggle = config?.showLanguageToggle ?? true;

  useEffect(() => setIsMobileMenuOpen(false), [location.pathname]);

  useEffect(() => {
    if (isMono && i18n.language !== currentLocale) {
      i18n.changeLanguage(currentLocale);
    }
  }, [isMono, currentLocale, i18n]);

  const toggleLanguage = () => {
    const newLang = resolveLocale(i18n.language) === "zh" ? "en" : "zh";
    i18n.changeLanguage(newLang);
  };

  return (
    <header className="sticky top-0 left-0 right-0 z-50 bg-surface/90 backdrop-blur border-b border-border">
      <nav className="py-3">
        <div className="max-w-3xl mx-auto px-4 md:px-content w-full">
          <div className="flex justify-between items-center gap-4">
            <BrandMark brandMode={brandMode} hideDefaultLogo />

            <div className="flex items-center flex-1 justify-end gap-2">
              <DesktopNavLinks items={navigation} variant="blog" />
              <HeaderUtilities />
              {showLanguageToggle && !isMono && (
                <button
                  type="button"
                  onClick={toggleLanguage}
                  className="hidden lg:inline text-xs text-on-surface-muted hover:text-primary transition-colors cursor-pointer ml-2"
                >
                  {resolveLocale(i18n.language) === "zh" ? "EN" : "中文"}
                </button>
              )}
              <button
                type="button"
                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                className="lg:hidden p-2 cursor-pointer text-on-surface"
                aria-label="Toggle menu"
              >
                <div className="w-5 h-4 flex flex-col justify-between">
                  <span className={`block h-0.5 w-full bg-current transition-all duration-200 ${isMobileMenuOpen ? "rotate-45 translate-y-[7px]" : ""}`} />
                  <span className={`block h-0.5 w-full bg-current transition-all duration-200 ${isMobileMenuOpen ? "opacity-0" : ""}`} />
                  <span className={`block h-0.5 w-full bg-current transition-all duration-200 ${isMobileMenuOpen ? "-rotate-45 -translate-y-[7px]" : ""}`} />
                </div>
              </button>
            </div>
          </div>

          <MobileNavPanel
            items={navigation}
            open={isMobileMenuOpen}
            onNavigate={() => setIsMobileMenuOpen(false)}
            variant="blog"
          />
        </div>
      </nav>
    </header>
  );
}
