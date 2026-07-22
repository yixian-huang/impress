import type { ChecklistItem } from "../utils/publishChecklist";

export function PublishChecklistDialog({
  open,
  items,
  busy,
  onCancel,
  onForcePublish,
  onAIFill,
}: {
  open: boolean;
  items: ChecklistItem[];
  busy?: boolean;
  onCancel: () => void;
  /** Only for warn-only lists — blocks cannot force */
  onForcePublish?: () => void;
  /** Open AI meta fill for missing SEO / title fields */
  onAIFill?: () => void;
}) {
  if (!open) return null;

  const blocks = items.filter((i) => i.severity === "block");
  const canForce = blocks.length === 0 && !!onForcePublish;
  const aiHelpful = items.some((i) =>
    ["zh-title", "zh-meta", "zh-meta-short", "en-title-missing", "slug"].includes(i.id),
  );

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="publish-checklist-title"
        className="w-full max-w-md rounded-xl bg-white shadow-xl border border-slate-200 overflow-hidden"
      >
        <div className="px-4 py-3 border-b border-slate-100">
          <h2 id="publish-checklist-title" className="text-sm font-semibold text-slate-900">
            发布前检查
          </h2>
          <p className="text-xs text-slate-500 mt-0.5">
            {blocks.length > 0
              ? "存在必须修复的问题，请修改后再发布。"
              : "以下项建议完善，也可忽略并继续发布。"}
          </p>
        </div>

        <ul className="max-h-72 overflow-y-auto px-4 py-3 space-y-2">
          {items.map((item) => (
            <li
              key={item.id}
              className={`rounded-lg border px-3 py-2 text-sm ${
                item.severity === "block"
                  ? "border-red-200 bg-red-50 text-red-900"
                  : "border-amber-200 bg-amber-50 text-amber-950"
              }`}
            >
              <div className="flex items-start gap-2">
                <span className="text-[10px] font-semibold uppercase tracking-wide mt-0.5 flex-shrink-0">
                  {item.severity === "block" ? "必改" : "建议"}
                </span>
                <div className="min-w-0">
                  <div className="font-medium">{item.message}</div>
                  {item.hint && (
                    <div className="text-xs opacity-80 mt-0.5">{item.hint}</div>
                  )}
                </div>
              </div>
            </li>
          ))}
        </ul>

        <div className="flex items-center justify-end gap-2 px-4 py-3 border-t border-slate-100 bg-slate-50">
          <button
            type="button"
            onClick={onCancel}
            disabled={busy}
            className="px-3 py-1.5 text-sm border border-slate-200 rounded-lg hover:bg-white disabled:opacity-50"
          >
            返回修改
          </button>
          {onAIFill && aiHelpful && (
            <button
              type="button"
              onClick={onAIFill}
              disabled={busy}
              className="px-3 py-1.5 text-sm border border-violet-200 bg-violet-50 text-violet-800 rounded-lg hover:bg-violet-100 disabled:opacity-50"
            >
              用 AI 补齐
            </button>
          )}
          {canForce && (
            <button
              type="button"
              onClick={onForcePublish}
              disabled={busy}
              className="px-3 py-1.5 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
            >
              {busy ? "发布中..." : "仍要发布"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
