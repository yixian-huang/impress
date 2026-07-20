import type { ArticleVersionDetail } from "@/api/articles";
import { FieldDiff } from "./TextDiffView";
import { snapStr } from "../utils/snapshot";
import { formatVersionTime } from "../utils/versionLabels";

export function VersionCompareView({
  left,
  right,
  restoring,
  canRestore,
  onRestoreLeft,
}: {
  left: ArticleVersionDetail;
  right: ArticleVersionDetail;
  restoring: boolean;
  canRestore: boolean;
  onRestoreLeft?: (version: number) => void;
}) {
  const leftSnap = left.snapshot;
  const rightSnap = right.snapshot;
  const rightIsCurrent = right.action === "current";

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-6">
      <div className="flex items-center gap-3 text-sm bg-slate-50 rounded-lg p-3 border border-slate-100">
        <div className="flex-1">
          <div className="text-xs text-slate-400">左</div>
          <div className="font-medium">v{left.version}</div>
          <div className="text-xs text-slate-500">{formatVersionTime(left.createdAt)}</div>
        </div>
        <div className="text-slate-300">→</div>
        <div className="flex-1 text-right">
          <div className="text-xs text-slate-400">右</div>
          <div className="font-medium">
            {rightIsCurrent ? "当前编辑" : `v${right.version}`}
          </div>
          <div className="text-xs text-slate-500">
            {rightIsCurrent ? "含未保存更改" : formatVersionTime(right.createdAt)}
          </div>
        </div>
      </div>

      {canRestore && onRestoreLeft && (
        <button
          type="button"
          disabled={restoring}
          onClick={() => onRestoreLeft(left.version)}
          className="w-full px-3 py-1.5 text-sm border border-amber-300 text-amber-800 rounded-lg hover:bg-amber-50 disabled:opacity-50"
        >
          {rightIsCurrent
            ? `用 v${left.version} 覆盖当前编辑`
            : `恢复左侧版本 v${left.version} 到编辑器`}
        </button>
      )}

      <FieldDiff label="中文标题" left={snapStr(leftSnap, "zhTitle")} right={snapStr(rightSnap, "zhTitle")} />
      <FieldDiff label="英文标题" left={snapStr(leftSnap, "enTitle")} right={snapStr(rightSnap, "enTitle")} />
      <FieldDiff label="Slug" left={snapStr(leftSnap, "slug")} right={snapStr(rightSnap, "slug")} />
      <FieldDiff label="状态" left={snapStr(leftSnap, "status")} right={snapStr(rightSnap, "status")} />
      <FieldDiff
        label="中文正文"
        left={snapStr(leftSnap, "zhBody")}
        right={snapStr(rightSnap, "zhBody")}
        asHtml
      />
      <FieldDiff
        label="英文正文"
        left={snapStr(leftSnap, "enBody")}
        right={snapStr(rightSnap, "enBody")}
        asHtml
      />
    </div>
  );
}
