import MetadataEditor from "@/components/admin/MetadataEditor";
import ArticleTypographySettings from "@/components/admin/articles/ArticleTypographySettings";

// ── SEO fields panel ──

export interface SeoFieldsProps {
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
}

export function SeoFieldsPanel({
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
}: SeoFieldsProps) {
  return (
    <div className="px-4 py-3 border-t border-slate-100 bg-slate-50 space-y-3 max-h-80 overflow-y-auto">
      <div className="grid grid-cols-2 gap-3">
        <Field label="中文 SEO 标题" value={zhSeoTitle} onChange={setZhSeoTitle} placeholder="SEO 标题" />
        <Field label="英文 SEO 标题" value={enSeoTitle} onChange={setEnSeoTitle} placeholder="SEO Title" />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="block text-xs font-medium text-slate-600 mb-1">中文 Meta 描述</label>
          <textarea value={zhMetaDescription} onChange={(e) => setZhMetaDescription(e.target.value)} rows={2}
            className="w-full px-2 py-1.5 text-sm border border-slate-200 rounded-lg" placeholder="Meta 描述" />
        </div>
        <div>
          <label className="block text-xs font-medium text-slate-600 mb-1">英文 Meta 描述</label>
          <textarea value={enMetaDescription} onChange={(e) => setEnMetaDescription(e.target.value)} rows={2}
            className="w-full px-2 py-1.5 text-sm border border-slate-200 rounded-lg" placeholder="Meta Description" />
        </div>
      </div>
      <Field label="OG Image URL" value={ogImage} onChange={setOgImage} placeholder="https://..." />
    </div>
  );
}

// ── Advanced settings panel ──

export interface AdvancedSettingsProps {
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
}

export function AdvancedSettingsPanel({
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
}: AdvancedSettingsProps) {
  return (
    <div className="px-4 py-3 border-t border-slate-100 bg-slate-50 space-y-3 max-h-80 overflow-y-auto">
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="block text-xs font-medium text-slate-600 mb-1">可见性</label>
          <select value={visibility} onChange={(e) => setVisibility(e.target.value)}
            className="w-full px-2 py-1.5 text-sm border border-slate-200 rounded-lg">
            <option value="public">公开</option>
            <option value="private">私密</option>
            <option value="password_protected">密码保护</option>
          </select>
        </div>
        <div className="flex items-end gap-4 pb-1">
          <CheckboxField label="自动摘要" checked={autoSummary} onChange={setAutoSummary} />
          <CheckboxField label="允许评论" checked={allowComments} onChange={setAllowComments} />
          <CheckboxField label="置顶" checked={pinned} onChange={setPinned} />
        </div>
      </div>
      <ArticleTypographySettings metadata={metadata} onChange={setMetadata} />
      <div>
        <label className="block text-xs font-medium text-slate-600 mb-1">元数据</label>
        <MetadataEditor value={metadata} onChange={setMetadata} />
      </div>
    </div>
  );
}

// ── Shared helper components (also exported for use in page.tsx) ──

export function PopoverButton({ label, active, onClick }: { label: string; active: boolean; onClick: () => void }) {
  return (
    <button type="button" onClick={onClick}
      className={`px-2.5 py-1.5 text-xs rounded-lg border transition-colors ${
        active ? "bg-blue-50 border-blue-300 text-blue-700" : "border-slate-200 text-slate-600 hover:bg-slate-50"
      }`}>
      {label}
    </button>
  );
}

export function Field({ label, value, onChange, placeholder }: { label: string; value: string; onChange: (v: string) => void; placeholder?: string }) {
  return (
    <div>
      <label className="block text-xs font-medium text-slate-600 mb-1">{label}</label>
      <input type="text" value={value} onChange={(e) => onChange(e.target.value)} placeholder={placeholder}
        className="w-full px-2 py-1.5 text-sm border border-slate-200 rounded-lg focus:ring-1 focus:ring-blue-500 focus:border-blue-500" />
    </div>
  );
}

export function CheckboxField({ label, checked, onChange }: { label: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <label className="flex items-center gap-1.5 text-xs text-slate-600 cursor-pointer">
      <input type="checkbox" checked={checked} onChange={(e) => onChange(e.target.checked)} className="rounded border-slate-200" />
      {label}
    </label>
  );
}
