import {
  ProductPoweredBy,
  useBranding,
  useGlobalConfig,
  useThemePages,
  type FooterChromeProps,
} from "@inkless/theme-host";

/** Coerce legacy/partial LocalizedString bags to a safe display string. */
function asDisplayText(value: unknown): string {
  if (typeof value === "string") return value;
  if (value && typeof value === "object" && !Array.isArray(value)) {
    const o = value as Record<string, unknown>;
    for (const key of ["zh", "en"] as const) {
      const v = o[key];
      if (typeof v === "string" && v.trim()) return v;
    }
  }
  return "";
}

export default function CorporateFooter({ config }: FooterChromeProps) {
  const { config: globalConfig } = useGlobalConfig();
  const { footerNavItems } = useThemePages();
  const branding = useBranding();
  const globalFooter = globalConfig.footer || {};

  const style = config?.style ?? "full";
  const logoSrc = config?.logo ?? branding.logo.light?.trim();
  const logoAlt = branding.siteName || "Site";
  const address =
    asDisplayText(config?.address) || asDisplayText(globalFooter.address);
  const phone =
    asDisplayText(config?.phone) || asDisplayText(globalFooter.phone);
  const links = footerNavItems.length > 0
    ? footerNavItems.map((item) => ({ label: item.label, href: item.path }))
    : (globalFooter.links ?? []);
  // Prefer branding (always string); never render a {zh}/{en} object as a child.
  const copyright =
    asDisplayText(config?.copyright) ||
    asDisplayText(globalFooter.copyright) ||
    branding.footer.copyright;

  if (style === "none") {
    return null;
  }

  if (style === "minimal") {
    return (
      <footer className="bg-primary text-on-primary">
        <div className="max-w-layout mx-auto px-4 md:px-content xl:px-8 py-6">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            {logoSrc && (
              <img src={logoSrc} alt={logoAlt} className="h-8 w-auto" />
            )}
            <p className="text-sm text-gray-300">{copyright}</p>
          </div>
        </div>
      </footer>
    );
  }

  const sections = config?.sections ?? [];

  return (
    <footer className="bg-primary text-on-primary">
      <div className="max-w-layout mx-auto px-4 md:px-content xl:px-8 py-12">
        {sections.length > 0 ? (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-8 xl:gap-10">
            <div>
              {logoSrc ? (
                <img src={logoSrc} alt={logoAlt} className="h-10 w-auto mb-4" />
              ) : (
                <div className="mb-4 text-lg font-semibold">{logoAlt}</div>
              )}
              <div className="space-y-2 text-sm text-on-primary/70">
                {address && <p>{address}</p>}
                {phone && <p>{phone}</p>}
              </div>
            </div>
            {sections.map((section, index) => (
              <div key={section.title || String(index)}>
                {section.title && (
                  <h3 className="text-sm font-semibold uppercase tracking-wider mb-4 text-on-primary">
                    {section.title}
                  </h3>
                )}
                {section.links && section.links.length > 0 && (
                  <ul className="space-y-2">
                    {section.links.map((link, linkIndex) => (
                      <li key={link.href || link.label || String(linkIndex)}>
                        <a
                          href={link.href || "#"}
                          className="text-sm text-on-primary/70 hover:text-accent transition-colors cursor-pointer"
                        >
                          {link.label}
                        </a>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            ))}
          </div>
        ) : (
          <div className="flex flex-col md:flex-row md:items-start gap-8 xl:gap-10">
            <div>
              {logoSrc ? (
                <img src={logoSrc} alt={logoAlt} className="h-10 w-auto mb-4" />
              ) : (
                <div className="mb-4 text-lg font-semibold">{logoAlt}</div>
              )}
              <div className="space-y-2 text-sm text-gray-300">
                {address && <p>{address}</p>}
                {phone && <p>{phone}</p>}
              </div>
            </div>
            {links.length > 0 && (
              <div className="md:ml-auto">
                <ul className="flex flex-wrap gap-4 text-sm">
                  {links.map((link, index) => (
                    <li key={link.href || link.label || String(index)}>
                      <a
                        href={link.href || "#"}
                        className="text-gray-300 hover:text-accent transition-colors cursor-pointer"
                      >
                        {link.label}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        )}
        <div className="mt-12 pt-8 border-t border-white/20 text-center space-y-2">
          <p className="text-sm text-gray-300">{copyright}</p>
          {branding.footer.icp && (
            <p className="text-xs text-gray-400 mt-1">{branding.footer.icp}</p>
          )}
          <ProductPoweredBy className="text-xs text-gray-400" />
        </div>
      </div>
    </footer>
  );
}
