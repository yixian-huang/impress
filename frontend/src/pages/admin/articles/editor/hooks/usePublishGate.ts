import { useCallback, useState } from "react";
import {
  evaluatePublishChecklist,
  type ChecklistItem,
  type PublishChecklistInput,
} from "../utils/publishChecklist";

/**
 * Publish-time readiness gate: collect issues, show dialog, force only for warns.
 */
export function usePublishGate(opts: {
  canPublish: boolean;
  collect: () => PublishChecklistInput;
  onPublish: () => void;
}) {
  const { canPublish, collect, onPublish } = opts;
  const [items, setItems] = useState<ChecklistItem[] | null>(null);

  const requestPublish = useCallback((opts?: { force?: boolean }) => {
    if (!canPublish) return;
    const next = evaluatePublishChecklist(collect());
    if (next.length > 0 && !opts?.force) {
      setItems(next);
      return;
    }
    if (opts?.force && next.some((i) => i.severity === "block")) {
      setItems(next);
      return;
    }
    setItems(null);
    onPublish();
  }, [canPublish, collect, onPublish]);

  const dismiss = useCallback(() => setItems(null), []);

  const hasBlocks = !!items?.some((i) => i.severity === "block");

  return {
    items,
    open: !!items && items.length > 0,
    hasBlocks,
    requestPublish,
    dismiss,
  };
}
