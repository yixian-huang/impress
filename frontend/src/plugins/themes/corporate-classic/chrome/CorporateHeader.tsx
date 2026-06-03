import { useState, useEffect } from "react";
import { Link, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { resolveLocale } from "@/utils/locale";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import { useBranding } from "@/hooks/useBranding";
import type { HeaderChromeProps } from "@/plugins/types";
import {
  DesktopNavLinks,
  MobileNavPanel,
  useSiteNavigation,
} from "@/theme/layouts/chrome";

export default function CorporateHeader({ config }: HeaderChromeProps) {
  const { i18n } = useTranslation("common");
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const location = useLocation();
  const { isMono, currentLocale } = useLocaleMode();
  const branding = useBranding();
  const navigation = useSiteNavigation(config?.navigation);

  useEffect(() => setIsMobileMenuOpen(false), [location.pathname]);

  const logoSrc = (config?.logo ?? branding.logo.light?.trim()) || "/images/logo.png";
  const logoAlt = branding.siteName || "Site";
  const showLanguageToggle = config?.showLanguageToggle ?? true;
  const style = config?.style ?? "sticky";

  useEffect(() => {
    if (isMono && i18n.language !== currentLocale) {
      i18n.changeLanguage(currentLocale);
    }
  }, [isMono, currentLocale, i18n]);

  useEffect(() => {
    if (style === "static") return;
    const handleScroll = () => {
      const heroEl = document.querySelector("[data-page-hero]");
      if (heroEl) {
        const rect = heroEl.getBoundingClientRect();
        setIsScrolled(rect.bottom <= 80);
      } else {
        setIsScrolled(window.scrollY > 50);
      }
    };
    window.addEventListener("scroll", handleScroll, { passive: true });
    handleScroll();
    return () => window.removeEventListener("scroll", handleScroll);
  }, [style]);

  const toggleLanguage = () => {
    const newLang = resolveLocale(i18n.language) === "zh" ? "en" : "zh";
    i18n.changeLanguage(newLang);
  };

  const isSticky = style === "sticky" || style === "transparent";
  const showScrollEffect = style !== "static";
  const scrolled = showScrollEffect && isScrolled;

  return (
    <header
      className={`${isSticky ? "sticky top-0 left-0 right-0 z-50" : "relative z-50"} transition-all duration-300 ${
        scrolled ? "bg-white/95 backdrop-blur-sm shadow-md" : "bg-transparent"
      }`}
    >
      {showLanguageToggle && !isMono && (
        <div className={`py-1.5 transition-colors duration-300 ${
          scrolled ? "bg-primary text-white" : "bg-transparent text-white/80"
        }`}>
          <div className="max-w-layout mx-auto px-4 md:px-content xl:px-8 flex justify-end">
            <button
              type="button"
              onClick={toggleLanguage}
              className="text-xs hover:opacity-80 transition-opacity cursor-pointer"
            >
              {resolveLocale(i18n.language) === "zh" ? "English" : "\u4E2D\u6587"}
            </button>
          </div>
        </div>
      )}

      <nav className="py-6">
        <div className="max-w-layout mx-auto px-4 md:px-content xl:px-8">
          <div className="flex justify-between items-center">
            <Link to="/" className="shrink-0">
              <img src={logoSrc} alt={logoAlt} className="h-10 w-auto" />
            </Link>

            <DesktopNavLinks items={navigation} variant="corporate" scrolled={scrolled} />

            <button
              type="button"
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              className={`lg:hidden p-2 cursor-pointer transition-colors ${
                scrolled ? "text-gray-700" : "text-white"
              }`}
              aria-label="Toggle menu"
            >
              <div className="w-5 h-4 flex flex-col justify-between">
                <span className={`block h-0.5 w-full bg-current transition-all duration-200 ${isMobileMenuOpen ? "rotate-45 translate-y-[7px]" : ""}`} />
                <span className={`block h-0.5 w-full bg-current transition-all duration-200 ${isMobileMenuOpen ? "opacity-0" : ""}`} />
                <span className={`block h-0.5 w-full bg-current transition-all duration-200 ${isMobileMenuOpen ? "-rotate-45 -translate-y-[7px]" : ""}`} />
              </div>
            </button>
          </div>

          <MobileNavPanel
            items={navigation}
            open={isMobileMenuOpen}
            onNavigate={() => setIsMobileMenuOpen(false)}
            variant="corporate"
          />

          {showLanguageToggle && !isMono && isMobileMenuOpen && (
            <div className="lg:hidden pt-3">
              <button
                type="button"
                onClick={() => { toggleLanguage(); setIsMobileMenuOpen(false); }}
                className="text-sm text-gray-500 hover:text-blue-600 transition-colors cursor-pointer"
              >
                {resolveLocale(i18n.language) === "zh" ? "Switch to English" : "切换到中文"}
              </button>
            </div>
          )}
        </div>
      </nav>
    </header>
  );
}
