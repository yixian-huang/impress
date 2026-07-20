import { useState, useEffect, useCallback, useRef, useMemo } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useEditor } from "@tiptap/react";
import type { Article } from "@/api/articles";
import {
  getAdminArticle,
  createArticle,
  updateArticle,
  isArticleVersionConflict,
  getCategories,
  getTags,
} from "@/api/articles";
import type { Category, Tag } from "@/api/articles";
import { ScheduledPublicationPanel } from "@/components/admin/ScheduledPublicationPanel";
import ImagePickerModal from "@/components/admin/ImagePickerModal";
import {
  getEditorExtensions,
  EditorToolbar,
  EditorModals,
  useModalState,
} from "@/components/admin/RichTextEditor";
import MarkdownToolbar from "@/components/admin/editor/MarkdownToolbar";
import type { MarkdownSelectionApi } from "@/components/admin/editor/MarkdownToolbar";
import { markdownToHtml, htmlToMarkdown } from "@/lib/markdown";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useUnsavedChangesGuard } from "@/hooks/useUnsavedChangesGuard";
import { useAuth } from "@/contexts/AuthContext";
import ArticleForm from "./ArticleForm";
import { SeoFieldsPanel, AdvancedSettingsPanel, PopoverButton } from "./SeoFields";
import { ArticleVersionHistoryPanel, type ArticleDraftSnapshot } from "./VersionHistoryPanel";
import ArticlePreviewModal, { type ArticlePreviewData } from "./ArticlePreviewModal";
import ArticleConflictDialog from "./ArticleConflictDialog";
import TemplatePickerModal from "./TemplatePickerModal";
import type { ArticleTemplate } from "./articleTemplates";
import { translateText } from "@/api/translation";
import { htmlToPlainText, plainTextToHtml } from "./bilingualUtils";
import { SaveStatusBadge } from "./saveStatus";
import {
  LEAVE_UNSAVED_MESSAGE,
  MODE_SWITCH_MESSAGE,
  resolveSaveStatus,
} from "./saveStatusUtils";
import { slugifyTitle } from "./utils/slugify";
import { AUTOSAVE_DEBOUNCE_MS, hasMeaningfulHtml, statusLabelOf, TOAST_MS } from "./utils/constants";
import { useDirtyState } from "./hooks/useDirtyState";
import { useEditorShortcuts } from "./hooks/useEditorShortcuts";
import { useOutsideClick } from "./hooks/useOutsideClick";
import { useSlashMediaBridge } from "./hooks/useSlashMediaBridge";
import { useArticleSchedule } from "./hooks/useArticleSchedule";
import { useWordStats } from "./hooks/useWordStats";
import { EditorMessageBars } from "./components/EditorMessageBars";
import { EditorLangBar } from "./components/EditorLangBar";
import { EditorWorkspace } from "./components/EditorWorkspace";

function toast(setMsg: (s: string) => void, msg: string, ms = TOAST_MS) {
  setMsg(msg);
  window.setTimeout(() => setMsg(""), ms);
}

export default function ArticleEditorPage() {
  useDocumentTitle("编辑文章");
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isEditing = !!id;
  const { hasPermission } = useAuth();
  const canPublish = hasPermission("articles:publish");

  // ── Form fields ──
  const [zhTitle, setZhTitle] = useState("");
  const [enTitle, setEnTitle] = useState("");
  const [slug, setSlug] = useState("");
  const [selectedCategoryIds, setSelectedCategoryIds] = useState<number[]>([]);
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>([]);
  const [coverImage, setCoverImage] = useState("");
  const [zhBody, setZhBody] = useState("");
  const [enBody, setEnBody] = useState("");
  const [zhSeoTitle, setZhSeoTitle] = useState("");
  const [enSeoTitle, setEnSeoTitle] = useState("");
  const [zhMetaDescription, setZhMetaDescription] = useState("");
  const [enMetaDescription, setEnMetaDescription] = useState("");
  const [ogImage, setOgImage] = useState("");
  const [author, setAuthor] = useState("");
  const [autoSummary, setAutoSummary] = useState(false);
  const [allowComments, setAllowComments] = useState(true);
  const [pinned, setPinned] = useState(false);
  const [visibility, setVisibility] = useState("public");
  const [metadata, setMetadata] = useState<Record<string, unknown>>({});
  const [articleCreatedAt, setArticleCreatedAt] = useState<string | null>(null);
  const [articlePublishedAt, setArticlePublishedAt] = useState<string | null>(null);
  const [articleStatus, setArticleStatus] = useState<"draft" | "published" | "scheduled">("draft");

  // ── Shell / meta UI ──
  const [loading, setLoading] = useState(isEditing);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState("");
  const [categories, setCategories] = useState<Category[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [showBasicInfo, setShowBasicInfo] = useState(false);
  const [showSeo, setShowSeo] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [showCoverPicker, setShowCoverPicker] = useState(false);
  const [showVersionHistory, setShowVersionHistory] = useState(false);
  const [versionDraftSnapshot, setVersionDraftSnapshot] = useState<ArticleDraftSnapshot | null>(null);
  const [showPreview, setShowPreview] = useState(false);
  const [previewData, setPreviewData] = useState<ArticlePreviewData | null>(null);
  const [showTemplatePicker, setShowTemplatePicker] = useState(false);
  const [baseUpdatedAt, setBaseUpdatedAt] = useState<string | null>(null);
  const [conflict, setConflict] = useState<{ serverUpdatedAt?: string } | null>(null);

  // ── Editor modes ──
  const [editorMode, setEditorMode] = useState<"richtext" | "markdown">("richtext");
  const [markdownContent, setMarkdownContent] = useState<Record<string, string>>({ zh: "", en: "" });
  const [markdownApi, setMarkdownApi] = useState<MarkdownSelectionApi | null>(null);
  const [viewLayout, setViewLayout] = useState<"focus" | "split">("focus");
  const [translateBusy, setTranslateBusy] = useState(false);
  const [enabledLangs, setEnabledLangs] = useState<string[]>(["zh"]);
  const [activeLangIdx, setActiveLangIdx] = useState(0);
  const [showLangMenu, setShowLangMenu] = useState(false);
  const langMenuRef = useRef<HTMLDivElement>(null);
  const loadedIdRef = useRef<string | null>(null);

  const dirty = useDirtyState(!isEditing);
  const {
    isDirty, savePhase, lastSavedAt, lastSaveWasAutosave,
    readyRef, savingRef, touch, track,
    markClean, markSaving, markError, markHydrated, resumeReady, pauseReady,
  } = dirty;

  // ── TipTap (per-language, isolated extension instances) ──
  const zhExtensions = useMemo(() => getEditorExtensions(), []);
  const enExtensions = useMemo(() => getEditorExtensions(), []);
  const zhEditor = useEditor({
    extensions: zhExtensions,
    content: zhBody,
    shouldRerenderOnTransaction: false,
    editorProps: { attributes: { class: "tiptap" } },
    onUpdate: () => touch(),
  });
  const enEditor = useEditor({
    extensions: enExtensions,
    content: enBody,
    shouldRerenderOnTransaction: false,
    editorProps: { attributes: { class: "tiptap" } },
    onUpdate: () => touch(),
  });
  const { modals: zhModals, state: zhModalState } = useModalState();
  const { modals: enModals, state: enModalState } = useModalState();
  const langEditors = useMemo(
    () => ({
      zh: { editor: zhEditor, modals: zhModals, state: zhModalState },
      en: { editor: enEditor, modals: enModals, state: enModalState },
    }),
    [zhEditor, enEditor, zhModals, enModals, zhModalState, enModalState],
  );

  const activeLang = enabledLangs[activeLangIdx] || "zh";
  const activeEntry = langEditors[activeLang as "zh" | "en"];

  useEffect(() => {
    zhEditor?.setEditable(viewLayout === "split" || activeLang === "zh");
    enEditor?.setEditable(viewLayout === "split" || activeLang === "en");
  }, [zhEditor, enEditor, activeLang, viewLayout]);

  useEffect(() => {
    if (zhEditor && zhBody && zhBody !== zhEditor.getHTML()) {
      zhEditor.commands.setContent(zhBody, { emitUpdate: false });
    }
  }, [zhBody, zhEditor]);
  useEffect(() => {
    if (enEditor && enBody && enBody !== enEditor.getHTML()) {
      enEditor.commands.setContent(enBody, { emitUpdate: false });
    }
  }, [enBody, enEditor]);

  useOutsideClick(langMenuRef, showLangMenu, () => setShowLangMenu(false));
  useSlashMediaBridge(activeEntry?.state);

  // Auto-enable English in split layout
  useEffect(() => {
    if (viewLayout === "split" && !enabledLangs.includes("en")) {
      setEnabledLangs((prev) => (prev.includes("en") ? prev : [...prev, "en"]));
    }
  }, [viewLayout, enabledLangs]);

  // ── Load ──
  useEffect(() => {
    void (async () => {
      try {
        const [cats, tgs] = await Promise.all([getCategories(), getTags()]);
        setCategories(cats || []);
        setTags(tgs || []);
      } catch { /* non-critical */ }
    })();
  }, []);

  const loadArticle = useCallback(async () => {
    if (!id || loadedIdRef.current === id) return;
    pauseReady();
    setLoading(true);
    setError(null);
    try {
      const article = await getAdminArticle(Number(id));
      setZhTitle(article.zhTitle || "");
      setEnTitle(article.enTitle || "");
      setSlug(article.slug || "");
      if (article.categoryIds?.length) setSelectedCategoryIds(article.categoryIds);
      else if (article.categories?.length) setSelectedCategoryIds(article.categories.map((c) => c.id));
      else if (article.categoryId) setSelectedCategoryIds([article.categoryId]);
      setSelectedTagIds(article.tags?.map((t) => t.id) || []);
      setCoverImage(article.coverImage || "");
      setZhBody(article.zhBody || "");
      setEnBody(article.enBody || "");
      setZhSeoTitle(article.zhSeoTitle || "");
      setEnSeoTitle(article.enSeoTitle || "");
      setZhMetaDescription(article.zhMetaDescription || "");
      setEnMetaDescription(article.enMetaDescription || "");
      setOgImage(article.ogImage || "");
      setAuthor(article.author || "");
      setAutoSummary(article.autoSummary || false);
      setAllowComments(article.allowComments !== false);
      setPinned(article.pinned || false);
      setVisibility(article.visibility || "public");
      setMetadata(article.metadata || {});
      setArticleCreatedAt(article.createdAt || null);
      setArticlePublishedAt(article.publishedAt || null);
      setArticleStatus(article.status || "draft");
      setMarkdownContent({ zh: "", en: "" });
      setEditorMode("richtext");
      if (article.enBody || article.enTitle) setEnabledLangs(["zh", "en"]);
      loadedIdRef.current = id;
      setBaseUpdatedAt(article.updatedAt || null);
      setConflict(null);
      markHydrated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load article");
    } finally {
      setLoading(false);
      resumeReady();
    }
  }, [id, pauseReady, resumeReady, markHydrated]);

  useEffect(() => {
    if (!id) {
      loadedIdRef.current = null;
      setLoading(false);
      readyRef.current = true;
      return;
    }
    if (loadedIdRef.current !== id) void loadArticle();
  }, [id, loadArticle, readyRef]);

  // ── Bodies / payload ──
  const resolveBodies = useCallback(() => {
    if (editorMode === "markdown") {
      const zhHtml = markdownToHtml(markdownContent.zh ?? "");
      const enHtml = markdownToHtml(markdownContent.en ?? "");
      zhEditor?.commands.setContent(zhHtml, { emitUpdate: false });
      enEditor?.commands.setContent(enHtml, { emitUpdate: false });
      const normalizedZh = zhEditor?.getHTML() || zhHtml;
      const normalizedEn = enEditor?.getHTML() || enHtml;
      setZhBody(normalizedZh);
      setEnBody(normalizedEn);
      return { zhBody: normalizedZh, enBody: normalizedEn };
    }
    return {
      zhBody: zhEditor?.getHTML() || zhBody || "",
      enBody: enEditor?.getHTML() || enBody || "",
    };
  }, [editorMode, markdownContent, zhEditor, enEditor, zhBody, enBody]);

  const buildPayload = useCallback((status: "draft" | "published", publishedAt?: string): Record<string, unknown> => {
    const bodies = resolveBodies();
    const finalSlug = slug.trim() || slugifyTitle(zhTitle);
    const payload: Record<string, unknown> = {
      zhTitle, enTitle, slug: finalSlug, coverImage,
      zhBody: bodies.zhBody, enBody: bodies.enBody,
      zhSeoTitle, enSeoTitle, zhMetaDescription, enMetaDescription, ogImage,
      status, categoryIds: selectedCategoryIds, tagIds: selectedTagIds,
      author, autoSummary, allowComments, pinned, visibility, metadata,
    };
    if (status === "published") payload.publishedAt = publishedAt ?? new Date().toISOString();
    return payload;
  }, [
    resolveBodies, slug, zhTitle, enTitle, coverImage,
    zhSeoTitle, enSeoTitle, zhMetaDescription, enMetaDescription, ogImage,
    selectedCategoryIds, selectedTagIds, author, autoSummary, allowComments,
    pinned, visibility, metadata,
  ]);

  const handleSave = useCallback(async (
    intent: "draft" | "publish" | "autosave" = "draft",
    opts?: { force?: boolean },
  ) => {
    if (savingRef.current) return;
    if (!zhTitle.trim()) {
      if (intent !== "autosave") setError("请填写中文标题");
      return;
    }
    const finalSlug = slug.trim() || slugifyTitle(zhTitle);
    if (!slug.trim()) setSlug(finalSlug);

    const status = resolveSaveStatus(intent, articleStatus);
    const silent = intent === "autosave";
    const force = !!opts?.force;

    savingRef.current = true;
    setSaving(true);
    markSaving();
    if (!silent) {
      setError(null);
      setSuccessMessage("");
    }
    try {
      const payload = buildPayload(status);
      const articleId = id ? Number(id) : loadedIdRef.current ? Number(loadedIdRef.current) : null;
      if (articleId) {
        const updated = await updateArticle(articleId, payload as Partial<Article>, {
          baseUpdatedAt: force ? null : baseUpdatedAt,
          force,
        });
        if (updated?.updatedAt) setBaseUpdatedAt(updated.updatedAt);
        setConflict(null);
        setArticleStatus(
          status === "published" ? "published" : articleStatus === "scheduled" ? "scheduled" : status,
        );
        if (!silent) toast(setSuccessMessage, intent === "publish" ? "已发布" : force ? "已强制覆盖保存" : "已保存");
      } else {
        const created = await createArticle(payload as Partial<Article>);
        setArticleStatus(status === "published" ? "published" : "draft");
        loadedIdRef.current = String(created.id);
        if (created.updatedAt) setBaseUpdatedAt(created.updatedAt);
        if (!silent) toast(setSuccessMessage, intent === "publish" ? "已创建并发布" : "已保存");
        navigate(`/admin/articles/edit/${created.id}`, { replace: true });
      }
      markClean({ autosave: silent });
    } catch (err: unknown) {
      const conf = isArticleVersionConflict(err);
      if (conf.conflict) {
        markError();
        setConflict({ serverUpdatedAt: conf.currentUpdatedAt });
        if (!silent) setError(conf.message || "保存冲突：文章已被他人修改");
        return;
      }
      markError();
      const ax = err as { response?: { data?: { error?: { message?: string } } } };
      const msg = ax?.response?.data?.error?.message;
      setError(msg || (err instanceof Error ? err.message : silent ? "自动保存失败，请手动保存" : "保存失败"));
    } finally {
      savingRef.current = false;
      setSaving(false);
    }
  }, [
    zhTitle, slug, articleStatus, buildPayload, id, navigate, baseUpdatedAt,
    savingRef, markSaving, markClean, markError,
  ]);

  useEffect(() => {
    if (!isDirty || saving || loading || !zhTitle.trim()) return;
    const t = window.setTimeout(() => void handleSave("autosave"), AUTOSAVE_DEBOUNCE_MS);
    return () => window.clearTimeout(t);
  }, [isDirty, saving, loading, zhTitle, handleSave]);

  const openPreview = useCallback(() => {
    const bodies = resolveBodies();
    const title = activeLang === "en" ? (enTitle || zhTitle) : (zhTitle || enTitle);
    const bodyHtml = activeLang === "en" ? (bodies.enBody || bodies.zhBody) : (bodies.zhBody || bodies.enBody);
    const finalSlug = slug.trim() || slugifyTitle(zhTitle);
    setPreviewData({
      title,
      bodyHtml,
      coverImage: coverImage || undefined,
      author: author || undefined,
      langLabel: activeLang === "en" ? "English" : "中文",
      statusLabel: statusLabelOf(articleStatus, isDirty),
      publicPath: articleStatus === "published" && finalSlug ? `/blog/${finalSlug}` : null,
      metadata,
    });
    setShowPreview(true);
  }, [resolveBodies, activeLang, enTitle, zhTitle, articleStatus, slug, coverImage, author, isDirty, metadata]);

  useEditorShortcuts({
    canPublish,
    onSave: (intent) => void handleSave(intent),
    onPreview: openPreview,
  });

  const { confirmLeave } = useUnsavedChangesGuard(isDirty, LEAVE_UNSAVED_MESSAGE);
  const handleBack = useCallback(() => {
    if (!confirmLeave()) return;
    navigate("/admin/articles");
  }, [confirmLeave, navigate]);

  const schedule = useArticleSchedule({
    id, isEditing, canPublish, zhTitle, slug, setSlug,
    articleStatus, setArticleStatus, buildPayload, navigate, setError,
  });

  const wordStats = useWordStats({
    editorMode, markdownContent, zhBody, enBody, zhEditor, enEditor,
    tick: `${isDirty}-${saving}`,
  });

  const buildCurrentDraftSnapshot = useCallback((): ArticleDraftSnapshot => {
    const bodies = resolveBodies();
    return {
      zhTitle, enTitle,
      slug: slug.trim() || slugifyTitle(zhTitle),
      status: articleStatus,
      zhBody: bodies.zhBody, enBody: bodies.enBody,
      coverImage, zhSeoTitle, enSeoTitle, zhMetaDescription, enMetaDescription, ogImage, author,
    };
  }, [
    resolveBodies, zhTitle, enTitle, slug, articleStatus, coverImage,
    zhSeoTitle, enSeoTitle, zhMetaDescription, enMetaDescription, ogImage, author,
  ]);

  const applyBodiesToEditors = useCallback((nextZh: string, nextEn: string) => {
    setZhBody(nextZh);
    setEnBody(nextEn);
    zhEditor?.commands.setContent(nextZh || "", { emitUpdate: false });
    enEditor?.commands.setContent(nextEn || "", { emitUpdate: false });
    if (editorMode === "markdown") {
      setMarkdownContent({
        zh: htmlToMarkdown(nextZh || ""),
        en: htmlToMarkdown(nextEn || ""),
      });
    }
  }, [zhEditor, enEditor, editorMode]);

  const handleRestoreVersion = useCallback((snapshot: ArticleDraftSnapshot) => {
    pauseReady();
    setZhTitle(typeof snapshot.zhTitle === "string" ? snapshot.zhTitle : "");
    setEnTitle(typeof snapshot.enTitle === "string" ? snapshot.enTitle : "");
    setSlug(typeof snapshot.slug === "string" ? snapshot.slug : slug);
    setCoverImage(typeof snapshot.coverImage === "string" ? snapshot.coverImage : "");
    setAuthor(typeof snapshot.author === "string" ? snapshot.author : author);
    if (typeof snapshot.zhSeoTitle === "string") setZhSeoTitle(snapshot.zhSeoTitle);
    if (typeof snapshot.enSeoTitle === "string") setEnSeoTitle(snapshot.enSeoTitle);
    if (typeof snapshot.zhMetaDescription === "string") setZhMetaDescription(snapshot.zhMetaDescription);
    if (typeof snapshot.enMetaDescription === "string") setEnMetaDescription(snapshot.enMetaDescription);
    if (typeof snapshot.ogImage === "string") setOgImage(snapshot.ogImage);
    applyBodiesToEditors(
      typeof snapshot.zhBody === "string" ? snapshot.zhBody : "",
      typeof snapshot.enBody === "string" ? snapshot.enBody : "",
    );
    setShowVersionHistory(false);
    toast(setSuccessMessage, "已恢复到所选版本（尚未保存，请检查后保存）", 4000);
    resumeReady();
    touch();
  }, [slug, author, applyBodiesToEditors, pauseReady, resumeReady, touch]);

  const handleCopyToOtherLang = useCallback((from: "zh" | "en") => {
    const to = from === "zh" ? "en" : "zh";
    if (editorMode === "markdown") {
      setMarkdownContent((prev) => ({ ...prev, [to]: prev[from] ?? "" }));
    } else {
      const srcEd = from === "zh" ? zhEditor : enEditor;
      const dstEd = to === "zh" ? zhEditor : enEditor;
      const html = srcEd?.getHTML() || (from === "zh" ? zhBody : enBody) || "";
      dstEd?.commands.setContent(html, { emitUpdate: false });
      if (to === "zh") setZhBody(html);
      else setEnBody(html);
    }
    if (from === "zh" && zhTitle.trim()) setEnTitle(zhTitle);
    if (from === "en" && enTitle.trim()) setZhTitle(enTitle);
    touch();
    toast(setSuccessMessage, from === "zh" ? "已复制中文到英文（未保存）" : "已复制英文到中文（未保存）", 2500);
  }, [editorMode, zhEditor, enEditor, zhBody, enBody, zhTitle, enTitle, touch]);

  const handleTranslateToOtherLang = useCallback(async (from: "zh" | "en") => {
    const to = from === "zh" ? "en" : "zh";
    setTranslateBusy(true);
    setError(null);
    try {
      const srcTitle = from === "zh" ? zhTitle : enTitle;
      let srcBodyPlain: string;
      if (editorMode === "markdown") {
        srcBodyPlain = markdownContent[from] ?? "";
      } else {
        const srcEd = from === "zh" ? zhEditor : enEditor;
        srcBodyPlain = htmlToPlainText(srcEd?.getHTML() || (from === "zh" ? zhBody : enBody) || "");
      }
      if (!srcTitle.trim() && !srcBodyPlain.trim()) {
        setError("源语言内容为空，无法翻译");
        return;
      }
      const sourceLang = from;
      const targetLang = to;
      if (srcTitle.trim()) {
        const tr = await translateText({ text: srcTitle, sourceLang, targetLang });
        if (to === "zh") setZhTitle(tr.translatedText);
        else setEnTitle(tr.translatedText);
      }
      if (srcBodyPlain.trim()) {
        const tr = await translateText({ text: srcBodyPlain, sourceLang, targetLang });
        if (editorMode === "markdown") {
          setMarkdownContent((prev) => ({ ...prev, [to]: tr.translatedText }));
        } else {
          const html = plainTextToHtml(tr.translatedText);
          const dstEd = to === "zh" ? zhEditor : enEditor;
          dstEd?.commands.setContent(html, { emitUpdate: false });
          if (to === "zh") setZhBody(html);
          else setEnBody(html);
        }
      }
      if (!enabledLangs.includes(to)) {
        setEnabledLangs((prev) => (prev.includes(to) ? prev : [...prev, to]));
      }
      touch();
      toast(setSuccessMessage, from === "zh" ? "已翻译到英文（未保存，请校对）" : "已翻译到中文（未保存，请校对）");
    } catch (err: unknown) {
      const ax = err as { response?: { data?: { error?: { message?: string } | string } } };
      const msg = (ax?.response?.data?.error as { message?: string })?.message
        ?? (typeof ax?.response?.data?.error === "string" ? ax.response.data.error : undefined);
      setError(msg || (err instanceof Error ? err.message : "翻译失败（请检查 AI/翻译配置）"));
    } finally {
      setTranslateBusy(false);
    }
  }, [zhTitle, enTitle, editorMode, markdownContent, zhEditor, enEditor, zhBody, enBody, enabledLangs, touch]);

  const handleModeChange = useCallback((newMode: "richtext" | "markdown") => {
    if (newMode === editorMode) return;
    const zhHtml = editorMode === "markdown"
      ? markdownToHtml(markdownContent.zh ?? "")
      : (zhEditor?.getHTML() || zhBody || "");
    const enHtml = editorMode === "markdown"
      ? markdownToHtml(markdownContent.en ?? "")
      : (enEditor?.getHTML() || enBody || "");
    if ((hasMeaningfulHtml(zhHtml) || hasMeaningfulHtml(enHtml) || isDirty)
      && !window.confirm(MODE_SWITCH_MESSAGE)) {
      return;
    }
    if (newMode === "markdown") {
      setMarkdownContent({
        zh: htmlToMarkdown(zhEditor?.getHTML() || zhBody || ""),
        en: htmlToMarkdown(enEditor?.getHTML() || enBody || ""),
      });
    } else {
      for (const lang of ["zh", "en"] as const) {
        const ed = lang === "zh" ? zhEditor : enEditor;
        const html = markdownToHtml(markdownContent[lang] ?? "");
        ed?.commands.setContent(html, { emitUpdate: false });
        const normalized = ed?.getHTML() || html;
        if (lang === "zh") setZhBody(normalized);
        else setEnBody(normalized);
      }
      setMarkdownApi(null);
    }
    setEditorMode(newMode);
    touch();
  }, [editorMode, markdownContent, zhEditor, enEditor, zhBody, enBody, isDirty, touch]);

  const handleApplyTemplate = useCallback((tpl: ArticleTemplate) => {
    if (tpl.id !== "blank") {
      const hasContent =
        zhTitle.trim() || enTitle.trim()
        || (zhEditor?.getText() || "").trim() || (enEditor?.getText() || "").trim()
        || (markdownContent.zh || "").trim() || (markdownContent.en || "").trim();
      if (hasContent && !window.confirm(`应用模板「${tpl.name}」将覆盖当前标题与正文，是否继续？`)) {
        return;
      }
    }
    pauseReady();
    if (tpl.zhTitle) setZhTitle(tpl.zhTitle);
    if (tpl.enTitle) setEnTitle(tpl.enTitle);
    applyBodiesToEditors(tpl.zhBody || "<p></p>", tpl.enBody || "<p></p>");
    if (tpl.enTitle || tpl.enBody) {
      setEnabledLangs((prev) => (prev.includes("en") ? prev : [...prev, "en"]));
    }
    setShowTemplatePicker(false);
    toast(setSuccessMessage, tpl.id === "blank" ? "已清空为空白文档" : `已应用模板「${tpl.name}」（未保存）`);
    resumeReady();
    touch();
  }, [zhTitle, enTitle, zhEditor, enEditor, markdownContent, applyBodiesToEditors, pauseReady, resumeReady, touch]);

  const toggleCategory = (catId: number) => {
    setSelectedCategoryIds((prev) => (prev.includes(catId) ? prev.filter((i) => i !== catId) : [...prev, catId]));
    touch();
  };
  const toggleTag = (tagId: number) => {
    setSelectedTagIds((prev) => (prev.includes(tagId) ? prev.filter((i) => i !== tagId) : [...prev, tagId]));
    touch();
  };
  const addLang = (langKey: string) => {
    if (!enabledLangs.includes(langKey)) {
      const next = [...enabledLangs, langKey];
      setEnabledLangs(next);
      setActiveLangIdx(next.length - 1);
    }
    setShowLangMenu(false);
  };
  const removeLang = (langKey: string) => {
    if (langKey === "zh") return;
    const next = enabledLangs.filter((l) => l !== langKey);
    setEnabledLangs(next);
    if (activeLangIdx >= next.length) setActiveLangIdx(next.length - 1);
  };
  const selectLangKey = (lang: string) => {
    const idx = enabledLangs.indexOf(lang);
    if (idx >= 0) setActiveLangIdx(idx);
  };

  const sidebarArticle = useMemo(
    () => (isEditing ? { slug, author, createdAt: articleCreatedAt, publishedAt: articlePublishedAt } : null),
    [isEditing, slug, author, articleCreatedAt, articlePublishedAt],
  );

  const langTitleMap = useMemo(
    () => ({
      zh: { title: zhTitle, setTitle: track(setZhTitle), placeholder: "输入中文标题" },
      en: { title: enTitle, setTitle: track(setEnTitle), placeholder: "Enter English title" },
    }),
    [zhTitle, enTitle, track],
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-gray-600">加载中...</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full min-h-0 bg-white">
      <div className="flex-shrink-0 z-20 bg-white border-b border-gray-200 shadow-sm">
        {/* Action bar */}
        <div className="flex items-center gap-3 px-4 py-2">
          <button type="button" onClick={handleBack} className="text-gray-500 hover:text-gray-700 text-sm flex-shrink-0">
            &larr; 返回
          </button>
          <input
            type="text"
            value={langTitleMap[activeLang as "zh" | "en"]?.title || ""}
            onChange={(e) => langTitleMap[activeLang as "zh" | "en"]?.setTitle(e.target.value)}
            className="flex-1 px-3 py-1.5 text-base font-semibold border border-transparent rounded-lg hover:border-gray-300 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none transition-colors bg-transparent"
            placeholder={langTitleMap[activeLang as "zh" | "en"]?.placeholder || "标题"}
          />
          <div className="flex items-center gap-1.5 flex-shrink-0">
            <SaveStatusBadge phase={savePhase} lastSavedAt={lastSavedAt} isAutosave={lastSaveWasAutosave} />
            <PopoverButton
              label="基本信息"
              active={showBasicInfo}
              onClick={() => { setShowBasicInfo(!showBasicInfo); setShowSeo(false); setShowAdvanced(false); }}
            />
            <PopoverButton
              label="SEO"
              active={showSeo}
              onClick={() => { setShowSeo(!showSeo); setShowBasicInfo(false); setShowAdvanced(false); }}
            />
            <PopoverButton
              label="高级"
              active={showAdvanced}
              onClick={() => { setShowAdvanced(!showAdvanced); setShowBasicInfo(false); setShowSeo(false); }}
            />
            {isEditing && (
              <PopoverButton
                label="历史版本"
                active={showVersionHistory}
                onClick={() => {
                  setVersionDraftSnapshot(buildCurrentDraftSnapshot());
                  setShowVersionHistory(true);
                }}
              />
            )}
            <button
              type="button"
              onClick={() => setShowTemplatePicker(true)}
              title="应用文章结构模板"
              className="px-2.5 py-1.5 text-xs border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700"
            >
              模板
            </button>
            <button
              type="button"
              onClick={openPreview}
              title="预览 (⌘P / Ctrl+P)"
              className="px-2.5 py-1.5 text-xs border border-gray-300 rounded-lg hover:bg-gray-50 text-gray-700"
            >
              预览
            </button>
            <span className="w-px h-6 bg-gray-200 mx-1" />
            <ScheduledPublicationPanel
              compact
              item={schedule.scheduledPublication}
              loading={schedule.scheduleLoading}
              busy={schedule.scheduleBusy}
              canPublish={canPublish}
              disabledReason="需要 articles:publish 权限才能安排定时发布。"
              onSchedule={schedule.handleSchedulePublish}
              onCancel={schedule.handleCancelSchedule}
              onRetry={schedule.handleRetrySchedule}
              onRefresh={schedule.loadArticleSchedule}
              title={articleStatus === "published" ? "定时更新" : "定时"}
            />
            <span className="w-px h-6 bg-gray-200 mx-1" />
            <button
              type="button"
              onClick={() => void handleSave("draft")}
              disabled={saving}
              title="保存 (⌘S / Ctrl+S)"
              className="px-3 py-1.5 text-sm border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50"
            >
              {saving ? "保存中..." : "保存"}
            </button>
            {canPublish && (
              <button
                type="button"
                onClick={() => void handleSave("publish")}
                disabled={saving}
                title="发布 (⌘⇧S / Ctrl+Shift+S)"
                className="px-3 py-1.5 text-sm bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
              >
                {saving ? "发布中..." : "发布"}
              </button>
            )}
          </div>
        </div>

        {showBasicInfo && (
          <ArticleForm
            slug={slug} setSlug={track(setSlug)}
            author={author} setAuthor={track(setAuthor)}
            coverImage={coverImage} setCoverImage={track(setCoverImage)}
            showCoverPicker={showCoverPicker} setShowCoverPicker={setShowCoverPicker}
            categories={categories} selectedCategoryIds={selectedCategoryIds} toggleCategory={toggleCategory}
            tags={tags} selectedTagIds={selectedTagIds} toggleTag={toggleTag}
          />
        )}
        {showSeo && (
          <SeoFieldsPanel
            zhSeoTitle={zhSeoTitle} setZhSeoTitle={track(setZhSeoTitle)}
            enSeoTitle={enSeoTitle} setEnSeoTitle={track(setEnSeoTitle)}
            zhMetaDescription={zhMetaDescription} setZhMetaDescription={track(setZhMetaDescription)}
            enMetaDescription={enMetaDescription} setEnMetaDescription={track(setEnMetaDescription)}
            ogImage={ogImage} setOgImage={track(setOgImage)}
          />
        )}
        {showAdvanced && (
          <AdvancedSettingsPanel
            visibility={visibility} setVisibility={track(setVisibility)}
            autoSummary={autoSummary} setAutoSummary={track(setAutoSummary)}
            allowComments={allowComments} setAllowComments={track(setAllowComments)}
            pinned={pinned} setPinned={track(setPinned)}
            metadata={metadata} setMetadata={track(setMetadata)}
          />
        )}

        <EditorLangBar
          enabledLangs={enabledLangs}
          activeLangIdx={activeLangIdx}
          viewLayout={viewLayout}
          wordStats={wordStats}
          editorMode={editorMode}
          translateBusy={translateBusy}
          showLangMenu={showLangMenu}
          langMenuRef={langMenuRef}
          onSelectLang={setActiveLangIdx}
          onRemoveLang={removeLang}
          onAddLang={addLang}
          onToggleLangMenu={() => setShowLangMenu((v) => !v)}
          onToggleSplit={() => setViewLayout((v) => (v === "split" ? "focus" : "split"))}
          onCopyZhToEn={() => handleCopyToOtherLang("zh")}
          onTranslateZhToEn={() => void handleTranslateToOtherLang("zh")}
          onModeChange={handleModeChange}
        />

        <div className="flex items-stretch border-t border-gray-200 bg-gray-50">
          {editorMode === "richtext" && activeEntry?.editor ? (
            <div className="flex-1 min-w-0 overflow-x-auto">
              <EditorToolbar editor={activeEntry.editor} modals={activeEntry.modals} />
            </div>
          ) : editorMode === "markdown" ? (
            <div className="flex-1 min-w-0 overflow-x-auto">
              <MarkdownToolbar api={markdownApi} />
            </div>
          ) : (
            <div className="flex-1 py-2" />
          )}
        </div>
      </div>

      <EditorMessageBars
        error={error}
        onClearError={() => setError(null)}
        successMessage={successMessage}
        scheduleMessage={schedule.scheduleMessage}
        onClearSuccess={() => {
          schedule.setScheduleMessage("");
          setSuccessMessage("");
        }}
      />

      <EditorWorkspace
        viewLayout={viewLayout}
        editorMode={editorMode}
        enabledLangs={enabledLangs}
        activeLang={activeLang}
        activeLangIdx={activeLangIdx}
        langEditors={langEditors}
        langTitleMap={langTitleMap}
        wordStats={wordStats}
        markdownContent={markdownContent}
        metadata={metadata}
        sidebarArticle={sidebarArticle}
        onSelectLangKey={selectLangKey}
        onMarkdownChange={(lang, val) => {
          setMarkdownContent((prev) => ({ ...prev, [lang]: val }));
          touch();
        }}
        onMarkdownApiReady={setMarkdownApi}
      />

      {Object.entries(langEditors).map(([lang, entry]) =>
        entry.editor ? <EditorModals key={lang} editor={entry.editor} state={entry.state} /> : null,
      )}

      <ImagePickerModal
        open={showCoverPicker}
        onClose={() => setShowCoverPicker(false)}
        onSelect={(item) => {
          setCoverImage(item.url);
          setShowCoverPicker(false);
          touch();
        }}
      />

      {showVersionHistory && isEditing && (
        <ArticleVersionHistoryPanel
          articleId={Number(id)}
          onClose={() => {
            setShowVersionHistory(false);
            setVersionDraftSnapshot(null);
          }}
          currentDraft={versionDraftSnapshot}
          onRestore={handleRestoreVersion}
          canRestore
        />
      )}

      <ArticlePreviewModal open={showPreview} data={previewData} onClose={() => setShowPreview(false)} />
      <TemplatePickerModal
        open={showTemplatePicker}
        onClose={() => setShowTemplatePicker(false)}
        onSelect={handleApplyTemplate}
      />

      {conflict && (
        <ArticleConflictDialog
          serverUpdatedAt={conflict.serverUpdatedAt}
          busy={saving}
          onDismiss={() => setConflict(null)}
          onReload={() => {
            setConflict(null);
            loadedIdRef.current = null;
            pauseReady();
            void loadArticle();
          }}
          onForceOverwrite={() => {
            setConflict(null);
            void handleSave("draft", { force: true });
          }}
        />
      )}
    </div>
  );
}
