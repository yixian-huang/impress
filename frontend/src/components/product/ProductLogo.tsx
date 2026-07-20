import { PRODUCT_BRAND } from "@/config/productBrand";

interface ProductLogoProps {
  collapsed?: boolean;
  className?: string;
  /**
   * default — black wordmark on light UI
   * onDark / ink — light wordmark on black rail
   */
  variant?: "default" | "onDark" | "ink";
}

export function ProductLogo({
  collapsed = false,
  className = "",
  variant = "default",
}: ProductLogoProps) {
  const ink = variant === "ink" || variant === "onDark";
  const markSrc = ink ? "/brand/inkless-mark-ink.svg" : "/brand/inkless-mark.svg";

  return (
    <span
      className={`inline-flex items-center gap-2.5 ${className}`}
      aria-label={PRODUCT_BRAND.fullName}
    >
      <img className="h-8 w-8" src={markSrc} alt="Inkless" />
      {!collapsed && (
        <span className="flex flex-col leading-none">
          <span
            className={`text-[1.05rem] font-semibold tracking-[-0.02em] ${
              ink ? "text-neutral-100" : "text-neutral-950"
            }`}
          >
            {PRODUCT_BRAND.name}
          </span>
          {ink ? (
            <span className="mt-0.5 text-[9px] font-medium uppercase tracking-[0.18em] text-neutral-500">
              Press · CMS
            </span>
          ) : null}
        </span>
      )}
    </span>
  );
}
