import ImagePickerModal from "@/components/admin/ImagePickerModal";
import type { Editor } from "@tiptap/react";
import { EditorModals } from "@/components/admin/RichTextEditor";
import type { ModalState } from "@/components/admin/RichTextEditor";
import { ArticleVersionHistoryPanel, type ArticleDraftSnapshot } from "../VersionHistoryPanel";
import ArticlePreviewModal, { type ArticlePreviewData } from "../ArticlePreviewModal";
import ArticleConflictDialog from "../ArticleConflictDialog";
import TemplatePickerModal from "../TemplatePickerModal";
import type { ArticleTemplate } from "../articleTemplates";
import { PublishChecklistDialog } from "./PublishChecklistDialog";
import type { ChecklistItem } from "../utils/publishChecklist";

type LangEntry = {
  editor: Editor | null;
  state: ModalState;
};

/**
 * All editor page overlays / dialogs (version, preview, conflict, checklist, …).
 */
export function EditorDialogs({
  langEditors,
  showCoverPicker,
  onCloseCoverPicker,
  onSelectCover,
  showVersionHistory,
  isEditing,
  articleId,
  versionDraftSnapshot,
  onCloseVersionHistory,
  onRestoreVersion,
  showPreview,
  previewData,
  onClosePreview,
  showTemplatePicker,
  onCloseTemplatePicker,
  onSelectTemplate,
  conflict,
  saving,
  onDismissConflict,
  onReloadConflict,
  onForceOverwrite,
  publishChecklistOpen,
  publishChecklistItems,
  onCancelPublishChecklist,
  onForcePublish,
}: {
  langEditors: Record<string, LangEntry>;
  showCoverPicker: boolean;
  onCloseCoverPicker: () => void;
  onSelectCover: (url: string) => void;
  showVersionHistory: boolean;
  isEditing: boolean;
  articleId: string | undefined;
  versionDraftSnapshot: ArticleDraftSnapshot | null;
  onCloseVersionHistory: () => void;
  onRestoreVersion: (snap: ArticleDraftSnapshot) => void;
  showPreview: boolean;
  previewData: ArticlePreviewData | null;
  onClosePreview: () => void;
  showTemplatePicker: boolean;
  onCloseTemplatePicker: () => void;
  onSelectTemplate: (tpl: ArticleTemplate) => void;
  conflict: { serverUpdatedAt?: string } | null;
  saving: boolean;
  onDismissConflict: () => void;
  onReloadConflict: () => void;
  onForceOverwrite: () => void;
  publishChecklistOpen: boolean;
  publishChecklistItems: ChecklistItem[];
  onCancelPublishChecklist: () => void;
  onForcePublish?: () => void;
}) {
  return (
    <>
      {Object.entries(langEditors).map(([lang, entry]) =>
        entry.editor ? (
          <EditorModals key={lang} editor={entry.editor} state={entry.state} />
        ) : null,
      )}

      <ImagePickerModal
        open={showCoverPicker}
        onClose={onCloseCoverPicker}
        onSelect={(item) => onSelectCover(item.url)}
      />

      {showVersionHistory && isEditing && articleId && (
        <ArticleVersionHistoryPanel
          articleId={Number(articleId)}
          onClose={onCloseVersionHistory}
          currentDraft={versionDraftSnapshot}
          onRestore={onRestoreVersion}
          canRestore
        />
      )}

      <ArticlePreviewModal
        open={showPreview}
        data={previewData}
        onClose={onClosePreview}
      />
      <TemplatePickerModal
        open={showTemplatePicker}
        onClose={onCloseTemplatePicker}
        onSelect={onSelectTemplate}
      />

      {conflict && (
        <ArticleConflictDialog
          serverUpdatedAt={conflict.serverUpdatedAt}
          busy={saving}
          onDismiss={onDismissConflict}
          onReload={onReloadConflict}
          onForceOverwrite={onForceOverwrite}
        />
      )}

      <PublishChecklistDialog
        open={publishChecklistOpen}
        items={publishChecklistItems}
        busy={saving}
        onCancel={onCancelPublishChecklist}
        onForcePublish={onForcePublish}
      />
    </>
  );
}
