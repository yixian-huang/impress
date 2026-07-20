import { PRODUCT_BRAND } from "@/config/productBrand";

interface ProductPoweredByProps {
  className?: string;
}

/** Small host product attribution for public footers. */
export default function ProductPoweredBy({
  className = "text-xs text-on-surface-muted/70",
}: ProductPoweredByProps) {
  return (
    <p className={className}>
      Powered by{" "}
      <a
        href={PRODUCT_BRAND.origin}
        className="underline decoration-border underline-offset-2 hover:text-on-surface transition-colors"
        target="_blank"
        rel="noopener noreferrer"
      >
        {PRODUCT_BRAND.name}
      </a>
    </p>
  );
}
