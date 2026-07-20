import { useEffect } from "react";

type SaveIntent = "draft" | "publish" | "autosave";

/**
 * Global editor shortcuts:
 * - ⌘/Ctrl+S save, ⌘⇧S publish, ⌘/Ctrl+P preview
 * - ⌘/Ctrl+F find, ⌘/Ctrl+\ toggle zen, Esc exit zen / close find
 */
export function useEditorShortcuts(opts: {
  canPublish: boolean;
  zenMode: boolean;
  findOpen: boolean;
  onSave: (intent: SaveIntent) => void;
  onPreview: () => void;
  onFind: () => void;
  onToggleZen: () => void;
  onExitOverlay: () => void;
}) {
  const {
    canPublish,
    zenMode,
    findOpen,
    onSave,
    onPreview,
    onFind,
    onToggleZen,
    onExitOverlay,
  } = opts;

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      const mod = e.metaKey || e.ctrlKey;

      if (e.key === "Escape" && (zenMode || findOpen)) {
        // Let FindReplaceBar handle Esc first when focus is inside it;
        // still allow global exit when zen-only.
        if (zenMode && !findOpen) {
          e.preventDefault();
          onExitOverlay();
        } else if (findOpen && !(e.target instanceof HTMLInputElement)) {
          e.preventDefault();
          onExitOverlay();
        }
        return;
      }

      if (!mod) return;

      if (e.key === "s" || e.key === "S") {
        e.preventDefault();
        if (e.shiftKey) {
          if (canPublish) onSave("publish");
        } else {
          onSave("draft");
        }
        return;
      }
      if (e.key === "p" || e.key === "P") {
        e.preventDefault();
        onPreview();
        return;
      }
      if (e.key === "f" || e.key === "F") {
        // Don't steal when user is in a non-editor input with native find expectation
        // — we own the whole article editor page.
        e.preventDefault();
        onFind();
        return;
      }
      if (e.key === "\\" || e.code === "Backslash") {
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
    onSave,
    onPreview,
    onFind,
    onToggleZen,
    onExitOverlay,
  ]);
}
