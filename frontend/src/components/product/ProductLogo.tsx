import { PRODUCT_BRAND } from "@/config/productBrand";

interface ProductLogoProps {
  collapsed?: boolean;
  className?: string;
  /** Use light wordmark on dark admin sidebar. */
  variant?: "default" | "onDark";
}

export function ProductLogo({
  collapsed = false,
  className = "",
  variant = "default",
}: ProductLogoProps) {
  const onDark = variant === "onDark";
  return (
    <span
      className={`inline-flex items-center gap-2.5 ${className}`}
      aria-label={PRODUCT_BRAND.fullName}
    >
      <img
        className="h-8 w-8 drop-shadow-sm"
        src="/brand/inkless-mark.svg"
        alt="Inkless"
      />
      {!collapsed && (
        <span
          className={`text-[1.05rem] font-semibold tracking-tight ${
            onDark ? "text-white" : "text-slate-900"
          }`}
        >
          {PRODUCT_BRAND.name}
          {onDark ? (
            <span className="ml-1.5 text-[10px] font-medium uppercase tracking-[0.12em] text-slate-400">
              CMS
            </span>
          ) : null}
        </span>
      )}
    </span>
  );
}
