import { Link } from "react-router-dom";
import {
  ProductPoweredBy,
  useBranding,
  useThemePages,
  type FooterChromeProps,
} from "@inkless/theme-host";

/** Coerce accidental object labels to a safe display string. */
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

/**
 * Paper/editorial footer — surface-alt + top border (not classic full primary-blue).
 */
export default function EditorialFooter({ config }: FooterChromeProps) {
  const branding = useBranding();
  const { footerNavItems } = useThemePages();
  const style = config?.style ?? "minimal";

  if (style === "none") {
    return null;
  }

  const siteName =
    typeof branding.siteName === "string" && branding.siteName.trim()
      ? branding.siteName
      : "Site";
  const tagline = asDisplayText(branding.tagline);
  const copyright =
    asDisplayText(config?.copyright) || branding.footer.copyright;
  const icp =
    typeof branding.footer.icp === "string" ? branding.footer.icp.trim() : "";

  const links = footerNavItems
    .map((item) => ({
      label: typeof item.label === "string" ? item.label : "",
      path: typeof item.path === "string" ? item.path : "",
    }))
    .filter((item) => item.label && item.path);

  return (
    <footer className="mt-auto bg-surface-alt border-t border-border text-on-surface">
      <div className="max-w-layout mx-auto px-4 md:px-content xl:px-8 py-12 md:py-16">
        <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-10">
          <div className="max-w-md space-y-3">
            <div className="font-heading tracking-wide uppercase text-xl md:text-2xl font-semibold text-on-surface">
              {siteName}
            </div>
            {tagline ? (
              <p className="text-sm text-on-surface-muted leading-relaxed">{tagline}</p>
            ) : null}
          </div>

          {links.length > 0 ? (
            <nav aria-label="Footer">
              <ul className="flex flex-wrap gap-x-6 gap-y-2">
                {links.map((link) => (
                  <li key={link.path}>
                    <Link
                      to={link.path}
                      className="text-sm text-on-surface-muted hover:text-accent transition-colors"
                    >
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </nav>
          ) : null}
        </div>

        <div className="mt-12 pt-8 border-t border-border space-y-2 text-center">
          {copyright ? (
            <p className="text-sm text-on-surface-muted">{copyright}</p>
          ) : null}
          {icp ? (
            <p className="text-xs text-on-surface-muted/80">{icp}</p>
          ) : null}
          <ProductPoweredBy className="text-xs text-on-surface-muted/70" />
        </div>
      </div>
    </footer>
  );
}
