/**
 * Browser local draft recovery for the article editor.
 * Complements server autosave when the tab crashes or network is offline.
 */

export const LOCAL_DRAFT_VERSION = 1 as const;
export const LOCAL_DRAFT_PREFIX = "inkless:article-local-draft:";
/** Debounce writes while typing */
export const LOCAL_DRAFT_WRITE_MS = 800;
/** Ignore drafts older than 14 days */
export const LOCAL_DRAFT_MAX_AGE_MS = 14 * 24 * 60 * 60 * 1000;

export type LocalDraftEditorMode = "richtext" | "markdown";

export type LocalDraftSnapshot = {
  v: typeof LOCAL_DRAFT_VERSION;
  key: string;
  savedAt: string;
  /** Server article.updatedAt when this draft was last aligned (if any) */
  baseUpdatedAt: string | null;
  editorMode: LocalDraftEditorMode;
  enabledLangs: string[];
  zhTitle: string;
  enTitle: string;
  slug: string;
  coverImage: string;
  zhBody: string;
  enBody: string;
  zhSeoTitle: string;
  enSeoTitle: string;
  zhMetaDescription: string;
  enMetaDescription: string;
  ogImage: string;
  author: string;
  markdownZh: string;
  markdownEn: string;
};

export function localDraftStorageKey(articleId: string | undefined | null): string {
  return LOCAL_DRAFT_PREFIX + (articleId ? String(articleId) : "new");
}

export function isEmptyHtml(html: string | undefined | null): boolean {
  if (!html) return true;
  const t = html.replace(/<[^>]+>/g, "").replace(/&nbsp;/g, " ").trim();
  return t.length === 0;
}

/** Whether the draft has enough content to offer recovery. */
export function isLocalDraftUseful(draft: LocalDraftSnapshot | null | undefined): boolean {
  if (!draft || draft.v !== LOCAL_DRAFT_VERSION) return false;
  if (draft.zhTitle.trim()) return true;
  if (draft.enTitle.trim()) return true;
  if (!isEmptyHtml(draft.zhBody) || !isEmptyHtml(draft.enBody)) return true;
  if ((draft.markdownZh || "").trim() || (draft.markdownEn || "").trim()) return true;
  return false;
}

export function isLocalDraftExpired(draft: LocalDraftSnapshot, now = Date.now()): boolean {
  const t = Date.parse(draft.savedAt);
  if (Number.isNaN(t)) return true;
  return now - t > LOCAL_DRAFT_MAX_AGE_MS;
}

/**
 * Offer recovery when local has useful content and either:
 * - differs from the server-hydrated snapshot, or
 * - is newer than the known server baseUpdatedAt.
 */
export function shouldOfferLocalDraft(
  local: LocalDraftSnapshot | null,
  server: {
    zhTitle: string;
    enTitle: string;
    zhBody: string;
    enBody: string;
    baseUpdatedAt: string | null;
  },
): boolean {
  if (!local || !isLocalDraftUseful(local) || isLocalDraftExpired(local)) return false;

  const titleDiff =
    local.zhTitle.trim() !== (server.zhTitle || "").trim()
    || local.enTitle.trim() !== (server.enTitle || "").trim();
  // Normalize empty editor shells (<p></p>) so they compare equal to ""
  const bodyDiff =
    !isEmptyHtml(local.zhBody) !== !isEmptyHtml(server.zhBody)
    || (!isEmptyHtml(local.zhBody)
      && (local.zhBody || "") !== (server.zhBody || ""))
    || !isEmptyHtml(local.enBody) !== !isEmptyHtml(server.enBody)
    || (!isEmptyHtml(local.enBody)
      && (local.enBody || "") !== (server.enBody || ""));
  if (titleDiff || bodyDiff) return true;

  if (local.baseUpdatedAt && server.baseUpdatedAt && local.baseUpdatedAt !== server.baseUpdatedAt) {
    return true;
  }

  // Same content as server — no need to recover
  return false;
}

export function readLocalDraft(articleId: string | undefined | null): LocalDraftSnapshot | null {
  if (typeof localStorage === "undefined") return null;
  try {
    const raw = localStorage.getItem(localDraftStorageKey(articleId));
    if (!raw) return null;
    const parsed = JSON.parse(raw) as LocalDraftSnapshot;
    if (!parsed || parsed.v !== LOCAL_DRAFT_VERSION) return null;
    if (isLocalDraftExpired(parsed)) {
      clearLocalDraft(articleId);
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

export function writeLocalDraft(draft: LocalDraftSnapshot): boolean {
  if (typeof localStorage === "undefined") return false;
  try {
    localStorage.setItem(localDraftStorageKey(draft.key === "new" ? null : draft.key), JSON.stringify(draft));
    return true;
  } catch {
    // Quota exceeded or private mode
    return false;
  }
}

export function clearLocalDraft(articleId: string | undefined | null): void {
  if (typeof localStorage === "undefined") return;
  try {
    localStorage.removeItem(localDraftStorageKey(articleId));
  } catch {
    /* ignore */
  }
}

/** Format savedAt for banner UI. */
export function formatLocalDraftTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString("zh-CN", {
      month: "numeric",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}
