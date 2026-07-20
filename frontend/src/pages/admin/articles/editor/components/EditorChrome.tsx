import type { RefObject } from "react";
import type { Category, Tag } from "@/api/articles";
import type { ScheduledPublication } from "@/api/scheduledPublications";
import type { Editor } from "@tiptap/react";
import { EditorToolbar, type ModalControls } from "@/components/admin/RichTextEditor";
import MarkdownToolbar from "@/components/admin/editor/MarkdownToolbar";
import type { MarkdownSelectionApi } from "@/components/admin/editor/MarkdownToolbar";
import ArticleForm from "../ArticleForm";
import { SeoFieldsPanel, AdvancedSettingsPanel } from "../SeoFields";
import type { EditorSavePhase } from "../saveStatusUtils";
import type { EditorMetaPanel } from "../hooks/useEditorShell";
import { EditorActionBar } from "./EditorActionBar";
import { EditorLangBar } from "./EditorLangBar";
import { FindReplaceBar } from "./FindReplaceBar";

type WordStat = { chars: number; words: number };

type LangEntry = {
  editor: Editor | null;
  modals: ModalControls;
};

/**
 * Sticky top chrome: action bar, meta panels, lang bar, toolbars, find bar.
 */
export function EditorChrome({
  zenMode,
  title,
  titlePlaceholder,
  onTitleChange,
  onBack,
  savePhase,
  lastSavedAt,
  lastSaveWasAutosave,
  metaPanel,
  onToggleMetaPanel,
  showVersionHistory,
  isEditing,
  canPublish,
  saving,
  articleStatus,
  scheduledPublication,
  scheduleLoading,
  scheduleBusy,
  onToggleZen,
  onOpenHistory,
  onOpenTemplate,
  onPreview,
  onFind,
  onSave,
  onPublish,
  onSchedule,
  onCancelSchedule,
  onRetrySchedule,
  onRefreshSchedule,
  // meta forms
  formSlug,
  setSlug,
  formAuthor,
  setAuthor,
  formCover,
  setCover,
  showCoverPicker,
  setShowCoverPicker,
  categories,
  selectedCategoryIds,
  onToggleCategory,
  tags,
  selectedTagIds,
  onToggleTag,
  zhSeoTitle,
  setZhSeoTitle,
  enSeoTitle,
  setEnSeoTitle,
  zhMetaDescription,
  setZhMetaDescription,
  enMetaDescription,
  setEnMetaDescription,
  ogImage,
  setOgImage,
  visibility,
  setVisibility,
  autoSummary,
  setAutoSummary,
  allowComments,
  setAllowComments,
  pinned,
  setPinned,
  metadata,
  setMetadata,
  // lang
  enabledLangs,
  activeLangIdx,
  viewLayout,
  wordStats,
  editorMode,
  translateBusy,
  showLangMenu,
  langMenuRef,
  onSelectLang,
  onRemoveLang,
  onAddLang,
  onToggleLangMenu,
  onToggleSplit,
  onCopyZhToEn,
  onTranslateZhToEn,
  onModeChange,
  // toolbar / find
  activeEditorEntry,
  markdownApi,
  findOpen,
  onCloseFind,
}: {
  zenMode: boolean;
  title: string;
  titlePlaceholder: string;
  onTitleChange: (v: string) => void;
  onBack: () => void;
  savePhase: EditorSavePhase;
  lastSavedAt: Date | null;
  lastSaveWasAutosave: boolean;
  metaPanel: EditorMetaPanel;
  onToggleMetaPanel: (p: Exclude<EditorMetaPanel, null>) => void;
  showVersionHistory: boolean;
  isEditing: boolean;
  canPublish: boolean;
  saving: boolean;
  articleStatus: "draft" | "published" | "scheduled";
  scheduledPublication: ScheduledPublication | null;
  scheduleLoading: boolean;
  scheduleBusy: boolean;
  onToggleZen: () => void;
  onOpenHistory: () => void;
  onOpenTemplate: () => void;
  onPreview: () => void;
  onFind: () => void;
  onSave: () => void;
  onPublish: () => void;
  onSchedule: (at: string) => void;
  onCancelSchedule: () => void;
  onRetrySchedule: () => void;
  onRefreshSchedule: () => void;
  formSlug: string;
  setSlug: (v: string) => void;
  formAuthor: string;
  setAuthor: (v: string) => void;
  formCover: string;
  setCover: (v: string) => void;
  showCoverPicker: boolean;
  setShowCoverPicker: (v: boolean) => void;
  categories: Category[];
  selectedCategoryIds: number[];
  onToggleCategory: (id: number) => void;
  tags: Tag[];
  selectedTagIds: number[];
  onToggleTag: (id: number) => void;
  zhSeoTitle: string;
  setZhSeoTitle: (v: string) => void;
  enSeoTitle: string;
  setEnSeoTitle: (v: string) => void;
  zhMetaDescription: string;
  setZhMetaDescription: (v: string) => void;
  enMetaDescription: string;
  setEnMetaDescription: (v: string) => void;
  ogImage: string;
  setOgImage: (v: string) => void;
  visibility: string;
  setVisibility: (v: string) => void;
  autoSummary: boolean;
  setAutoSummary: (v: boolean) => void;
  allowComments: boolean;
  setAllowComments: (v: boolean) => void;
  pinned: boolean;
  setPinned: (v: boolean) => void;
  metadata: Record<string, unknown>;
  setMetadata: (v: Record<string, unknown>) => void;
  enabledLangs: string[];
  activeLangIdx: number;
  viewLayout: "focus" | "split";
  wordStats: Record<"zh" | "en", WordStat>;
  editorMode: "richtext" | "markdown";
  translateBusy: boolean;
  showLangMenu: boolean;
  langMenuRef: RefObject<HTMLDivElement | null>;
  onSelectLang: (idx: number) => void;
  onRemoveLang: (key: string) => void;
  onAddLang: (key: string) => void;
  onToggleLangMenu: () => void;
  onToggleSplit: () => void;
  onCopyZhToEn: () => void;
  onTranslateZhToEn: () => void;
  onModeChange: (mode: "richtext" | "markdown") => void;
  activeEditorEntry: LangEntry | undefined;
  markdownApi: MarkdownSelectionApi | null;
  findOpen: boolean;
  onCloseFind: () => void;
}) {
  return (
    <div className="flex-shrink-0 z-20 bg-white border-b border-slate-200 shadow-sm">
      <EditorActionBar
        title={title}
        titlePlaceholder={titlePlaceholder}
        onTitleChange={onTitleChange}
        onBack={onBack}
        savePhase={savePhase}
        lastSavedAt={lastSavedAt}
        lastSaveWasAutosave={lastSaveWasAutosave}
        showBasicInfo={metaPanel === "basic"}
        showSeo={metaPanel === "seo"}
        showAdvanced={metaPanel === "advanced"}
        showVersionHistory={showVersionHistory}
        isEditing={isEditing}
        canPublish={canPublish}
        saving={saving}
        articleStatus={articleStatus}
        scheduledPublication={scheduledPublication}
        scheduleLoading={scheduleLoading}
        scheduleBusy={scheduleBusy}
        zenMode={zenMode}
        onToggleZen={onToggleZen}
        onToggleBasic={() => onToggleMetaPanel("basic")}
        onToggleSeo={() => onToggleMetaPanel("seo")}
        onToggleAdvanced={() => onToggleMetaPanel("advanced")}
        onOpenHistory={onOpenHistory}
        onOpenTemplate={onOpenTemplate}
        onPreview={onPreview}
        onFind={onFind}
        onSave={onSave}
        onPublish={onPublish}
        onSchedule={onSchedule}
        onCancelSchedule={onCancelSchedule}
        onRetrySchedule={onRetrySchedule}
        onRefreshSchedule={onRefreshSchedule}
      />

      {!zenMode && metaPanel === "basic" && (
        <ArticleForm
          slug={formSlug}
          setSlug={setSlug}
          author={formAuthor}
          setAuthor={setAuthor}
          coverImage={formCover}
          setCoverImage={setCover}
          showCoverPicker={showCoverPicker}
          setShowCoverPicker={setShowCoverPicker}
          categories={categories}
          selectedCategoryIds={selectedCategoryIds}
          toggleCategory={onToggleCategory}
          tags={tags}
          selectedTagIds={selectedTagIds}
          toggleTag={onToggleTag}
        />
      )}
      {!zenMode && metaPanel === "seo" && (
        <SeoFieldsPanel
          zhSeoTitle={zhSeoTitle}
          setZhSeoTitle={setZhSeoTitle}
          enSeoTitle={enSeoTitle}
          setEnSeoTitle={setEnSeoTitle}
          zhMetaDescription={zhMetaDescription}
          setZhMetaDescription={setZhMetaDescription}
          enMetaDescription={enMetaDescription}
          setEnMetaDescription={setEnMetaDescription}
          ogImage={ogImage}
          setOgImage={setOgImage}
        />
      )}
      {!zenMode && metaPanel === "advanced" && (
        <AdvancedSettingsPanel
          visibility={visibility}
          setVisibility={setVisibility}
          autoSummary={autoSummary}
          setAutoSummary={setAutoSummary}
          allowComments={allowComments}
          setAllowComments={setAllowComments}
          pinned={pinned}
          setPinned={setPinned}
          metadata={metadata}
          setMetadata={setMetadata}
        />
      )}

      {!zenMode && (
        <EditorLangBar
          enabledLangs={enabledLangs}
          activeLangIdx={activeLangIdx}
          viewLayout={viewLayout}
          wordStats={wordStats}
          editorMode={editorMode}
          translateBusy={translateBusy}
          showLangMenu={showLangMenu}
          langMenuRef={langMenuRef}
          onSelectLang={onSelectLang}
          onRemoveLang={onRemoveLang}
          onAddLang={onAddLang}
          onToggleLangMenu={onToggleLangMenu}
          onToggleSplit={onToggleSplit}
          onCopyZhToEn={onCopyZhToEn}
          onTranslateZhToEn={onTranslateZhToEn}
          onModeChange={onModeChange}
        />
      )}

      {!zenMode && (
        <div className="flex items-stretch border-t border-slate-200 bg-slate-50">
          {editorMode === "richtext" && activeEditorEntry?.editor ? (
            <div className="flex-1 min-w-0 overflow-x-auto">
              <EditorToolbar
                editor={activeEditorEntry.editor}
                modals={activeEditorEntry.modals}
              />
            </div>
          ) : editorMode === "markdown" ? (
            <div className="flex-1 min-w-0 overflow-x-auto">
              <MarkdownToolbar api={markdownApi} />
            </div>
          ) : (
            <div className="flex-1 py-2" />
          )}
        </div>
      )}

      {editorMode === "richtext" && (
        <FindReplaceBar
          open={findOpen}
          editor={activeEditorEntry?.editor ?? null}
          onClose={onCloseFind}
        />
      )}
    </div>
  );
}
