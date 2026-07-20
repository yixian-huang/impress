import type { ReactNode } from "react";

type Tone = "success" | "warning" | "neutral" | "info" | "danger";

export interface AdminBadgeProps {
  children: ReactNode;
  tone?: Tone;
  className?: string;
}

/** Monochrome print tones — density instead of hue. */
const toneClass: Record<Tone, string> = {
  success: "bg-neutral-100 text-neutral-800 ring-neutral-300/80",
  warning: "bg-neutral-200/80 text-neutral-800 ring-neutral-300",
  neutral: "bg-neutral-100 text-neutral-600 ring-neutral-200",
  info: "bg-neutral-50 text-neutral-700 ring-neutral-200",
  danger: "bg-neutral-900 text-white ring-neutral-900",
};

export default function AdminBadge({ children, tone = "neutral", className = "" }: AdminBadgeProps) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium tracking-wide ring-1 ring-inset ${toneClass[tone]} ${className}`}
    >
      {children}
    </span>
  );
}
