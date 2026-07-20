import { useState, useEffect } from "react";
import { listUnifiedPageVersions } from "@/api/unifiedPages";
import { AdminButton, AdminLoading } from "@/components/admin/ui";

export function VersionHistoryPanel({
  pageId,
  onClose,
  onRollback,
  canRollback,
}: {
  pageId: number;
  onClose: () => void;
  onRollback: (version: number) => void;
  canRollback: boolean;
}) {
  const [versions, setVersions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    listUnifiedPageVersions(pageId)
      .then((data: any) => {
        setVersions(Array.isArray(data) ? data : data?.items || []);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [pageId]);

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-slate-900/40 backdrop-blur-[1px]">
      <div
        className="absolute inset-0"
        onClick={onClose}
        aria-hidden
      />
      <div className="relative flex h-full w-full max-w-md flex-col border-l border-slate-200/80 bg-white shadow-2xl">
        <div className="flex items-center justify-between border-b border-slate-100 px-5 py-4">
          <h3 className="text-base font-semibold tracking-tight text-slate-900">版本历史</h3>
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg px-2 py-1 text-lg leading-none text-slate-400 transition hover:bg-slate-100 hover:text-slate-700"
            aria-label="关闭"
          >
            ×
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-4">
          {loading ? (
            <AdminLoading />
          ) : versions.length === 0 ? (
            <div className="py-8 text-center text-sm text-slate-500">暂无版本记录</div>
          ) : (
            <div className="space-y-2">
              {versions.map((v: any) => (
                <div
                  key={v.version ?? v.id}
                  className="flex items-center justify-between rounded-xl border border-slate-200 p-3 shadow-sm"
                >
                  <div>
                    <div className="text-sm font-medium text-slate-800">
                      版本 {v.version ?? v.id}
                    </div>
                    <div className="text-xs text-slate-500">
                      {v.createdAt ? new Date(v.createdAt).toLocaleString() : ""}
                    </div>
                  </div>
                  {canRollback && (
                    <AdminButton
                      size="sm"
                      variant="soft"
                      onClick={() => onRollback(v.version ?? v.id)}
                    >
                      回滚
                    </AdminButton>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export function ConflictDialog({
  currentVersion,
  onReload,
  onDismiss,
}: {
  currentVersion: number;
  onReload: () => void;
  onDismiss: () => void;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/45 p-4 backdrop-blur-[2px]">
      <div className="w-full max-w-sm rounded-2xl border border-slate-200/80 bg-white p-6 shadow-[0_24px_64px_rgba(15,23,42,0.18)]">
        <h3 className="mb-2 text-lg font-semibold text-red-700">版本冲突</h3>
        <p className="mb-4 text-sm text-slate-600">
          此页面已被他人编辑，当前服务端版本为 <strong>{currentVersion}</strong>。
          请重新加载后再编辑。
        </p>
        <div className="flex justify-end gap-2">
          <AdminButton variant="secondary" size="sm" onClick={onDismiss}>
            关闭
          </AdminButton>
          <AdminButton size="sm" onClick={onReload}>
            重新加载
          </AdminButton>
        </div>
      </div>
    </div>
  );
}
