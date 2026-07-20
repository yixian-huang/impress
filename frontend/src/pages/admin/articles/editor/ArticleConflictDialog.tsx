export default function ArticleConflictDialog({
  serverUpdatedAt,
  onReload,
  onForceOverwrite,
  onDismiss,
  busy,
}: {
  serverUpdatedAt?: string | null;
  onReload: () => void;
  onForceOverwrite: () => void;
  onDismiss: () => void;
  busy?: boolean;
}) {
  const timeLabel = serverUpdatedAt
    ? (() => {
        try {
          return new Date(serverUpdatedAt).toLocaleString("zh-CN");
        } catch {
          return serverUpdatedAt;
        }
      })()
    : null;

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/40 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-md mx-4 p-6">
        <h3 className="text-lg font-semibold text-red-700 mb-2">保存冲突</h3>
        <p className="text-sm text-gray-600 mb-2">
          这篇文章在服务器上已被其他人（或另一个标签页）修改。继续保存可能覆盖对方的更改。
        </p>
        {timeLabel && (
          <p className="text-xs text-gray-500 mb-4">
            服务器最新更新时间：<strong>{timeLabel}</strong>
          </p>
        )}
        <div className="flex flex-col sm:flex-row sm:justify-end gap-2">
          <button
            type="button"
            onClick={onDismiss}
            disabled={busy}
            className="px-4 py-2 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
          >
            继续编辑
          </button>
          <button
            type="button"
            onClick={onForceOverwrite}
            disabled={busy}
            className="px-4 py-2 text-sm border border-amber-400 text-amber-900 rounded-lg hover:bg-amber-50 disabled:opacity-50"
          >
            强制覆盖
          </button>
          <button
            type="button"
            onClick={onReload}
            disabled={busy}
            className="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            重新加载
          </button>
        </div>
      </div>
    </div>
  );
}
