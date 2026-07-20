import { useState, useEffect, useCallback } from "react";
import { listMedia } from "@/api/media";
import type { MediaItem } from "@/api/media";
import ImageCropUpload from "@/components/admin/ImageCropUpload";
import { formatUploadError, uploadMediaTracked } from "@/lib/mediaUploadTracked";

interface ImagePickerModalProps {
  open: boolean;
  onClose: () => void;
  onSelect: (item: MediaItem) => void;
}

export default function ImagePickerModal({ open, onClose, onSelect }: ImagePickerModalProps) {
  const [items, setItems] = useState<MediaItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [showUpload, setShowUpload] = useState(false);
  const pageSize = 12;

  const loadMedia = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await listMedia(page, pageSize);
      setItems(data.items || []);
      setTotal(data.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载图片失败");
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    if (open) {
      loadMedia();
    }
  }, [open, loadMedia]);

  const handleUpload = (item: MediaItem) => {
    setShowUpload(false);
    onSelect(item);
  };

  const handleDirectUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setError(null);
    setUploading(true);
    try {
      // Progress/retry surface via MediaUploadTray when mounted (e.g. article editor)
      const item = await uploadMediaTracked(file);
      onSelect(item);
    } catch (err) {
      setError(formatUploadError(err, "上传失败"));
    } finally {
      setUploading(false);
    }
    e.target.value = "";
  };

  const totalPages = Math.ceil(total / pageSize);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/45 p-4 backdrop-blur-[2px]">
      <div className="flex max-h-[80vh] w-[90vw] max-w-4xl flex-col overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-[0_24px_64px_rgba(15,23,42,0.18)]">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-slate-900">选择图片</h3>
          <div className="flex items-center gap-3">
            <label
              className={`inline-flex h-8 cursor-pointer items-center rounded-lg border border-slate-200 bg-white px-3 text-xs font-medium text-slate-700 shadow-sm hover:bg-slate-50 ${
                uploading ? "opacity-50 pointer-events-none" : ""
              }`}
            >
              {uploading ? "上传中…" : "直接上传"}
              <input
                type="file"
                accept="image/*"
                onChange={handleDirectUpload}
                className="hidden"
                disabled={uploading}
              />
            </label>
            <button
              onClick={() => setShowUpload(!showUpload)}
              className="px-3 py-1.5 text-xs bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              裁剪上传
            </button>
            <button
              onClick={onClose}
              className="text-slate-400 hover:text-slate-600 text-xl leading-none"
            >
              ×
            </button>
          </div>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto p-6">
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-800 text-sm">
              {error}
            </div>
          )}

          {showUpload && (
            <div className="mb-6 p-4 border rounded-lg bg-slate-50">
              <ImageCropUpload onUpload={handleUpload} />
            </div>
          )}

          {loading ? (
            <div className="flex items-center justify-center h-48">
              <div className="text-slate-600">加载中...</div>
            </div>
          ) : items.length === 0 ? (
            <div className="flex items-center justify-center h-48">
              <p className="text-slate-500">暂无图片</p>
            </div>
          ) : (
            <div className="grid grid-cols-3 md:grid-cols-4 gap-3">
              {items.map((item) => (
                <button
                  key={item.id}
                  onClick={() => onSelect(item)}
                  className="group relative aspect-square bg-slate-100 rounded-lg overflow-hidden border-2 border-transparent hover:border-blue-500 transition-colors"
                >
                  <img
                    src={item.url}
                    alt={item.filename}
                    className="w-full h-full object-cover"
                    loading="lazy"
                  />
                  <div className="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/60 to-transparent p-2">
                    <p className="text-xs text-white truncate">{item.filename}</p>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Footer with pagination */}
        {totalPages > 1 && (
          <div className="px-6 py-3 border-t border-slate-100 flex items-center justify-center gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="inline-flex h-8 items-center rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-700 shadow-sm hover:bg-slate-50 disabled:opacity-50"
            >
              上一页
            </button>
            <span className="text-sm text-slate-600">
              {page} / {totalPages}
            </span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="inline-flex h-8 items-center rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-700 shadow-sm hover:bg-slate-50 disabled:opacity-50"
            >
              下一页
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
