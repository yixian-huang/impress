import { useState, useEffect, useCallback } from "react";
import {
  getCategoryTree,
  createCategory,
  updateCategory,
  deleteCategory,
} from "@/api/articles";
import type { Category } from "@/api/articles";
import MetadataEditor from "@/components/admin/MetadataEditor";
import {
  AdminBadge,
  AdminButton,
  AdminCard,
  AdminCheckbox,
  AdminEmptyState,
  AdminErrorBanner,
  AdminField,
  AdminInput,
  AdminLoading,
  AdminPageHeader,
  AdminSelect,
  AdminTextarea,
  AdminTextButton,
  useAdminConfirm,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

// Flatten tree into a list with depth info for rendering
function flattenTree(cats: Category[], depth = 0): { cat: Category; depth: number }[] {
  const result: { cat: Category; depth: number }[] = [];
  for (const cat of cats) {
    result.push({ cat, depth });
    if (cat.children && cat.children.length > 0) {
      result.push(...flattenTree(cat.children, depth + 1));
    }
  }
  return result;
}

// Collect all categories as flat list for parent selector
function flattenAll(cats: Category[]): Category[] {
  const result: Category[] = [];
  for (const cat of cats) {
    result.push(cat);
    if (cat.children && cat.children.length > 0) {
      result.push(...flattenAll(cat.children));
    }
  }
  return result;
}

export default function CategoriesPage() {
  useDocumentTitle("分类管理");
  const { confirm, confirmDialog } = useAdminConfirm();

  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<number | null>(null);

  // Inline edit state
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editSlug, setEditSlug] = useState("");
  const [editZhName, setEditZhName] = useState("");
  const [editEnName, setEditEnName] = useState("");
  const [editParentId, setEditParentId] = useState<number | null>(null);
  const [editCoverImage, setEditCoverImage] = useState("");
  const [editZhDescription, setEditZhDescription] = useState("");
  const [editEnDescription, setEditEnDescription] = useState("");
  const [editHideFromList, setEditHideFromList] = useState(false);
  const [editPreventCascade, setEditPreventCascade] = useState(false);
  const [editSortOrder, setEditSortOrder] = useState(0);
  const [editMetadata, setEditMetadata] = useState<Record<string, unknown>>({});
  const [saving, setSaving] = useState(false);

  // New category form
  const [showNew, setShowNew] = useState(false);
  const [newSlug, setNewSlug] = useState("");
  const [newZhName, setNewZhName] = useState("");
  const [newEnName, setNewEnName] = useState("");
  const [newParentId, setNewParentId] = useState<number | null>(null);
  const [newCoverImage, setNewCoverImage] = useState("");
  const [newZhDescription, setNewZhDescription] = useState("");
  const [newEnDescription, setNewEnDescription] = useState("");
  const [newHideFromList, setNewHideFromList] = useState(false);
  const [newPreventCascade, setNewPreventCascade] = useState(false);
  const [newSortOrder, setNewSortOrder] = useState(0);
  const [newMetadata, setNewMetadata] = useState<Record<string, unknown>>({});

  const loadCategories = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getCategoryTree();
      setCategories(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load categories");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadCategories();
  }, [loadCategories]);

  const allFlat = flattenAll(categories);
  const flatList = flattenTree(categories);

  const handleCreate = async () => {
    if (!newSlug.trim() || !newZhName.trim()) {
      setError("Slug and Chinese name are required");
      return;
    }

    setSaving(true);
    setError(null);
    try {
      await createCategory({
        slug: newSlug,
        zhName: newZhName,
        enName: newEnName,
        parentId: newParentId,
        coverImage: newCoverImage,
        zhDescription: newZhDescription,
        enDescription: newEnDescription,
        hideFromList: newHideFromList,
        preventCascade: newPreventCascade,
        sortOrder: newSortOrder,
        metadata: newMetadata,
      });
      setShowNew(false);
      setNewSlug("");
      setNewZhName("");
      setNewEnName("");
      setNewParentId(null);
      setNewCoverImage("");
      setNewZhDescription("");
      setNewEnDescription("");
      setNewHideFromList(false);
      setNewPreventCascade(false);
      setNewSortOrder(0);
      setNewMetadata({});
      await loadCategories();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    } finally {
      setSaving(false);
    }
  };

  const startEdit = (cat: Category) => {
    setEditingId(cat.id);
    setEditSlug(cat.slug);
    setEditZhName(cat.zhName);
    setEditEnName(cat.enName);
    setEditParentId(cat.parentId ?? null);
    setEditCoverImage(cat.coverImage || "");
    setEditZhDescription(cat.zhDescription || "");
    setEditEnDescription(cat.enDescription || "");
    setEditHideFromList(cat.hideFromList || false);
    setEditPreventCascade(cat.preventCascade || false);
    setEditSortOrder(cat.sortOrder || 0);
    setEditMetadata(cat.metadata || {});
  };

  const handleUpdate = async () => {
    if (!editingId || !editSlug.trim() || !editZhName.trim()) {
      setError("Slug and Chinese name are required");
      return;
    }

    setSaving(true);
    setError(null);
    try {
      await updateCategory(editingId, {
        slug: editSlug,
        zhName: editZhName,
        enName: editEnName,
        parentId: editParentId,
        coverImage: editCoverImage,
        zhDescription: editZhDescription,
        enDescription: editEnDescription,
        hideFromList: editHideFromList,
        preventCascade: editPreventCascade,
        sortOrder: editSortOrder,
        metadata: editMetadata,
      });
      setEditingId(null);
      await loadCategories();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Update failed");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (cat: Category) => {
    const name = cat.zhName || cat.enName;
    const ok = await confirm({
      title: "删除分类",
      message: `确定删除分类「${name}」吗？此操作不可撤销。`,
      confirmLabel: "删除",
      danger: true,
    });
    if (!ok) return;

    setDeleting(cat.id);
    setError(null);
    try {
      await deleteCategory(cat.id);
      await loadCategories();
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除失败");
    } finally {
      setDeleting(null);
    }
  };

  // Category form fields shared between new and edit
  const renderFormFields = (
    mode: "new" | "edit",
    vals: {
      slug: string; zhName: string; enName: string; parentId: number | null;
      coverImage: string; zhDescription: string; enDescription: string;
      hideFromList: boolean; preventCascade: boolean; sortOrder: number;
      metadata: Record<string, unknown>;
    },
    setters: {
      setSlug: (v: string) => void; setZhName: (v: string) => void; setEnName: (v: string) => void;
      setParentId: (v: number | null) => void; setCoverImage: (v: string) => void;
      setZhDescription: (v: string) => void; setEnDescription: (v: string) => void;
      setHideFromList: (v: boolean) => void; setPreventCascade: (v: boolean) => void;
      setSortOrder: (v: number) => void; setMetadata: (v: Record<string, unknown>) => void;
    },
    excludeId?: number,
  ) => (
    <div className="space-y-4">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <AdminField label="Slug">
          <AdminInput
            type="text"
            value={vals.slug}
            onChange={(e) => setters.setSlug(e.target.value)}
            placeholder="category-slug"
          />
        </AdminField>
        <AdminField label="中文名称">
          <AdminInput
            type="text"
            value={vals.zhName}
            onChange={(e) => setters.setZhName(e.target.value)}
            placeholder="中文名称"
          />
        </AdminField>
        <AdminField label="English Name">
          <AdminInput
            type="text"
            value={vals.enName}
            onChange={(e) => setters.setEnName(e.target.value)}
            placeholder="English name"
          />
        </AdminField>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <AdminField label="父级分类">
          <AdminSelect
            value={vals.parentId ?? ""}
            onChange={(e) => setters.setParentId(e.target.value ? Number(e.target.value) : null)}
            className="w-full"
          >
            <option value="">无 (顶级分类)</option>
            {allFlat
              .filter((c) => c.id !== excludeId)
              .map((c) => (
                <option key={c.id} value={c.id}>
                  {c.zhName || c.enName}
                </option>
              ))}
          </AdminSelect>
        </AdminField>
        <AdminField label="封面图片 URL">
          <AdminInput
            type="text"
            value={vals.coverImage}
            onChange={(e) => setters.setCoverImage(e.target.value)}
            placeholder="https://..."
          />
        </AdminField>
        <AdminField label="排序">
          <AdminInput
            type="number"
            value={vals.sortOrder}
            onChange={(e) => setters.setSortOrder(Number(e.target.value))}
          />
        </AdminField>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <AdminField label="中文描述">
          <AdminTextarea
            value={vals.zhDescription}
            onChange={(e) => setters.setZhDescription(e.target.value)}
            rows={2}
            placeholder="中文描述"
          />
        </AdminField>
        <AdminField label="English Description">
          <AdminTextarea
            value={vals.enDescription}
            onChange={(e) => setters.setEnDescription(e.target.value)}
            rows={2}
            placeholder="English description"
          />
        </AdminField>
      </div>

      <div className="flex flex-wrap items-center gap-6">
        <AdminCheckbox
          checked={vals.hideFromList}
          onChange={(e) => setters.setHideFromList(e.target.checked)}
          label="从列表隐藏"
        />
        <AdminCheckbox
          checked={vals.preventCascade}
          onChange={(e) => setters.setPreventCascade(e.target.checked)}
          label="阻止级联"
        />
      </div>

      <AdminField label="元数据">
        <MetadataEditor value={vals.metadata} onChange={setters.setMetadata} />
      </AdminField>
    </div>
  );

  return (
    <div>
      {confirmDialog}
      <AdminPageHeader
        title="分类管理"
        description="管理文章分类与层级"
        breadcrumbs={[
          { label: "文章管理", to: "/admin/articles" },
          { label: "分类" },
        ]}
        actions={
          <AdminButton size="sm" onClick={() => setShowNew(!showNew)}>
            {showNew ? "取消" : "新建分类"}
          </AdminButton>
        }
      />

      {error && <AdminErrorBanner message={error} onDismiss={() => setError(null)} />}

      {showNew && (
        <AdminCard className="mb-6" title="新建分类">
          {renderFormFields(
            "new",
            {
              slug: newSlug, zhName: newZhName, enName: newEnName, parentId: newParentId,
              coverImage: newCoverImage, zhDescription: newZhDescription, enDescription: newEnDescription,
              hideFromList: newHideFromList, preventCascade: newPreventCascade, sortOrder: newSortOrder,
              metadata: newMetadata,
            },
            {
              setSlug: setNewSlug, setZhName: setNewZhName, setEnName: setNewEnName,
              setParentId: setNewParentId, setCoverImage: setNewCoverImage,
              setZhDescription: setNewZhDescription, setEnDescription: setNewEnDescription,
              setHideFromList: setNewHideFromList, setPreventCascade: setNewPreventCascade,
              setSortOrder: setNewSortOrder, setMetadata: setNewMetadata,
            },
          )}
          <AdminButton className="mt-4" size="sm" onClick={handleCreate} disabled={saving}>
            {saving ? "创建中…" : "创建"}
          </AdminButton>
        </AdminCard>
      )}

      {loading ? (
        <AdminLoading />
      ) : flatList.length === 0 ? (
        <AdminEmptyState title="暂无分类" description="点击「新建分类」创建第一个分类。" />
      ) : (
        <div className="space-y-2">
          {flatList.map(({ cat, depth }) => (
            <AdminCard key={cat.id} padded={false} className="overflow-hidden">
              {editingId === cat.id ? (
                <div className="p-5 sm:p-6">
                  <h3 className="mb-4 text-sm font-semibold text-slate-900">编辑分类</h3>
                  {renderFormFields(
                    "edit",
                    {
                      slug: editSlug, zhName: editZhName, enName: editEnName, parentId: editParentId,
                      coverImage: editCoverImage, zhDescription: editZhDescription, enDescription: editEnDescription,
                      hideFromList: editHideFromList, preventCascade: editPreventCascade, sortOrder: editSortOrder,
                      metadata: editMetadata,
                    },
                    {
                      setSlug: setEditSlug, setZhName: setEditZhName, setEnName: setEditEnName,
                      setParentId: setEditParentId, setCoverImage: setEditCoverImage,
                      setZhDescription: setEditZhDescription, setEnDescription: setEditEnDescription,
                      setHideFromList: setEditHideFromList, setPreventCascade: setEditPreventCascade,
                      setSortOrder: setEditSortOrder, setMetadata: setEditMetadata,
                    },
                    cat.id,
                  )}
                  <div className="mt-4 flex items-center gap-2">
                    <AdminButton size="sm" onClick={handleUpdate} disabled={saving}>
                      {saving ? "保存中…" : "保存"}
                    </AdminButton>
                    <AdminButton size="sm" variant="secondary" onClick={() => setEditingId(null)}>
                      取消
                    </AdminButton>
                  </div>
                </div>
              ) : (
                <div
                  className="flex items-center justify-between px-5 py-4 transition hover:bg-slate-50/80 sm:px-6"
                  style={{ paddingLeft: `${1.25 + depth * 1.25}rem` }}
                >
                  <div className="flex min-w-0 items-center gap-3">
                    {depth > 0 && <span className="text-sm text-slate-300">{"└─"}</span>}
                    {cat.coverImage && (
                      <img
                        src={cat.coverImage}
                        alt=""
                        className="h-8 w-8 shrink-0 rounded-lg object-cover"
                        onError={(e) => { (e.target as HTMLImageElement).style.display = "none"; }}
                      />
                    )}
                    <div className="min-w-0">
                      <div className="truncate text-sm font-medium text-slate-900">
                        {cat.zhName || cat.enName}
                        {cat.enName && cat.zhName && (
                          <span className="ml-2 font-normal text-slate-400">{cat.enName}</span>
                        )}
                      </div>
                      <div className="text-xs text-slate-500">{cat.slug}</div>
                    </div>
                    {cat.hideFromList && <AdminBadge>隐藏</AdminBadge>}
                    {cat.sortOrder > 0 && (
                      <span className="text-xs text-slate-400">#{cat.sortOrder}</span>
                    )}
                  </div>
                  <div className="flex shrink-0 items-center gap-3">
                    <AdminTextButton onClick={() => startEdit(cat)}>编辑</AdminTextButton>
                    <AdminTextButton
                      tone="danger"
                      onClick={() => handleDelete(cat)}
                      disabled={deleting === cat.id}
                    >
                      {deleting === cat.id ? "…" : "删除"}
                    </AdminTextButton>
                  </div>
                </div>
              )}
            </AdminCard>
          ))}
        </div>
      )}
    </div>
  );
}
