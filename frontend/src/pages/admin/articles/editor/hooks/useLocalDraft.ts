import { useCallback, useEffect, useRef, useState } from "react";
import {
  LOCAL_DRAFT_WRITE_MS,
  clearLocalDraft,
  formatLocalDraftTime,
  readLocalDraft,
  shouldOfferLocalDraft,
  writeLocalDraft,
  type LocalDraftSnapshot,
} from "../utils/localDraft";

export type LocalDraftOffer = {
  draft: LocalDraftSnapshot;
  savedAtLabel: string;
};

/**
 * Persist unsaved editor state to localStorage and offer recovery after load.
 */
export function useLocalDraft(opts: {
  articleId: string | undefined;
  loading: boolean;
  isDirty: boolean;
  baseUpdatedAt: string | null;
  /** Build a full snapshot of current editor state (read at write time). */
  getSnapshot: () => Omit<LocalDraftSnapshot, "v" | "key" | "savedAt">;
  /** Apply a recovered draft into form + editors. */
  applySnapshot: (draft: LocalDraftSnapshot) => void;
  /** Server-hydrated comparison fields (after load). */
  getServerCompare: () => {
    zhTitle: string;
    enTitle: string;
    zhBody: string;
    enBody: string;
  };
}) {
  const {
    articleId,
    loading,
    isDirty,
    baseUpdatedAt,
    getSnapshot,
    applySnapshot,
    getServerCompare,
  } = opts;

  const [offer, setOffer] = useState<LocalDraftOffer | null>(null);
  const checkedKeyRef = useRef<string | null>(null);
  const getSnapshotRef = useRef(getSnapshot);
  getSnapshotRef.current = getSnapshot;
  const getServerCompareRef = useRef(getServerCompare);
  getServerCompareRef.current = getServerCompare;
  const applySnapshotRef = useRef(applySnapshot);
  applySnapshotRef.current = applySnapshot;

  // After server load (or on new article), check for recoverable local draft once per id.
  useEffect(() => {
    if (loading) return;
    const key = articleId ? String(articleId) : "new";
    if (checkedKeyRef.current === key) return;
    checkedKeyRef.current = key;

    const local = readLocalDraft(articleId);
    if (!local) {
      setOffer(null);
      return;
    }
    const server = getServerCompareRef.current();
    if (
      shouldOfferLocalDraft(local, {
        ...server,
        baseUpdatedAt,
      })
    ) {
      setOffer({ draft: local, savedAtLabel: formatLocalDraftTime(local.savedAt) });
    } else {
      // Stale / identical — drop to avoid future noise
      if (local && local.baseUpdatedAt && baseUpdatedAt && local.baseUpdatedAt === baseUpdatedAt) {
        // keep if dirty offline? already identical — clear
        clearLocalDraft(articleId);
      }
      setOffer(null);
    }
  }, [loading, articleId, baseUpdatedAt]);

  // Debounced write while dirty
  useEffect(() => {
    if (!isDirty || loading) return;
    const t = window.setTimeout(() => {
      const snap = getSnapshotRef.current();
      const key = articleId ? String(articleId) : "new";
      writeLocalDraft({
        v: 1,
        key,
        savedAt: new Date().toISOString(),
        ...snap,
      });
    }, LOCAL_DRAFT_WRITE_MS);
    return () => window.clearTimeout(t);
  }, [isDirty, loading, articleId, baseUpdatedAt]);

  // Also flush on page hide / unload while dirty
  useEffect(() => {
    const flush = () => {
      if (!isDirty) return;
      const snap = getSnapshotRef.current();
      const key = articleId ? String(articleId) : "new";
      writeLocalDraft({
        v: 1,
        key,
        savedAt: new Date().toISOString(),
        ...snap,
      });
    };
    window.addEventListener("pagehide", flush);
    window.addEventListener("beforeunload", flush);
    return () => {
      window.removeEventListener("pagehide", flush);
      window.removeEventListener("beforeunload", flush);
    };
  }, [isDirty, articleId]);

  const clear = useCallback(() => {
    clearLocalDraft(articleId);
    // Also clear "new" when we just created an article
    if (articleId) clearLocalDraft(null);
    setOffer(null);
  }, [articleId]);

  const restore = useCallback(() => {
    if (!offer) return;
    applySnapshotRef.current(offer.draft);
    setOffer(null);
    // Keep local until server save succeeds (still dirty)
  }, [offer]);

  const dismiss = useCallback(() => {
    clearLocalDraft(articleId);
    setOffer(null);
  }, [articleId]);

  return {
    offer,
    restore,
    dismiss,
    clear,
  };
}
