import type { LocalDraftSnapshot } from "./localDraft";
import type { EditorMode } from "../hooks/useArticleEditors";

/** Form fields mutated when restoring a local draft. */
export type LocalDraftFormApply = {
  setZhTitle: (v: string) => void;
  setEnTitle: (v: string) => void;
  setSlug: (v: string) => void;
  setCoverImage: (v: string) => void;
  setZhSeoTitle: (v: string) => void;
  setEnSeoTitle: (v: string) => void;
  setZhMetaDescription: (v: string) => void;
  setEnMetaDescription: (v: string) => void;
  setOgImage: (v: string) => void;
  setAuthor: (v: string) => void;
};

export type LocalDraftEditorApply = {
  applyBodiesToEditors: (zh: string, en: string) => void;
  setMarkdownContent: (v: Record<string, string> | ((prev: Record<string, string>) => Record<string, string>)) => void;
  setEditorMode: (m: EditorMode) => void;
  ensureEnEnabled: () => void;
};

/** Pure side-effect helper used by page restore (keeps page free of field lists). */
export function applyLocalDraftToFormAndEditors(
  draft: LocalDraftSnapshot,
  form: LocalDraftFormApply,
  editors: LocalDraftEditorApply,
): void {
  form.setZhTitle(draft.zhTitle || "");
  form.setEnTitle(draft.enTitle || "");
  form.setSlug(draft.slug || "");
  form.setCoverImage(draft.coverImage || "");
  form.setZhSeoTitle(draft.zhSeoTitle || "");
  form.setEnSeoTitle(draft.enSeoTitle || "");
  form.setZhMetaDescription(draft.zhMetaDescription || "");
  form.setEnMetaDescription(draft.enMetaDescription || "");
  form.setOgImage(draft.ogImage || "");
  form.setAuthor(draft.author || "");
  editors.applyBodiesToEditors(draft.zhBody || "<p></p>", draft.enBody || "<p></p>");
  if (draft.editorMode === "markdown") {
    editors.setMarkdownContent({
      zh: draft.markdownZh || "",
      en: draft.markdownEn || "",
    });
    editors.setEditorMode("markdown");
  } else {
    editors.setEditorMode("richtext");
  }
  if (draft.enabledLangs?.includes("en") || draft.enTitle || draft.enBody) {
    editors.ensureEnEnabled();
  }
}
