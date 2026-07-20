import type { Category, Tag } from "@/api/articles";
import ImagePickerModal from "@/components/admin/ImagePickerModal";
import { Field } from "./SeoFields";

export interface ArticleFormProps {
  slug: string;
  setSlug: (v: string) => void;
  author: string;
  setAuthor: (v: string) => void;
  coverImage: string;
  setCoverImage: (v: string) => void;
  showCoverPicker: boolean;
  setShowCoverPicker: (v: boolean) => void;
  categories: Category[];
  selectedCategoryIds: number[];
  toggleCategory: (id: number) => void;
  tags: Tag[];
  selectedTagIds: number[];
  toggleTag: (id: number) => void;
}

export default function ArticleForm({
  slug,
  setSlug,
  author,
  setAuthor,
  coverImage,
  setCoverImage,
  showCoverPicker,
  setShowCoverPicker,
  categories,
  selectedCategoryIds,
  toggleCategory,
  tags,
  selectedTagIds,
  toggleTag,
}: ArticleFormProps) {
  return (
    <>
      <div className="px-4 py-3 border-t border-slate-100 bg-slate-50 space-y-3 max-h-80 overflow-y-auto">
        <div className="grid grid-cols-2 gap-3">
          <Field label="Slug" value={slug} onChange={setSlug} placeholder="article-url-slug" />
          <Field label="作者" value={author} onChange={setAuthor} placeholder="作者名" />
        </div>
        <div>
          <label className="block text-xs font-medium text-slate-600 mb-1">封面图</label>
          <div className="flex items-center gap-2">
            <input type="text" value={coverImage} onChange={(e) => setCoverImage(e.target.value)}
              className="flex-1 px-2 py-1.5 text-sm border border-slate-200 rounded-lg" placeholder="URL 或点击选择" />
            <button type="button" onClick={() => setShowCoverPicker(true)}
              className="px-3 py-1.5 text-xs border border-slate-200 rounded-lg hover:bg-slate-100">选择</button>
          </div>
          {coverImage && <img src={coverImage} alt="封面" className="mt-1.5 max-h-20 rounded border border-slate-200"
            onError={(e) => { (e.target as HTMLImageElement).style.display = "none"; }} />}
        </div>
        <div>
          <label className="block text-xs font-medium text-slate-600 mb-1">分类</label>
          {categories.length === 0 ? <span className="text-xs text-slate-400">无分类</span> : (
            <div className="flex flex-wrap gap-1.5">
              {categories.map((cat) => (
                <button key={cat.id} type="button" onClick={() => toggleCategory(cat.id)}
                  className={`px-2.5 py-1 text-xs rounded-full border transition-colors ${
                    selectedCategoryIds.includes(cat.id) ? "bg-purple-100 border-purple-300 text-purple-800" : "bg-white border-slate-200 text-slate-600 hover:bg-slate-50"
                  }`}>{cat.zhName || cat.enName}</button>
              ))}
            </div>
          )}
        </div>
        <div>
          <label className="block text-xs font-medium text-slate-600 mb-1">标签</label>
          {tags.length === 0 ? <span className="text-xs text-slate-400">无标签</span> : (
            <div className="flex flex-wrap gap-1.5">
              {tags.map((tag) => (
                <button key={tag.id} type="button" onClick={() => toggleTag(tag.id)}
                  className={`px-2.5 py-1 text-xs rounded-full border transition-colors ${
                    selectedTagIds.includes(tag.id) ? "bg-blue-100 border-blue-300 text-blue-800" : "bg-white border-slate-200 text-slate-600 hover:bg-slate-50"
                  }`}>{tag.zhName || tag.enName}</button>
              ))}
            </div>
          )}
        </div>
      </div>

      <ImagePickerModal
        open={showCoverPicker}
        onClose={() => setShowCoverPicker(false)}
        onSelect={(item) => { setCoverImage(item.url); setShowCoverPicker(false); }}
      />
    </>
  );
}
