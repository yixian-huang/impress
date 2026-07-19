import { PRODUCT_BRAND } from "@/config/productBrand";

interface ProductLogoProps {
  collapsed?: boolean;
  className?: string;
}

export function ProductLogo({ collapsed = false, className = "" }: ProductLogoProps) {
  return (
    <span className={`inline-flex items-center gap-2 ${className}`} aria-label={PRODUCT_BRAND.fullName}>
      <img className="h-8 w-8" src="/brand/inkless-mark.svg" alt="Inkless" />
      {!collapsed && (
        <span className="text-lg font-bold tracking-normal text-blue-400">{PRODUCT_BRAND.name}</span>
      )}
    </span>
  );
}
