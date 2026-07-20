import type { ArticleVersionListItem } from "@/api/articles";
import type { ArticleDraftSnapshot } from "../VersionHistoryPanel";
import { formatVersionTime, versionActionLabel } from "../utils/versionLabels";

export function VersionHistoryList({
  versions,
  loading,
  currentDraft,
  leftVer,
  rightVer,
  comparing,
  restoring,
  canRestore,
  onLeftChange,
  onRightChange,
  onCompare,
  onRestore,
  onQuickCompare,
}: {
  versions: ArticleVersionListItem[];
  loading: boolean;
  currentDraft?: ArticleDraftSnapshot | null;
  leftVer: number | null;
  rightVer: number | "current" | null;
  comparing: boolean;
  restoring: boolean;
  canRestore: boolean;
  onLeftChange: (v: number) => void;
  onRightChange: (v: number | "current") => void;
  onCompare: () => void;
  onRestore?: (version: number) => void;
  onQuickCompare: (version: number) => void;
}) {
  return (
    <>
      <div className="px-4 py-3 border-b border-gray-100 space-y-2 flex-shrink-0 bg-gray-50">
        <p className="text-xs text-gray-500">
          选择两个版本比对；右侧可选「当前编辑」对比未保存内容。
        </p>
        <div className="flex items-center gap-2">
          <select
            value={leftVer ?? ""}
            onChange={(e) => onLeftChange(Number(e.target.value))}
            className="flex-1 text-sm border border-gray-300 rounded-lg px-2 py-1.5"
            disabled={versions.length === 0}
          >
            {versions.map((v) => (
              <option key={`L-${v.version}`} value={v.version}>
                v{v.version} · {versionActionLabel(v.action)} · {formatVersionTime(v.createdAt)}
              </option>
            ))}
          </select>
          <span className="text-xs text-gray-400">vs</span>
          <select
            value={rightVer === "current" ? "current" : (rightVer ?? "")}
            onChange={(e) => {
              const v = e.target.value;
              onRightChange(v === "current" ? "current" : Number(v));
            }}
            className="flex-1 text-sm border border-gray-300 rounded-lg px-2 py-1.5"
            disabled={versions.length === 0 && !currentDraft}
          >
            {currentDraft && (
              <option value="current">当前编辑（含未保存）</option>
            )}
            {versions.map((v) => (
              <option key={`R-${v.version}`} value={v.version}>
                v{v.version} · {versionActionLabel(v.action)} · {formatVersionTime(v.createdAt)}
              </option>
            ))}
          </select>
        </div>
        <button
          type="button"
          onClick={onCompare}
          disabled={
            comparing
            || leftVer == null
            || rightVer == null
            || (versions.length === 0 && rightVer !== "current")
          }
          className="w-full px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
        >
          {comparing ? "比对中…" : "开始比对"}
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {loading ? (
          <div className="text-center text-gray-500 py-8">加载中…</div>
        ) : versions.length === 0 ? (
          <div className="text-center text-gray-500 py-8 text-sm">
            暂无版本记录
            <p className="mt-1 text-xs text-gray-400">保存或发布文章后会自动生成版本快照。</p>
          </div>
        ) : (
          <div className="space-y-2">
            {versions.map((v) => (
              <div
                key={v.id}
                className="p-3 border border-gray-200 rounded-lg hover:border-blue-200 transition-colors"
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="text-sm font-medium text-gray-800">
                    版本 {v.version}
                    <span className="ml-2 text-[10px] px-1.5 py-0.5 rounded bg-gray-100 text-gray-600">
                      {versionActionLabel(v.action)}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    {canRestore && onRestore && (
                      <button
                        type="button"
                        disabled={restoring}
                        className="text-xs text-amber-700 hover:text-amber-900 disabled:opacity-50"
                        onClick={() => onRestore(v.version)}
                      >
                        恢复
                      </button>
                    )}
                    <button
                      type="button"
                      className="text-xs text-blue-600 hover:text-blue-800"
                      onClick={() => onQuickCompare(v.version)}
                    >
                      {currentDraft ? "与当前比对" : "查看"}
                    </button>
                  </div>
                </div>
                <div className="text-xs text-gray-500 mt-1">{formatVersionTime(v.createdAt)}</div>
                {(v.zhTitle || v.summary) && (
                  <div className="text-xs text-gray-600 mt-1 truncate">
                    {v.zhTitle || v.summary}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
}
