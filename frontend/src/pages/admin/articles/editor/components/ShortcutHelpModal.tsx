import { useMemo, useRef } from "react";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import {
  EDITOR_SHORTCUTS,
  SHORTCUT_GROUP_LABELS,
  detectApplePlatform,
  formatShortcutKeys,
  type ShortcutDef,
} from "../utils/editorShortcuts";

const GROUP_ORDER = ["file", "view", "edit"] as const;

function KeyCap({ children }: { children: string }) {
  return (
    <kbd className="inline-flex min-w-[1.4rem] items-center justify-center rounded border border-slate-300 bg-slate-50 px-1.5 py-0.5 text-[11px] font-medium text-slate-700 shadow-sm">
      {children}
    </kbd>
  );
}

function ShortcutRow({ item, isApple }: { item: ShortcutDef; isApple: boolean }) {
  const keys = formatShortcutKeys(item.keys, isApple);
  return (
    <li className="flex items-center justify-between gap-3 py-1 text-sm">
      <span className="text-slate-700">{item.label}</span>
      <span className="flex flex-shrink-0 items-center gap-0.5">
        {keys.map((k, i) => (
          <span key={`${item.id}-${i}`} className="flex items-center gap-0.5">
            {i > 0 && <span className="text-xs text-slate-300">+</span>}
            <KeyCap>{k}</KeyCap>
          </span>
        ))}
      </span>
    </li>
  );
}

export function ShortcutHelpModal({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const panelRef = useRef<HTMLDivElement>(null);
  const closeRef = useRef<HTMLButtonElement>(null);
  const isApple = useMemo(() => detectApplePlatform(), []);

  useFocusTrap(open, panelRef, closeRef);

  const groups = useMemo(
    () =>
      GROUP_ORDER.map((group) => ({
        group,
        label: SHORTCUT_GROUP_LABELS[group],
        items: EDITOR_SHORTCUTS.filter((s) => s.group === group),
      })).filter((g) => g.items.length > 0),
    [],
  );

  if (!open) return null;

  const helpKeys = formatShortcutKeys(["⌘", "/"], isApple);

  return (
    <div
      className="fixed inset-0 z-[80] flex items-center justify-center bg-black/40 p-4"
      onClick={onClose}
      role="presentation"
    >
      <div
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="shortcut-help-title"
        className="w-full max-w-md overflow-hidden rounded-xl border border-slate-200 bg-white shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3">
          <h2 id="shortcut-help-title" className="text-sm font-semibold text-slate-900">
            键盘快捷键
          </h2>
          <button
            ref={closeRef}
            type="button"
            onClick={onClose}
            className="px-1 text-lg leading-none text-slate-400 hover:text-slate-700"
            aria-label="关闭"
          >
            ×
          </button>
        </div>

        <div className="max-h-[70vh] space-y-4 overflow-y-auto px-4 py-3">
          {groups.map(({ group, label, items }) => (
            <section key={group}>
              <h3 className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-400">
                {label}
              </h3>
              <ul className="space-y-1.5">
                {items.map((item) => (
                  <ShortcutRow key={item.id} item={item} isApple={isApple} />
                ))}
              </ul>
            </section>
          ))}
          <p className="pt-1 text-[11px] text-slate-400">
            在编辑器内按{" "}
            {helpKeys.map((k, i) => (
              <span key={k}>
                {i > 0 && " + "}
                <KeyCap>{k}</KeyCap>
              </span>
            ))}{" "}
            可随时打开本面板。
          </p>
        </div>

        <div className="flex justify-end border-t border-slate-100 bg-slate-50 px-4 py-3">
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg border border-slate-200 px-3 py-1.5 text-sm hover:bg-white"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  );
}
