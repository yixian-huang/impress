import { useState, useEffect, useCallback } from "react";
import { listMedia, deleteMedia, uploadMedia, getMediaUsages, renameMedia } from "@/api/media";
import type { MediaItem, MediaUsage } from "@/api/media";
import ImageCropUpload from "@/components/admin/ImageCropUpload";
import RecropModal from "@/components/admin/RecropModal";
import {
  AdminButton,
  AdminCard,
  AdminEmptyState,
  AdminErrorBanner,
  AdminInfoBanner,
  AdminInput,
  AdminLoading,
  AdminModal,
  AdminPageHeader,
  AdminPagination,
  AdminSuccessBanner,
  useAdminConfirm,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { invalidateAdminQueryPrefix, useAdminQuery } from "@/lib/adminQuery";
import { adminQueryKeys } from "@/lib/adminQueryKeys";

export default function MediaPage() {
  useDocumentTitle("媒体管理");
  const { confirm, confirmDialog } = useAdminConfirm();
  const [actionError, setActionError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [deleting, setDeleting] = useState<number | null>(null);
  const [showUpload, setShowUpload] = useState(false);
  const [cropItem, setCropItem] = useState<MediaItem | null>(null);
  const [usageItem, setUsageItem] = useState<MediaItem | null>(null);
  const [usages, setUsages] = useState<MediaUsage[]>([]);
  const [loadingUsages, setLoadingUsages] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [isDragging, setIsDragging] = useState(false);
  const [uploadingPaste, setUploadingPaste] = useState(false);
  const [uploadSuccess, setUploadSuccess] = useState(false);
  const pageSize = 20;

  const { data, error, loading, isFetching, refetch } = useAdminQuery(
    [...adminQueryKeys.media, page, pageSize],
    () => listMedia(page, pageSize),
  );

  const items = data?.items ?? [];
  const total = data?.total ?? 0;
  const displayError = actionError ?? (error ? error.message : null);

  const loadMedia = useCallback(async () => {
    setActionError(null);
    invalidateAdminQueryPrefix(adminQueryKeys.media);
    invalidateAdminQueryPrefix(adminQueryKeys.dashboardStats);
    await refetch({ force: true });
  }, [refetch]);

  const handleFilesUpload = useCallback(async (files: File[]) => {
    if (files.length === 0) return;
    setUploadingPaste(true);
    setUploadSuccess(false);
    setActionError(null);
    try {
      for (const file of files) {
        await uploadMedia(file);
      }
      setUploadSuccess(true);
      await loadMedia();
      setTimeout(() => setUploadSuccess(false), 3000);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "上传失败");
    } finally {
      setUploadingPaste(false);
    }
  }, [loadMedia]);

  // Paste listener
  useEffect(() => {
    const handlePaste = (e: ClipboardEvent) => {
      const files: File[] = [];
      if (e.clipboardData?.items) {
        for (const item of e.clipboardData.items) {
          if (item.type.startsWith("image/")) {
            const file = item.getAsFile();
            if (file) files.push(file);
          }
        }
      }
      if (files.length > 0) {
        e.preventDefault();
        handleFilesUpload(files);
      }
    };
    document.addEventListener("paste", handlePaste);
    return () => document.removeEventListener("paste", handlePaste);
  }, [handleFilesUpload]);

  // Drag event handlers
  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.currentTarget === e.target || !e.currentTarget.contains(e.relatedTarget as Node)) {
      setIsDragging(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    const files = Array.from(e.dataTransfer.files);
    if (files.length > 0) {
      handleFilesUpload(files);
    }
  };

  const handleDelete = async (item: MediaItem) => {
    const ok = await confirm({
      title: "删除媒体",
      message: `确定要删除「${item.filename}」吗？此操作不可撤销。`,
      confirmLabel: "删除",
      danger: true,
    });
    if (!ok) return;

    setDeleting(item.id);
    setActionError(null);
    try {
      await deleteMedia(item.id);
      await loadMedia();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "删除失败");
    } finally {
      setDeleting(null);
    }
  };

  const handleCopyUrl = (url: string) => {
    navigator.clipboard.writeText(url).catch(() => {
      const input = document.createElement("input");
      input.value = url;
      document.body.appendChild(input);
      input.select();
      document.execCommand("copy");
      document.body.removeChild(input);
    });
  };

  const handleUpload = () => {
    setShowUpload(false);
    loadMedia();
  };

  const handleDirectUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setActionError(null);
    try {
      await uploadMedia(file);
      await loadMedia();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "上传失败");
    }
    e.target.value = "";
  };

  const handleShowUsages = async (item: MediaItem) => {
    setUsageItem(item);
    setLoadingUsages(true);
    try {
      const data = await getMediaUsages(item.id);
      setUsages(data);
    } catch {
      setUsages([]);
    } finally {
      setLoadingUsages(false);
    }
  };

  const handleRename = async (item: MediaItem, newName: string) => {
    const trimmed = newName.trim();
    if (!trimmed || trimmed === item.filename) {
      setEditingId(null);
      return;
    }
    try {
      await renameMedia(item.id, trimmed);
      await loadMedia();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "重命名失败");
    }
    setEditingId(null);
  };

  const totalPages = Math.ceil(total / pageSize);

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  const usageTypeLabel: Record<string, string> = {
    article: "文章",
    page: "页面",
    content_document: "内容",
  };

  return (
    <div
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      className="relative"
    >
      {confirmDialog}
      {/* Drag overlay */}
      {isDragging && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-blue-500/10 backdrop-blur-sm">
          <div className="border-4 border-dashed border-blue-500 rounded-2xl p-16 bg-white/80">
            <p className="text-xl font-semibold text-blue-600">松开鼠标上传文件</p>
          </div>
        </div>
      )}

      <AdminPageHeader
        title="媒体管理"
        description={
          loading
            ? "支持拖拽与粘贴上传"
            : `共 ${total} 个文件 · 支持拖拽与粘贴上传${isFetching ? " · 刷新中" : ""}`
        }
        actions={
          <>
            <label className="inline-flex h-8 cursor-pointer items-center justify-center gap-1.5 rounded-lg border border-slate-200/90 bg-white px-3 text-xs font-medium text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50">
              直接上传
              <input
                type="file"
                accept="image/*,video/*,audio/*"
                onChange={handleDirectUpload}
                className="hidden"
              />
            </label>
            <AdminButton size="sm" onClick={() => setShowUpload(!showUpload)}>
              裁剪上传
            </AdminButton>
          </>
        }
      />

      <p className="mb-4 text-xs text-slate-500">
        提示：可直接粘贴 (Ctrl+V) 或拖放图片到页面上传
      </p>

      {uploadingPaste && (
        <AdminInfoBanner
          message={
            <>
              <svg className="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24" aria-hidden>
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              正在上传…
            </>
          }
        />
      )}

      {uploadSuccess && <AdminSuccessBanner message="上传成功" />}

      {displayError && (
        <AdminErrorBanner message={displayError} onDismiss={() => setActionError(null)} />
      )}

      {showUpload && (
        <AdminCard className="mb-6" title="裁剪上传">
          <ImageCropUpload onUpload={handleUpload} />
        </AdminCard>
      )}

      {loading ? (
        <AdminLoading />
      ) : items.length === 0 ? (
        <AdminEmptyState
          title="暂无媒体文件"
          description="点击上传按钮、粘贴 (Ctrl+V) 或拖放文件到此处"
        />
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
            {items.map((item) => (
              <div
                key={item.id}
                className="group relative overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.04)]"
              >
                <div className="flex aspect-square items-center justify-center overflow-hidden bg-slate-100">
                  {item.mimeType.startsWith("image/") ? (
                    <img
                      src={item.url}
                      alt={item.filename}
                      className="h-full w-full object-cover"
                      loading="lazy"
                    />
                  ) : (
                    <div className="flex flex-col items-center gap-2 text-slate-400">
                      <span className="text-4xl">
                        {item.mimeType.startsWith("video/")
                          ? "🎬"
                          : item.mimeType.startsWith("audio/")
                            ? "🎵"
                            : "📄"}
                      </span>
                      <span className="text-xs uppercase text-slate-500">
                        {item.mimeType.split("/")[1]}
                      </span>
                    </div>
                  )}
                </div>

                <div className="p-3">
                  {editingId === item.id ? (
                    <AdminInput
                      type="text"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      onBlur={() => handleRename(item, editName)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") handleRename(item, editName);
                        if (e.key === "Escape") setEditingId(null);
                      }}
                      className="px-2 py-1 text-xs"
                      autoFocus
                    />
                  ) : (
                    <p
                      className="cursor-pointer truncate text-xs font-medium text-slate-700 hover:text-blue-600"
                      title="点击重命名"
                      onClick={() => {
                        setEditingId(item.id);
                        setEditName(item.filename);
                      }}
                    >
                      {item.filename}
                    </p>
                  )}
                  <p className="mt-1 text-xs text-slate-500">
                    {formatFileSize(item.size)}
                    {item.width && item.height && ` · ${item.width}×${item.height}`}
                  </p>
                  <p className="mt-0.5 text-xs text-slate-400">
                    {new Date(item.createdAt).toLocaleDateString("zh-CN")}
                  </p>
                </div>

                <div className="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-slate-950/45 opacity-0 transition-opacity group-hover:opacity-100">
                  <div className="flex items-center gap-2">
                    <AdminButton size="sm" variant="secondary" onClick={() => handleCopyUrl(item.url)}>
                      复制 URL
                    </AdminButton>
                    <AdminButton size="sm" variant="secondary" onClick={() => setCropItem(item)}>
                      裁剪
                    </AdminButton>
                  </div>
                  <div className="flex items-center gap-2">
                    <AdminButton size="sm" variant="secondary" onClick={() => handleShowUsages(item)}>
                      查看引用
                    </AdminButton>
                    <AdminButton
                      size="sm"
                      variant="danger"
                      onClick={() => handleDelete(item)}
                      disabled={deleting === item.id}
                    >
                      {deleting === item.id ? "删除中…" : "删除"}
                    </AdminButton>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <AdminPagination
            page={page}
            totalPages={totalPages}
            total={total}
            onPageChange={setPage}
          />
        </>
      )}

      {/* Recrop modal */}
      {cropItem && (
        <RecropModal
          item={cropItem}
          onClose={() => setCropItem(null)}
          onSuccess={() => {
            void loadMedia();
            setCropItem(null);
          }}
        />
      )}

      <AdminModal
        open={Boolean(usageItem)}
        title={usageItem ? `图片引用 · ${usageItem.filename}` : "图片引用"}
        onClose={() => setUsageItem(null)}
      >
        {loadingUsages ? (
          <p className="text-sm text-slate-500">加载中…</p>
        ) : usages.length === 0 ? (
          <p className="text-sm text-slate-500">该图片未被任何页面或文章引用</p>
        ) : (
          <ul className="space-y-3">
            {usages.map((u, i) => (
              <li
                key={i}
                className="flex items-center gap-2 rounded-xl border border-slate-100 bg-slate-50 px-3 py-2.5"
              >
                <span className="whitespace-nowrap rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700 ring-1 ring-inset ring-blue-600/15">
                  {usageTypeLabel[u.type] || u.type}
                </span>
                <span className="truncate text-sm font-medium text-slate-900">{u.title}</span>
                <span className="whitespace-nowrap text-xs text-slate-500">({u.field})</span>
              </li>
            ))}
          </ul>
        )}
      </AdminModal>
    </div>
  );
}
