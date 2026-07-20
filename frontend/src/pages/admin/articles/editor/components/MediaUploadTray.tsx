import type { UploadTrayItem } from "../hooks/useMediaUploadTray";

function formatSize(bytes: number): string {
  if (!bytes) return "";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

export function MediaUploadTray({
  items,
  onDismiss,
  onRetry,
}: {
  items: UploadTrayItem[];
  onDismiss: (id: string) => void;
  onRetry: (id: string) => void;
}) {
  if (items.length === 0) return null;

  return (
    <div
      className="pointer-events-none fixed bottom-4 right-4 z-[70] flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-2"
      role="status"
      aria-live="polite"
    >
      {items.map((item) => (
        <div
          key={item.id}
          className={`pointer-events-auto rounded-xl border shadow-lg bg-white overflow-hidden ${
            item.status === "error"
              ? "border-red-200"
              : item.status === "success"
                ? "border-emerald-200"
                : "border-slate-200"
          }`}
        >
          <div className="flex items-start gap-2 px-3 py-2">
            <div className="min-w-0 flex-1">
              <div className="text-xs font-medium text-slate-800 truncate" title={item.name}>
                {item.name}
                {item.size > 0 && (
                  <span className="ml-1.5 font-normal text-slate-400">{formatSize(item.size)}</span>
                )}
              </div>
              <div className="mt-0.5 text-[11px] text-slate-500">
                {item.status === "uploading" && `上传中 ${item.percent}%`}
                {item.status === "success" && "上传成功"}
                {item.status === "error" && (item.message || "上传失败")}
              </div>
            </div>
            <div className="flex items-center gap-1 flex-shrink-0">
              {item.status === "error" && item.retry && (
                <button
                  type="button"
                  onClick={() => onRetry(item.id)}
                  className="px-2 py-0.5 text-[11px] rounded-md bg-red-50 text-red-700 border border-red-200 hover:bg-red-100"
                >
                  重试
                </button>
              )}
              <button
                type="button"
                onClick={() => onDismiss(item.id)}
                className="px-1.5 py-0.5 text-slate-400 hover:text-slate-700 text-sm leading-none"
                aria-label="关闭"
              >
                ×
              </button>
            </div>
          </div>
          {item.status === "uploading" && (
            <div className="h-1 bg-slate-100">
              <div
                className="h-full bg-blue-500 transition-[width] duration-150 ease-out"
                style={{ width: `${Math.max(4, item.percent)}%` }}
              />
            </div>
          )}
          {item.status === "success" && <div className="h-1 bg-emerald-400" />}
          {item.status === "error" && <div className="h-1 bg-red-400" />}
        </div>
      ))}
    </div>
  );
}
