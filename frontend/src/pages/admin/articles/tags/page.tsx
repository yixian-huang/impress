import { useState, useEffect, useCallback } from "react";
import { getTags, createTag, updateTag, deleteTag } from "@/api/articles";
import type { Tag } from "@/api/articles";
import MetadataEditor from "@/components/admin/MetadataEditor";
import {
  AdminButton,
  AdminCard,
  AdminEmptyState,
  AdminErrorBanner,
  AdminField,
  AdminInput,
  AdminLoading,
  AdminPageHeader,
  AdminTextButton,
  useAdminConfirm,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

export default function TagsPage() {
  useDocumentTitle("标签管理");
  const { confirm, confirmDialog } = useAdminConfirm();

  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<number | null>(null);
  const [saving, setSaving] = useState(false);

  // New tag form
  const [showNew, setShowNew] = useState(false);
  const [newSlug, setNewSlug] = useState("");
  const [newZhName, setNewZhName] = useState("");
  const [newEnName, setNewEnName] = useState("");
  const [newColor, setNewColor] = useState("#6B7280");
  const [newCoverImage, setNewCoverImage] = useState("");
  const [newMetadata, setNewMetadata] = useState<Record<string, unknown>>({});

  // Edit tag state
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editSlug, setEditSlug] = useState("");
  const [editZhName, setEditZhName] = useState("");
  const [editEnName, setEditEnName] = useState("");
  const [editColor, setEditColor] = useState("#6B7280");
  const [editCoverImage, setEditCoverImage] = useState("");
  const [editMetadata, setEditMetadata] = useState<Record<string, unknown>>({});

  const loadTags = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getTags();
      setTags(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load tags");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadTags();
  }, [loadTags]);

  const handleCreate = async () => {
    if (!newSlug.trim() || !newZhName.trim()) {
      setError("Slug and Chinese name are required");
      return;
    }

    setSaving(true);
    setError(null);
    try {
      const created = await createTag({
        slug: newSlug,
        zhName: newZhName,
        enName: newEnName,
        color: newColor,
        coverImage: newCoverImage,
        metadata: newMetadata,
      });
      setTags((prev) => [...prev, created]);
      setShowNew(false);
      setNewSlug("");
      setNewZhName("");
      setNewEnName("");
      setNewColor("#6B7280");
      setNewCoverImage("");
      setNewMetadata({});
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (tag: Tag) => {
    setEditingId(tag.id);
    setEditSlug(tag.slug);
    setEditZhName(tag.zhName);
    setEditEnName(tag.enName);
    setEditColor(tag.color || "#6B7280");
    setEditCoverImage(tag.coverImage || "");
    setEditMetadata(tag.metadata || {});
  };

  const handleUpdate = async () => {
    if (!editingId || !editSlug.trim() || !editZhName.trim()) {
      setError("Slug and Chinese name are required");
      return;
    }

    setSaving(true);
    setError(null);
    try {
      const updated = await updateTag(editingId, {
        slug: editSlug,
        zhName: editZhName,
        enName: editEnName,
        color: editColor,
        coverImage: editCoverImage,
        metadata: editMetadata,
      });
      setTags((prev) => prev.map((t) => (t.id === editingId ? updated : t)));
      setEditingId(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Update failed");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (tag: Tag) => {
    const name = tag.zhName || tag.enName;
    const ok = await confirm({
      title: "删除标签",
      message: `确定删除标签「${name}」吗？此操作不可撤销。`,
      confirmLabel: "删除",
      danger: true,
    });
    if (!ok) return;

    setDeleting(tag.id);
    setError(null);
    try {
      await deleteTag(tag.id);
      setTags((prev) => prev.filter((t) => t.id !== tag.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除失败");
    } finally {
      setDeleting(null);
    }
  };

  return (
    <div>
      {confirmDialog}
      <AdminPageHeader
        title="标签管理"
        description="管理文章标签"
        breadcrumbs={[
          { label: "文章管理", to: "/admin/articles" },
          { label: "标签" },
        ]}
        actions={
          <AdminButton size="sm" onClick={() => setShowNew(!showNew)}>
            {showNew ? "取消" : "新建标签"}
          </AdminButton>
        }
      />

      {error && <AdminErrorBanner message={error} onDismiss={() => setError(null)} />}

      {showNew && (
        <AdminCard className="mb-6" title="新建标签">
          <div className="space-y-4">
            <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
              <AdminField label="Slug">
                <AdminInput value={newSlug} onChange={(e) => setNewSlug(e.target.value)} placeholder="tag-slug" />
              </AdminField>
              <AdminField label="中文名称">
                <AdminInput value={newZhName} onChange={(e) => setNewZhName(e.target.value)} placeholder="中文名称" />
              </AdminField>
              <AdminField label="English Name">
                <AdminInput value={newEnName} onChange={(e) => setNewEnName(e.target.value)} placeholder="English name" />
              </AdminField>
            </div>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <AdminField label="颜色">
                <div className="flex items-center gap-2">
                  <input
                    type="color"
                    value={newColor}
                    onChange={(e) => setNewColor(e.target.value)}
                    className="h-10 w-10 cursor-pointer rounded-lg border border-slate-200"
                  />
                  <AdminInput value={newColor} onChange={(e) => setNewColor(e.target.value)} placeholder="#6B7280" />
                </div>
              </AdminField>
              <AdminField label="封面图片 URL">
                <AdminInput value={newCoverImage} onChange={(e) => setNewCoverImage(e.target.value)} placeholder="https://..." />
              </AdminField>
            </div>
            <AdminField label="元数据">
              <MetadataEditor value={newMetadata} onChange={setNewMetadata} />
            </AdminField>
            <AdminButton size="sm" onClick={handleCreate} disabled={saving}>
              {saving ? "创建中…" : "创建"}
            </AdminButton>
          </div>
        </AdminCard>
      )}

      {loading ? (
        <AdminLoading />
      ) : tags.length === 0 ? (
        <AdminEmptyState title="暂无标签" description="点击「新建标签」创建第一个标签。" />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {tags.map((tag) => (
            <AdminCard key={tag.id} padded={false} className="overflow-hidden">
              {editingId === tag.id ? (
                <div className="space-y-3 p-4">
                  <AdminField label="Slug">
                    <AdminInput value={editSlug} onChange={(e) => setEditSlug(e.target.value)} />
                  </AdminField>
                  <div className="grid grid-cols-2 gap-2">
                    <AdminField label="中文名称">
                      <AdminInput value={editZhName} onChange={(e) => setEditZhName(e.target.value)} />
                    </AdminField>
                    <AdminField label="English">
                      <AdminInput value={editEnName} onChange={(e) => setEditEnName(e.target.value)} />
                    </AdminField>
                  </div>
                  <AdminField label="颜色">
                    <div className="flex items-center gap-2">
                      <input
                        type="color"
                        value={editColor}
                        onChange={(e) => setEditColor(e.target.value)}
                        className="h-9 w-9 cursor-pointer rounded-lg border border-slate-200"
                      />
                      <AdminInput value={editColor} onChange={(e) => setEditColor(e.target.value)} />
                    </div>
                  </AdminField>
                  <AdminField label="封面图片 URL">
                    <AdminInput value={editCoverImage} onChange={(e) => setEditCoverImage(e.target.value)} placeholder="https://..." />
                  </AdminField>
                  <AdminField label="元数据">
                    <MetadataEditor value={editMetadata} onChange={setEditMetadata} />
                  </AdminField>
                  <div className="flex items-center gap-2 pt-1">
                    <AdminButton size="sm" onClick={handleUpdate} disabled={saving}>
                      {saving ? "保存中…" : "保存"}
                    </AdminButton>
                    <AdminButton size="sm" variant="secondary" onClick={() => setEditingId(null)}>
                      取消
                    </AdminButton>
                  </div>
                </div>
              ) : (
                <>
                  {tag.coverImage && (
                    <img
                      src={tag.coverImage}
                      alt=""
                      className="h-32 w-full object-cover"
                      onError={(e) => { (e.target as HTMLImageElement).style.display = "none"; }}
                    />
                  )}
                  <div className="p-4">
                    <div className="mb-2 flex items-center gap-2">
                      <span
                        className="h-4 w-4 shrink-0 rounded-full border border-slate-200"
                        style={{ backgroundColor: tag.color || "#6B7280" }}
                      />
                      <span className="truncate text-sm font-medium text-slate-900">
                        {tag.zhName || tag.enName}
                      </span>
                    </div>
                    {tag.enName && tag.zhName && (
                      <div className="mb-1 text-xs text-slate-500">{tag.enName}</div>
                    )}
                    <div className="mb-3 text-xs text-slate-400">{tag.slug}</div>
                    <div className="flex items-center gap-3">
                      <AdminTextButton onClick={() => startEdit(tag)}>编辑</AdminTextButton>
                      <AdminTextButton
                        tone="danger"
                        onClick={() => handleDelete(tag)}
                        disabled={deleting === tag.id}
                      >
                        {deleting === tag.id ? "…" : "删除"}
                      </AdminTextButton>
                    </div>
                  </div>
                </>
              )}
            </AdminCard>
          ))}
        </div>
      )}
    </div>
  );
}
