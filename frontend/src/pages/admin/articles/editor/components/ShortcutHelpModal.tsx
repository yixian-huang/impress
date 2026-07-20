import {
  EDITOR_SHORTCUTS,
  SHORTCUT_GROUP_LABELS,
  detectApplePlatform,
  formatShortcutKeys,
  type ShortcutDef,
} from "../utils/editorShortcuts";

function KeyCap({ children }: { children: string }) {
  return (
    <kbd className="inline-flex min-w-[1.4rem] items-center justify-center rounded border border-slate-300 bg-slate-50 px-1.5 py-0.5 text-[11px] font-medium text-slate-700 shadow-sm">
      {children}
    </kbd>
  );
}

export function ShortcutHelpModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  if (!open) return null;

  const isApple = detectApplePlatform();
  const groups = (["file", "view", "edit"] as const).map((group) => ({
    group,
    label: SHORTCUT_GROUP_LABELS[group],
    items: EDITOR_SHORTCUTS.filter((s) => s.group === group),
  }));

  return (
    <div
      className="fixed inset-0 z-[80] flex items-center justify-center bg-black/40 p-4"
      onClick={onClose}
      role="presentation"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="shortcut-help-title"
        className="w-full max-w-md rounded-xl bg-white shadow-xl border border-slate-200 overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-100">
          <h2 id="shortcut-help-title" className="text-sm font-semibold text-slate-900">
            键盘快捷键
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-slate-700 text-lg leading-none px-1"
            aria-label="关闭"
          >
            ×
          </button>
        </div>

        <div className="max-h-[70vh] overflow-y-auto px-4 py-3 space-y-4">
          {groups.map(({ group, label, items }) =>
            items.length === 0 ? null : (
              <section key={group}>
                <h3 className="text-[11px] font-semibold uppercase tracking-wide text-slate-400 mb-2">
                  {label}
                </h3>
                <ul className="space-y-1.5">
                  {items.map((item: ShortcutDef) => (
                    <li
                      key={item.id}
                      className="flex items-center justify-between gap-3 text-sm py-1"
                    >
                      <span className="text-slate-700">{item.label}</span>
                      <span className="flex items-center gap-0.5 flex-shrink-0">
                        {formatShortcutKeys(item.keys, isApple).map((k, i) => (
                          <span key={`${item.id}-${i}`} className="flex items-center gap-0.5">
                            {i > 0 && <span className="text-slate-300 text-xs">+</span>}
                            <KeyCap>{k}</KeyCap>
                          </span>
                        ))}
                      </span>
                    </li>
                  ))}
                </ul>
              </section>
            ),
          )}
          <p className="text-[11px] text-slate-400 pt-1">
            在编辑器内按{" "}
            {formatShortcutKeys(["⌘", "/"], isApple).map((k, i) => (
              <span key={k}>
                {i > 0 && " + "}
                <KeyCap>{k}</KeyCap>
              </span>
            ))}{" "}
            可随时打开本面板。
          </p>
        </div>

        <div className="px-4 py-3 border-t border-slate-100 bg-slate-50 flex justify-end">
          <button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm border border-slate-200 rounded-lg hover:bg-white"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  );
}
