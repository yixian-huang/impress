import { useCallback, useRef, useState } from "react";
import type { EditorSavePhase } from "../saveStatusUtils";

/**
 * Tracks unsaved edits without thrashing on programmatic hydrate/setContent.
 */
export function useDirtyState(initialReady: boolean) {
  const [isDirty, setIsDirty] = useState(false);
  const [savePhase, setSavePhase] = useState<EditorSavePhase>("clean");
  const [lastSavedAt, setLastSavedAt] = useState<Date | null>(null);
  const [lastSaveWasAutosave, setLastSaveWasAutosave] = useState(false);
  const readyRef = useRef(initialReady);
  const savingRef = useRef(false);

  const touch = useCallback(() => {
    if (!readyRef.current) return;
    setIsDirty(true);
    setSavePhase((p) => (p === "saving" ? p : "dirty"));
  }, []);

  const track = useCallback(
    <T,>(setter: (v: T) => void) =>
      (v: T) => {
        setter(v);
        touch();
      },
    [touch],
  );

  const markClean = useCallback((opts?: { autosave?: boolean }) => {
    setIsDirty(false);
    setSavePhase("saved");
    setLastSavedAt(new Date());
    setLastSaveWasAutosave(!!opts?.autosave);
  }, []);

  const markSaving = useCallback(() => {
    setSavePhase("saving");
  }, []);

  const markError = useCallback(() => {
    setSavePhase("error");
  }, []);

  const markHydrated = useCallback(() => {
    setIsDirty(false);
    setSavePhase("clean");
    setLastSavedAt(null);
    setLastSaveWasAutosave(false);
  }, []);

  /** Double-rAF so TipTap setContent settles before dirty tracking resumes. */
  const resumeReady = useCallback(() => {
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        readyRef.current = true;
      });
    });
  }, []);

  const pauseReady = useCallback(() => {
    readyRef.current = false;
  }, []);

  return {
    isDirty,
    setIsDirty,
    savePhase,
    setSavePhase,
    lastSavedAt,
    lastSaveWasAutosave,
    readyRef,
    savingRef,
    touch,
    track,
    markClean,
    markSaving,
    markError,
    markHydrated,
    resumeReady,
    pauseReady,
  };
}
