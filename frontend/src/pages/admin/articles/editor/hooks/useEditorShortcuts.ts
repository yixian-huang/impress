import { useEffect } from "react";

type SaveIntent = "draft" | "publish" | "autosave";

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
      const overlayOpen = zenMode || findOpen || shortcutHelpOpen;

      if (e.key === "Escape" && overlayOpen) {
        if (shortcutHelpOpen || zenMode || (findOpen && !(e.target instanceof HTMLInputElement))) {
          e.preventDefault();
          onExitOverlay();
        }
        return;
      }

      if (!mod) return;

      // ⌘/ — help (slash key; some layouts use e.key === "/" with shift on non-US)
      if (e.key === "/" || e.code === "Slash") {
        e.preventDefault();
        onToggleShortcutHelp();
        return;
      }

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
    shortcutHelpOpen,
    onSave,
    onPreview,
    onFind,
    onToggleZen,
    onToggleShortcutHelp,
    onExitOverlay,
  ]);
}
