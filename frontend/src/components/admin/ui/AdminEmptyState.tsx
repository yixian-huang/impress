import type { ReactNode } from "react";
import { Inbox } from "lucide-react";

export interface AdminEmptyStateProps {
  title: string;
  description?: string;
  action?: ReactNode;
  icon?: ReactNode;
  className?: string;
}

export default function AdminEmptyState({
  title,
  description,
  action,
  icon,
  className = "",
}: AdminEmptyStateProps) {
  return (
    <div
      className={`flex flex-col items-center justify-center rounded-2xl border border-dashed border-slate-200/90 bg-gradient-to-b from-slate-50/80 to-white px-6 py-14 text-center ${className}`}
    >
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-2xl bg-slate-100 text-slate-400">
        {icon ?? <Inbox className="h-6 w-6" aria-hidden />}
      </div>
      <p className="text-sm font-semibold text-slate-800">{title}</p>
      {description ? (
        <p className="mt-1.5 max-w-sm text-sm leading-relaxed text-slate-500">{description}</p>
      ) : null}
      {action ? <div className="mt-5">{action}</div> : null}
    </div>
  );
}
