import { AlertCircle, X } from "lucide-react";

export interface AdminErrorBannerProps {
  message: string;
  className?: string;
  onDismiss?: () => void;
}

export default function AdminErrorBanner({
  message,
  className = "",
  onDismiss,
}: AdminErrorBannerProps) {
  return (
    <div
      role="alert"
      className={`mb-4 flex items-start gap-3 rounded-2xl border border-red-200/80 bg-red-50/90 px-4 py-3 text-sm text-red-800 shadow-sm shadow-red-900/5 ${className}`}
    >
      <AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-red-500" aria-hidden />
      <span className="min-w-0 flex-1 break-words leading-relaxed">{message}</span>
      {onDismiss ? (
        <button
          type="button"
          onClick={onDismiss}
          className="shrink-0 rounded-lg p-1 text-red-500 transition hover:bg-red-100 hover:text-red-800"
          aria-label="关闭"
        >
          <X className="h-4 w-4" />
        </button>
      ) : null}
    </div>
  );
}
