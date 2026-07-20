import { useState, useEffect, useCallback } from "react";
import { listMedia } from "@/api/media";
import type { MediaItem } from "@/api/media";

interface GalleryPickerModalProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (items: MediaItem[]) => void;
}

export default function GalleryPickerModal({ open, onClose, onConfirm }: GalleryPickerModalProps) {
  const [items, setItems] = useState<MediaItem[]>([]);
  const [selected, setSelected] = useState<MediaItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const pageSize = 20;

  const loadMedia = useCallback(async () => {
    setLoading(true);
    try {
      const data = await listMedia(page, pageSize, "image");
      setItems(data.items || []);
      setTotal(data.total);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [page]);

  useEffect(() => {
    if (open) {
      setSelected([]);
      setPage(1);
      loadMedia();
    }
  }, [open, loadMedia]);

  useEffect(() => {
    if (open) loadMedia();
  }, [page, open, loadMedia]);

  const toggleSelect = (item: MediaItem) => {
    setSelected((prev) =>
      prev.some((s) => s.id === item.id)
        ? prev.filter((s) => s.id !== item.id)
        : [...prev, item]
    );
  };

  const totalPages = Math.ceil(total / pageSize);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/45 p-4 backdrop-blur-[2px]">
      <div className="flex max-h-[80vh] w-[90vw] max-w-4xl flex-col overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-[0_24px_64px_rgba(15,23,42,0.18)]">
        <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-slate-900">
            选择图片集 <span className="text-sm font-normal text-slate-500">({selected.length} 张已选)</span>
          </h3>
          <button onClick={onClose} className="text-slate-400 hover:text-slate-600 text-xl leading-none">&times;</button>
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          {loading ? (
            <div className="flex items-center justify-center h-48">
              <div className="text-slate-600">加载中...</div>
            </div>
          ) : (
            <div className="grid grid-cols-4 md:grid-cols-5 gap-3">
              {items.map((item) => {
                const isSelected = selected.some((s) => s.id === item.id);
                return (
                  <button
                    key={item.id}
                    onClick={() => toggleSelect(item)}
                    className={`relative aspect-square bg-slate-100 rounded-lg overflow-hidden border-2 transition-colors ${
                      isSelected ? "border-blue-500 ring-2 ring-blue-200" : "border-transparent hover:border-slate-200"
                    }`}
                  >
                    <img src={item.url} alt={item.filename} className="w-full h-full object-cover" loading="lazy" />
                    {isSelected && (
                      <div className="absolute top-1 right-1 w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center text-white text-xs font-bold">
                        {selected.findIndex((s) => s.id === item.id) + 1}
                      </div>
                    )}
                  </button>
                );
              })}
            </div>
          )}
        </div>

        <div className="px-6 py-3 border-t border-slate-100 flex items-center justify-between">
          <div className="flex items-center gap-2">
            {totalPages > 1 && (
              <>
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="inline-flex h-8 items-center rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-700 shadow-sm hover:bg-slate-50 disabled:opacity-50"
                >
                  上一页
                </button>
                <span className="text-sm text-slate-600">{page} / {totalPages}</span>
                <button
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                  className="inline-flex h-8 items-center rounded-lg border border-slate-200 bg-white px-3 text-sm text-slate-700 shadow-sm hover:bg-slate-50 disabled:opacity-50"
                >
                  下一页
                </button>
              </>
            )}
          </div>
          <div className="flex items-center gap-2">
            <button onClick={onClose} className="inline-flex h-9 items-center rounded-xl border border-slate-200 bg-white px-4 text-sm font-medium text-slate-700 shadow-sm hover:bg-slate-50">取消</button>
            <button
              onClick={() => { if (selected.length > 0) onConfirm(selected); }}
              disabled={selected.length === 0}
              className="inline-flex h-9 items-center rounded-xl bg-blue-600 px-4 text-sm font-medium text-white shadow-sm hover:bg-blue-700 disabled:opacity-50"
            >
              确认选择 ({selected.length})
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
