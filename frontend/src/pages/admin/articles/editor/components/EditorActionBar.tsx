import type { ScheduledPublication } from "@/api/scheduledPublications";
import { ScheduledPublicationPanel } from "@/components/admin/ScheduledPublicationPanel";
import { PopoverButton } from "../SeoFields";
import { SaveStatusBadge } from "../saveStatus";
import type { EditorSavePhase } from "../saveStatusUtils";

export function EditorActionBar({
  title,
  titlePlaceholder,
  onTitleChange,
  onBack,
  savePhase,
  lastSavedAt,
  lastSaveWasAutosave,
  showBasicInfo,
  showSeo,
  showAdvanced,
  showVersionHistory,
  isEditing,
  canPublish,
  saving,
  articleStatus,
  scheduledPublication,
  scheduleLoading,
  scheduleBusy,
  zenMode,
  onToggleZen,
  onOpenShortcutHelp,
  onToggleBasic,
  onToggleSeo,
  onToggleAdvanced,
  onOpenHistory,
  onOpenTemplate,
  onPreview,
  onFind,
  onOpenAIMeta,
  aiMetaBusy,
  onSave,
  onPublish,
  onSchedule,
  onCancelSchedule,
  onRetrySchedule,
  onRefreshSchedule,
}: {
  title: string;
  titlePlaceholder: string;
  onTitleChange: (v: string) => void;
  onBack: () => void;
  savePhase: EditorSavePhase;
  lastSavedAt: Date | null;
  lastSaveWasAutosave: boolean;
  showBasicInfo: boolean;
  showSeo: boolean;
  showAdvanced: boolean;
  showVersionHistory: boolean;
  isEditing: boolean;
  canPublish: boolean;
  saving: boolean;
  articleStatus: "draft" | "published" | "scheduled";
  scheduledPublication: ScheduledPublication | null;
  scheduleLoading: boolean;
  scheduleBusy: boolean;
  zenMode?: boolean;
  onToggleZen?: () => void;
  onOpenShortcutHelp?: () => void;
  onToggleBasic: () => void;
  onToggleSeo: () => void;
  onToggleAdvanced: () => void;
  onOpenHistory: () => void;
  onOpenTemplate: () => void;
  onPreview: () => void;
  onFind?: () => void;
  onOpenAIMeta?: () => void;
  aiMetaBusy?: boolean;
  onSave: () => void;
  onPublish: () => void;
  onSchedule: (at: string) => void;
  onCancelSchedule: () => void;
  onRetrySchedule: () => void;
  onRefreshSchedule: () => void;
}) {
  return (
    <div className="flex items-center gap-3 px-4 py-2">
      {!zenMode && (
        <button type="button" onClick={onBack} className="text-slate-500 hover:text-slate-700 text-sm flex-shrink-0">
          &larr; 返回
        </button>
      )}
      <input
        type="text"
        value={title}
        onChange={(e) => onTitleChange(e.target.value)}
        className="flex-1 px-3 py-1.5 text-base font-semibold border border-transparent rounded-lg hover:border-slate-200 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none transition-colors bg-transparent"
        placeholder={titlePlaceholder}
      />
      <div className="flex items-center gap-1.5 flex-shrink-0">
        <SaveStatusBadge phase={savePhase} lastSavedAt={lastSavedAt} isAutosave={lastSaveWasAutosave} />
        {!zenMode && (
          <>
            <PopoverButton label="基本信息" active={showBasicInfo} onClick={onToggleBasic} />
            <PopoverButton label="SEO" active={showSeo} onClick={onToggleSeo} />
            <PopoverButton label="高级" active={showAdvanced} onClick={onToggleAdvanced} />
            {isEditing && (
              <PopoverButton label="历史版本" active={showVersionHistory} onClick={onOpenHistory} />
            )}
            <button
              type="button"
              onClick={onOpenTemplate}
              title="应用文章结构模板"
              className="px-2.5 py-1.5 text-xs border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-700"
            >
              模板
            </button>
            {onOpenAIMeta && (
              <button
                type="button"
                onClick={onOpenAIMeta}
                disabled={aiMetaBusy}
                title="根据正文生成标题、slug、SEO 与 Meta"
                className="px-2.5 py-1.5 text-xs border border-violet-200 bg-violet-50 text-violet-800 rounded-lg hover:bg-violet-100 disabled:opacity-50"
              >
                {aiMetaBusy ? "AI…" : "AI 元数据"}
              </button>
            )}
            <button
              type="button"
              onClick={onPreview}
              title="预览 (⌘P / Ctrl+P)"
              className="px-2.5 py-1.5 text-xs border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-700"
            >
              预览
            </button>
            {onFind && (
              <button
                type="button"
                onClick={onFind}
                title="查找替换 (⌘F / Ctrl+F)"
                className="px-2.5 py-1.5 text-xs border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-700"
              >
                查找
              </button>
            )}
            <span className="w-px h-6 bg-slate-200 mx-1" />
            <ScheduledPublicationPanel
              compact
              item={scheduledPublication}
              loading={scheduleLoading}
              busy={scheduleBusy}
              canPublish={canPublish}
              disabledReason="需要 articles:publish 权限才能安排定时发布。"
              onSchedule={onSchedule}
              onCancel={onCancelSchedule}
              onRetry={onRetrySchedule}
              onRefresh={onRefreshSchedule}
              title={articleStatus === "published" ? "定时更新" : "定时"}
            />
            <span className="w-px h-6 bg-slate-200 mx-1" />
          </>
        )}
        {onToggleZen && (
          <button
            type="button"
            onClick={onToggleZen}
            title={zenMode ? "退出专注模式 (Esc)" : "专注模式 (⌘\\ / Ctrl+\\)"}
            className={`px-2.5 py-1.5 text-xs border rounded-lg ${
              zenMode
                ? "border-blue-300 bg-blue-50 text-blue-800"
                : "border-slate-200 hover:bg-slate-50 text-slate-700"
            }`}
          >
            {zenMode ? "退出专注" : "专注"}
          </button>
        )}
        {onOpenShortcutHelp && (
          <button
            type="button"
            onClick={onOpenShortcutHelp}
            title="快捷键 (⌘/ / Ctrl+/)"
            className="px-2.5 py-1.5 text-xs border border-slate-200 rounded-lg hover:bg-slate-50 text-slate-700"
          >
            快捷键
          </button>
        )}
        <button
          type="button"
          onClick={onSave}
          disabled={saving}
          title="保存 (⌘S / Ctrl+S)"
          className="px-3 py-1.5 text-sm border border-slate-200 rounded-lg hover:bg-slate-50 disabled:opacity-50"
        >
          {saving ? "保存中..." : "保存"}
        </button>
        {canPublish && !zenMode && (
          <button
            type="button"
            onClick={onPublish}
            disabled={saving}
            title="发布 (⌘⇧S / Ctrl+Shift+S)"
            className="px-3 py-1.5 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
          >
            {saving ? "发布中..." : "发布"}
          </button>
        )}
      </div>
    </div>
  );
}
