import { useBranding } from "@/hooks/useBranding";
import { useContentMaxWidth } from "@/plugins/hooks";
import type { FooterChromeProps } from "@/plugins/types";
import ProductPoweredBy from "@/components/feature/ProductPoweredBy";

export default function MinimalFooter({ config }: FooterChromeProps) {
  const branding = useBranding();
  const maxWidth = useContentMaxWidth();
  const copyright = config?.copyright ?? branding.footer.copyright;
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
