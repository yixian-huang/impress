import type { ReactNode } from "react";
import { adminTheme } from "./adminTheme";

export interface AdminStatCardProps {
  label: string;
  value: string | number;
  icon: ReactNode;
  colorClass?: string;
  loading?: boolean;
}

export default function AdminStatCard({
  label,
  value,
  icon,
  colorClass = "bg-blue-600",
  loading = false,
}: AdminStatCardProps) {
  return (
    <div className={`${adminTheme.card} overflow-hidden`}>
      {loading ? (
        <div className="p-5">
          <div className="flex animate-pulse items-center gap-4">
            <div className="h-12 w-12 rounded-2xl bg-slate-100" />
            <div className="flex-1">
              <div className="mb-2 h-3 w-16 rounded bg-slate-100" />
              <div className="h-7 w-12 rounded bg-slate-100" />
            </div>
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-4 p-5">
          <div
            className={`${colorClass} shrink-0 rounded-2xl p-3 text-white shadow-sm shadow-slate-900/10`}
          >
            {icon}
          </div>
          <div className="min-w-0">
            <p className="truncate text-sm text-slate-500">{label}</p>
            <p className="text-2xl font-semibold tracking-tight text-slate-900 tabular-nums">
              {value}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
