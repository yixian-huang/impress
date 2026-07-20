/** User preferences for the article editor (localStorage). */

export type PreferredEditorMode = "richtext" | "markdown";

export const EDITOR_MODE_PREF_KEY = "inkless.article.editorMode";

export function readPreferredEditorMode(): PreferredEditorMode {
  if (typeof localStorage === "undefined") return "richtext";
  try {
    const v = localStorage.getItem(EDITOR_MODE_PREF_KEY);
    if (v === "markdown" || v === "richtext") return v;
  } catch {
    /* ignore */
  }
  return "richtext";
}

export function writePreferredEditorMode(mode: PreferredEditorMode): void {
  if (typeof localStorage === "undefined") return;
  try {
    localStorage.setItem(EDITOR_MODE_PREF_KEY, mode);
  } catch {
    /* ignore */
  }
}
