import { useCallback, useMemo, useState } from "react";
import type { ArticleDraftSnapshot } from "../VersionHistoryPanel";
import type { ArticlePreviewData } from "../ArticlePreviewModal";

/** Exclusive secondary panel under the action bar. */
export type EditorMetaPanel = "basic" | "seo" | "advanced" | null;

/**
 * Chrome / dialog visibility for the article editor page.
 * Keeps page.tsx free of 15+ independent booleans.
 */
export function useEditorShell() {
  const [metaPanel, setMetaPanel] = useState<EditorMetaPanel>(null);
  const [showCoverPicker, setShowCoverPicker] = useState(false);
  const [showVersionHistory, setShowVersionHistory] = useState(false);
  const [versionDraftSnapshot, setVersionDraftSnapshot] = useState<ArticleDraftSnapshot | null>(null);
  const [showPreview, setShowPreview] = useState(false);
  const [previewData, setPreviewData] = useState<ArticlePreviewData | null>(null);
  const [showTemplatePicker, setShowTemplatePicker] = useState(false);
  const [showLangMenu, setShowLangMenu] = useState(false);
  const [zenMode, setZenMode] = useState(false);
  const [findOpen, setFindOpen] = useState(false);
  const [showShortcutHelp, setShowShortcutHelp] = useState(false);

  const toggleMetaPanel = useCallback((panel: Exclude<EditorMetaPanel, null>) => {
    setMetaPanel((cur) => (cur === panel ? null : panel));
  }, []);

  const openPreviewWith = useCallback((data: ArticlePreviewData) => {
    setPreviewData(data);
    setShowPreview(true);
  }, []);

  const closePreview = useCallback(() => setShowPreview(false), []);

  const openVersionHistory = useCallback((snapshot: ArticleDraftSnapshot) => {
    setVersionDraftSnapshot(snapshot);
    setShowVersionHistory(true);
  }, []);

  const closeVersionHistory = useCallback(() => {
    setShowVersionHistory(false);
    setVersionDraftSnapshot(null);
  }, []);

  const toggleZen = useCallback(() => setZenMode((z) => !z), []);
  const openFind = useCallback(() => setFindOpen(true), []);
  const closeFind = useCallback(() => setFindOpen(false), []);
  const toggleLangMenu = useCallback(() => setShowLangMenu((v) => !v), []);
  const openShortcutHelp = useCallback(() => setShowShortcutHelp(true), []);
  const closeShortcutHelp = useCallback(() => setShowShortcutHelp(false), []);
  const toggleShortcutHelp = useCallback(() => setShowShortcutHelp((v) => !v), []);

  return useMemo(
    () => ({
      metaPanel,
      showBasicInfo: metaPanel === "basic",
      showSeo: metaPanel === "seo",
      showAdvanced: metaPanel === "advanced",
      toggleMetaPanel,
      showCoverPicker,
      setShowCoverPicker,
      showVersionHistory,
      versionDraftSnapshot,
      openVersionHistory,
      closeVersionHistory,
      showPreview,
      previewData,
      openPreviewWith,
      closePreview,
      showTemplatePicker,
      setShowTemplatePicker,
      showLangMenu,
      setShowLangMenu,
      toggleLangMenu,
      zenMode,
      toggleZen,
      findOpen,
      openFind,
      closeFind,
      showShortcutHelp,
      openShortcutHelp,
      closeShortcutHelp,
      toggleShortcutHelp,
    }),
    [
      metaPanel,
      toggleMetaPanel,
      showCoverPicker,
      showVersionHistory,
      versionDraftSnapshot,
      openVersionHistory,
      closeVersionHistory,
      showPreview,
      previewData,
      openPreviewWith,
      closePreview,
      showTemplatePicker,
      showLangMenu,
      toggleLangMenu,
      zenMode,
      toggleZen,
      findOpen,
      openFind,
      closeFind,
      showShortcutHelp,
      openShortcutHelp,
      closeShortcutHelp,
      toggleShortcutHelp,
    ],
  );
}

export type EditorShell = ReturnType<typeof useEditorShell>;
