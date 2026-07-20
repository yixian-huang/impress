import type { EditorSavePhase } from "./saveStatusUtils";

/** Save / dirty status chip shown in the article editor action bar. */
export function SaveStatusBadge({
  phase,
  lastSavedAt,
  isAutosave,
}: {
  phase: EditorSavePhase;
  lastSavedAt: Date | null;
  isAutosave?: boolean;
}) {
  const time =
    lastSavedAt &&
    lastSavedAt.toLocaleTimeString("zh-CN", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });

  if (phase === "saving") {
    return (
      <span className="text-xs text-blue-600 whitespace-nowrap" title="正在保存">
        保存中…
      </span>
    );
  }
  if (phase === "dirty") {
    return (
      <span className="text-xs text-amber-600 whitespace-nowrap" title="有未保存的更改">
        未保存
      </span>
    );
  }
  if (phase === "saved" || phase === "clean") {
    if (!time) {
      return (
        <span className="text-xs text-slate-400 whitespace-nowrap">已保存</span>
      );
    }
    return (
      <span
        className="text-xs text-slate-500 whitespace-nowrap"
        title={isAutosave ? "自动保存" : "手动保存"}
      >
        {isAutosave ? "自动保存" : "已保存"} · {time}
      </span>
    );
  }
  if (phase === "error") {
    return (
      <span className="text-xs text-red-600 whitespace-nowrap">保存失败</span>
    );
  }
  return null;
}
