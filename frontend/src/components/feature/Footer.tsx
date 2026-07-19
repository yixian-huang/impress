/**
 * @deprecated Use active theme layoutChrome (CorporateFooter / BlogFooter) via SiteLayout instead.
 */
import { useBranding } from "@/hooks/useBranding";
import { pickLocaleValue } from "@/lib/locale";

export default function Footer() {
  const branding = useBranding();
  const logoSrc = branding.logo.light;
  const logoAlt = branding.siteName || 'Site';

  return (
    <footer className="bg-primary text-white">
      <div className="max-w-layout mx-auto px-4 md:px-6 py-12">
        <div className="flex flex-col md:flex-row md:items-start gap-8">
          <div>
            {logoSrc ? (
              <img src={logoSrc} alt={logoAlt} className="h-10 w-auto mb-4" />
            ) : (
              <div className="mb-4 text-lg font-semibold">{logoAlt}</div>
            )}
            {branding.author.bio && <p className="text-sm text-gray-300 max-w-md">{branding.author.bio}</p>}
          </div>
          {branding.footer.extraLinks.length > 0 && (
            <div className="md:ml-auto">
              <ul className="flex flex-wrap gap-4 text-sm">
                {branding.footer.extraLinks.map((link, i) => (
                  <li key={i}>
                    <a
                      href={link.url || '#'}
                      className="text-gray-300 hover:text-accent transition-colors cursor-pointer"
                    >
                      {pickLocaleValue({
                        value: link.label,
                        mode: branding.localeMode,
                        defaultLocale: branding.defaultLocale,
                        currentLocale: branding.currentLocale,
                      })}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
        <div className="mt-12 pt-8 border-t border-white/20 text-center text-sm text-gray-300">
          <p>{branding.footer.copyright}</p>
          {branding.footer.icp && <p className="mt-1">{branding.footer.icp}</p>}
        </div>
      </div>
    </footer>
  );
}
