import { useBranding } from "@/hooks/useBranding";
import { useContentMaxWidth } from "@/plugins/hooks";
import type { FooterChromeProps } from "@/plugins/types";
import ProductPoweredBy from "@/components/feature/ProductPoweredBy";

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

export default function MinimalFooter({ config }: FooterChromeProps) {
  const branding = useBranding();
  const maxWidth = useContentMaxWidth();
  const copyright =
    asDisplayText(config?.copyright) || branding.footer.copyright;
  const style = config?.style ?? "minimal";

  if (style === "none") return null;

  return (
    <footer className="mt-auto border-t border-border">
      <div className="mx-auto px-4 md:px-content py-6 w-full space-y-2" style={{ maxWidth }}>
        <p className="text-sm text-on-surface-muted text-center">{copyright}</p>
        <div className="text-center">
          <ProductPoweredBy />
        </div>
      </div>
    </footer>
  );
}
