import { forwardRef, type ButtonHTMLAttributes, type ReactNode } from "react";
import { adminTheme } from "./adminTheme";

type Variant = "primary" | "secondary" | "danger" | "ghost" | "soft";
type Size = "sm" | "md" | "lg";

export interface AdminButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  children: ReactNode;
}

const variantClass: Record<Variant, string> = {
  primary:
    "bg-neutral-950 text-white border border-transparent shadow-[0_1px_2px_rgba(0,0,0,0.15)] hover:bg-neutral-800 active:bg-black",
  secondary:
    "bg-white text-neutral-700 border border-neutral-200 shadow-[0_1px_0_rgba(0,0,0,0.03)] hover:bg-neutral-50 hover:border-neutral-300 active:bg-neutral-100",
  danger:
    "bg-neutral-900 text-white border border-transparent shadow-sm hover:bg-black active:bg-neutral-950",
  ghost:
    "bg-transparent text-neutral-600 border border-transparent hover:bg-neutral-100 hover:text-neutral-950",
  soft: "bg-neutral-100 text-neutral-900 border border-neutral-200 hover:bg-neutral-200/80 active:bg-neutral-200",
};

const sizeClass: Record<Size, string> = {
  sm: "h-8 px-3 text-xs rounded-lg tracking-wide",
  md: "h-9 px-3.5 text-sm rounded-lg tracking-wide",
  lg: "h-11 px-5 text-sm rounded-lg tracking-wide",
};

const AdminButton = forwardRef<HTMLButtonElement, AdminButtonProps>(function AdminButton(
  { variant = "primary", size = "md", className = "", disabled, children, type = "button", ...rest },
  ref,
) {
  return (
    <button
      ref={ref}
      type={type}
      disabled={disabled}
      className={`inline-flex items-center justify-center gap-1.5 font-medium ${adminTheme.transition} ${adminTheme.focusRing} ${variantClass[variant]} ${sizeClass[size]} ${
        disabled ? "opacity-50 cursor-not-allowed shadow-none hover:shadow-none" : ""
      } ${className}`}
      {...rest}
    >
      {children}
    </button>
  );
});

export default AdminButton;
