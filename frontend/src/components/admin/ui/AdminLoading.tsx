export interface AdminLoadingProps {
  label?: string;
  className?: string;
}

export default function AdminLoading({ label = "加载中…", className = "" }: AdminLoadingProps) {
  return (
    <div className={`flex flex-col items-center justify-center py-16 text-slate-500 ${className}`}>
      <div
        className="h-9 w-9 animate-spin rounded-full border-2 border-slate-200 border-t-blue-600"
        aria-hidden
      />
      <p className="mt-3.5 text-sm font-medium text-slate-500">{label}</p>
    </div>
  );
}

export function AdminRouteFallback() {
  return (
    <div className="animate-pulse space-y-4 p-1" aria-busy="true" aria-label="页面加载中">
      <div className="h-8 w-48 rounded-xl bg-slate-200/80" />
      <div className="h-4 w-72 rounded-lg bg-slate-200/60" />
      <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="h-24 rounded-2xl bg-slate-200/70" />
        <div className="h-24 rounded-2xl bg-slate-200/60" />
        <div className="h-24 rounded-2xl bg-slate-200/50" />
        <div className="h-24 rounded-2xl bg-slate-200/40" />
      </div>
      <div className="h-52 rounded-2xl bg-slate-200/50" />
    </div>
  );
}
