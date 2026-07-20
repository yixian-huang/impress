/** Pure helpers for article editor save / leave / mode-switch UX. */

export type EditorSavePhase = "clean" | "dirty" | "saving" | "saved" | "error";

/** Resolve which status field to send when saving (preserve published live content). */
export function resolveSaveStatus(
  intent: "draft" | "publish" | "autosave",
  articleStatus: "draft" | "published" | "scheduled",
): "draft" | "published" {
  if (intent === "publish") return "published";
  // Content updates on already-published articles stay published (no accidental unpublish).
  if (articleStatus === "published") return "published";
  return "draft";
}

export const MODE_SWITCH_MESSAGE =
  "切换编辑模式可能丢失部分格式（复杂排版、部分样式等）。是否继续？";

export const LEAVE_UNSAVED_MESSAGE = "有未保存的更改，确定离开？";
