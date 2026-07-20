import { useEffect } from "react";

type SaveIntent = "draft" | "publish" | "autosave";

function isEditableTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false;
  const tag = target.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return true;
  return target.isContentEditable;
}

/**
 * Global editor shortcuts:
 * - ⌘/Ctrl+S save, ⌘⇧S publish, ⌘/Ctrl+P preview
 * - ⌘/Ctrl+F find, ⌘/Ctrl+\ toggle zen, ⌘/ toggle help, Esc exit overlays
 */
export function useEditorShortcuts(opts: {
  canPublish: boolean;
  zenMode: boolean;
  findOpen: boolean;
  shortcutHelpOpen: boolean;
  onSave: (intent: SaveIntent) => void;
  onPreview: () => void;
  onFind: () => void;
  onToggleZen: () => void;
  onToggleShortcutHelp: () => void;
  onExitOverlay: () => void;
}) {
  const {
    canPublish,
    zenMode,
    findOpen,
    shortcutHelpOpen,
    onSave,
    onPreview,
    onFind,
    onToggleZen,
    onToggleShortcutHelp,
    onExitOverlay,
  } = opts;

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      const mod = e.metaKey || e.ctrlKey;

      // Esc: close help → find → zen (handled by onExitOverlay order in page)
      if (e.key === "Escape" && (shortcutHelpOpen || findOpen || zenMode)) {
        // Always allow Esc to leave help/zen; for find, skip when typing in inputs
        // outside the find bar so native clear still works in title fields.
        if (shortcutHelpOpen || zenMode) {
          e.preventDefault();
          onExitOverlay();
          return;
        }
        if (findOpen && !isEditableTarget(e.target)) {
          e.preventDefault();
          onExitOverlay();
        }
        return;
      }

      if (!mod || e.altKey) return;

      // Prefer physical Slash key so layouts that need Shift for "/" still work
      if (e.code === "Slash" || e.key === "/") {
        e.preventDefault();
        onToggleShortcutHelp();
        return;
      }

      const key = e.key.toLowerCase();

      if (key === "s") {
        e.preventDefault();
        if (e.shiftKey) {
          if (canPublish) onSave("publish");
        } else {
          onSave("draft");
        }
        return;
      }
      if (key === "p" && !e.shiftKey) {
        e.preventDefault();
        onPreview();
        return;
      }
      if (key === "f" && !e.shiftKey) {
        e.preventDefault();
        onFind();
        return;
      }
      if (e.code === "Backslash" || e.key === "\\") {
        e.preventDefault();
        onToggleZen();
      }
    };

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [
    canPublish,
    zenMode,
    findOpen,
    shortcutHelpOpen,
    onSave,
    onPreview,
    onFind,
    onToggleZen,
    onToggleShortcutHelp,
    onExitOverlay,
  ]);
}
