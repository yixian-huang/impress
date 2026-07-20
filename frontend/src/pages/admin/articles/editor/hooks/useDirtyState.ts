import { useCallback, useMemo, useRef, useState } from "react";
import type { EditorSavePhase } from "../saveStatusUtils";

/**
 * Tracks unsaved edits without thrashing on programmatic hydrate/setContent.
 * Return object identity is stable when values/callbacks are unchanged.
 */
export function useDirtyState(initialReady: boolean) {
  const [isDirty, setIsDirty] = useState(false);
  const [savePhase, setSavePhase] = useState<EditorSavePhase>("clean");
  const [lastSavedAt, setLastSavedAt] = useState<Date | null>(null);
  const [lastSaveWasAutosave, setLastSaveWasAutosave] = useState(false);
  const readyRef = useRef(initialReady);
  const savingRef = useRef(false);
  /** Cache track wrappers so dirty.track(setX) is referentially stable. */
  const trackedCache = useRef(new WeakMap<(v: never) => void, (v: never) => void>());

  const touch = useCallback(() => {
    if (!readyRef.current) return;
    setIsDirty(true);
    setSavePhase((p) => (p === "saving" ? p : "dirty"));
  }, []);

  const track = useCallback(
    <T,>(setter: (v: T) => void): ((v: T) => void) => {
      const cache = trackedCache.current as WeakMap<(v: T) => void, (v: T) => void>;
      let wrapped = cache.get(setter);
      if (!wrapped) {
        wrapped = (v: T) => {
          setter(v);
          touch();
        };
        cache.set(setter, wrapped);
      }
      return wrapped;
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

  return useMemo(
    () => ({
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
    }),
    [
      isDirty,
      savePhase,
      lastSavedAt,
      lastSaveWasAutosave,
      touch,
      track,
      markClean,
      markSaving,
      markError,
      markHydrated,
      resumeReady,
      pauseReady,
    ],
  );
}
