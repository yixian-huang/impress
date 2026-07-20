import type { UploadTrayItem } from "@/hooks/useMediaUploadTray";

function formatSize(bytes: number): string {
  if (!bytes) return "";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

const STATUS_BORDER: Record<UploadTrayItem["status"], string> = {
  uploading: "border-slate-200",
  success: "border-emerald-200",
  error: "border-red-200",
};

export function MediaUploadTray({
  items,
  onDismiss,
  onRetry,
  canRetry,
}: {
  items: UploadTrayItem[];
  onDismiss: (id: string) => void;
  onRetry: (id: string) => void;
  canRetry: (id: string) => boolean;
}) {
  if (items.length === 0) return null;

  return (
    <div
      className="pointer-events-none fixed bottom-4 right-4 z-[70] flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-2"
      role="status"
      aria-live="polite"
      aria-relevant="additions text"
    >
      {items.map((item) => {
        const showRetry = item.status === "error" && canRetry(item.id);
        return (
          <div
            key={item.id}
            className={`pointer-events-auto rounded-xl border bg-white shadow-lg overflow-hidden ${STATUS_BORDER[item.status]}`}
          >
            <div className="flex items-start gap-2 px-3 py-2">
              <div className="min-w-0 flex-1">
                <div className="truncate text-xs font-medium text-slate-800" title={item.name}>
                  {item.name}
                  {item.size > 0 && (
                    <span className="ml-1.5 font-normal text-slate-400">{formatSize(item.size)}</span>
                  )}
                </div>
                <div
                  className={`mt-0.5 text-[11px] ${
                    item.status === "error" ? "text-red-600" : "text-slate-500"
                  }`}
                >
                  {item.status === "uploading" && `上传中 ${item.percent}%`}
                  {item.status === "success" && "上传成功"}
                  {item.status === "error" && (item.message || "上传失败")}
                </div>
              </div>
              <div className="flex flex-shrink-0 items-center gap-1">
                {showRetry && (
                  <button
                    type="button"
                    onClick={() => onRetry(item.id)}
                    className="rounded-md border border-red-200 bg-red-50 px-2 py-0.5 text-[11px] text-red-700 hover:bg-red-100"
                  >
                    重试
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => onDismiss(item.id)}
                  className="px-1.5 py-0.5 text-sm leading-none text-slate-400 hover:text-slate-700"
                  aria-label={`关闭 ${item.name}`}
                >
                  ×
                </button>
              </div>
            </div>
            {item.status === "uploading" && (
              <div className="h-1 bg-slate-100" aria-hidden>
                <div
                  className="h-full bg-blue-500 transition-[width] duration-150 ease-out"
                  style={{ width: `${Math.max(4, item.percent)}%` }}
                />
              </div>
            )}
            {item.status === "success" && <div className="h-1 bg-emerald-400" aria-hidden />}
            {item.status === "error" && <div className="h-1 bg-red-400" aria-hidden />}
          </div>
        );
      })}
    </div>
  );
}
