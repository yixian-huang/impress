import { AdminField, AdminInput, AdminTextarea } from "@/components/admin/ui";

interface SeoFieldGroupProps {
  seoTitle: string;
  onSeoTitleChange: (value: string) => void;
  metaDescription: string;
  onMetaDescriptionChange: (value: string) => void;
  ogImage?: string;
  onOgImageChange?: (value: string) => void;
  keywords?: string;
  onKeywordsChange?: (value: string) => void;
  label?: string;
}

export default function SeoFieldGroup({
  seoTitle,
  onSeoTitleChange,
  metaDescription,
  onMetaDescriptionChange,
  ogImage,
  onOgImageChange,
  keywords,
  onKeywordsChange,
  label = "SEO",
}: SeoFieldGroupProps) {
  return (
    <div className="space-y-3">
      <h4 className="text-sm font-medium text-slate-700">{label}</h4>
      <AdminField label="SEO Title" hint={`${seoTitle.length}/70`}>
        <AdminInput
          type="text"
          value={seoTitle}
          onChange={(e) => onSeoTitleChange(e.target.value)}
          className="rounded-lg py-1.5"
          placeholder="Override page title for search engines"
          maxLength={70}
        />
      </AdminField>
      <AdminField label="Meta Description" hint={`${metaDescription.length}/160`}>
        <AdminTextarea
          value={metaDescription}
          onChange={(e) => onMetaDescriptionChange(e.target.value)}
          className="rounded-lg py-1.5"
          rows={2}
          placeholder="Description for search engine results"
          maxLength={160}
        />
      </AdminField>
      {onKeywordsChange && (
        <AdminField label="Keywords">
          <AdminInput
            type="text"
            value={keywords ?? ""}
            onChange={(e) => onKeywordsChange(e.target.value)}
            className="rounded-lg py-1.5"
            placeholder="Comma-separated keywords"
          />
        </AdminField>
      )}
      {onOgImageChange && (
        <AdminField label="OG Image URL">
          <AdminInput
            type="text"
            value={ogImage ?? ""}
            onChange={(e) => onOgImageChange(e.target.value)}
            className="rounded-lg py-1.5"
            placeholder="Image for social sharing preview"
          />
        </AdminField>
      )}
    </div>
  );
}
