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
    "bg-blue-600 text-white border border-transparent shadow-sm shadow-blue-600/20 hover:bg-blue-700 hover:shadow-md hover:shadow-blue-600/25 active:bg-blue-800",
  secondary:
    "bg-white text-slate-700 border border-slate-200/90 shadow-sm hover:bg-slate-50 hover:border-slate-300 active:bg-slate-100",
  danger:
    "bg-red-600 text-white border border-transparent shadow-sm shadow-red-600/15 hover:bg-red-700 active:bg-red-800",
  ghost: "bg-transparent text-slate-600 border border-transparent hover:bg-slate-100/80 hover:text-slate-900",
  soft: "bg-blue-50 text-blue-700 border border-blue-200/70 hover:bg-blue-100/80 active:bg-blue-100",
};

const sizeClass: Record<Size, string> = {
  sm: "h-8 px-3 text-xs rounded-lg",
  md: "h-9 px-3.5 text-sm rounded-xl",
  lg: "h-11 px-5 text-sm rounded-xl",
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
