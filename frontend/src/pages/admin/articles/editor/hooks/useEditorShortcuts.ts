import { useEffect } from "react";

type SaveIntent = "draft" | "publish" | "autosave";

/**
 * Global ⌘/Ctrl+S (save), ⌘⇧S (publish), ⌘/Ctrl+P (preview).
 */
export function useEditorShortcuts(opts: {
  canPublish: boolean;
  onSave: (intent: SaveIntent) => void;
  onPreview: () => void;
}) {
  const { canPublish, onSave, onPreview } = opts;

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (!(e.metaKey || e.ctrlKey)) return;
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
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [canPublish, onSave, onPreview]);
}
