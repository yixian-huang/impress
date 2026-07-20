import type { ArticleTypographyOverride } from "@/theme/typography";

interface ArticleTypographySettingsProps {
  metadata: Record<string, unknown>;
  onChange: (metadata: Record<string, unknown>) => void;
}

function getTypography(metadata: Record<string, unknown>): ArticleTypographyOverride {
  const t = metadata.typography;
  if (t && typeof t === "object") return t as ArticleTypographyOverride;
  return {};
}

function setTypography(metadata: Record<string, unknown>, patch: Partial<ArticleTypographyOverride>) {
  const current = getTypography(metadata);
  return {
    ...metadata,
    typography: { ...current, ...patch },
  };
}

export default function ArticleTypographySettings({ metadata, onChange }: ArticleTypographySettingsProps) {
  const typography = getTypography(metadata);

  return (
    <div className="border border-slate-200 rounded-lg p-3 space-y-3 bg-white">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium text-slate-700">本篇正文字体（覆盖主题默认）</span>
        <label className="inline-flex items-center gap-1.5 text-xs text-slate-600 cursor-pointer">
          <input
            type="checkbox"
            checked={typography.enabled === true}
            onChange={(e) => onChange(setTypography(metadata, { enabled: e.target.checked }))}
            className="rounded border-slate-200"
          />
          启用
        </label>
      </div>

      {typography.enabled && (
        <div>
          <label className="block text-xs text-slate-600 mb-1">正文字体</label>
          <select
            value={typography.bodyFontRole ?? "serif"}
            onChange={(e) =>
              onChange(setTypography(metadata, { bodyFontRole: e.target.value as "serif" | "sans" }))
            }
            className="w-full px-2 py-1.5 text-sm border border-slate-200 rounded-lg"
          >
            <option value="serif">衬线</option>
            <option value="sans">无衬线</option>
          </select>
          <p className="text-xs text-slate-400 mt-1">字体栈与字号在「主题 → 样式定制」中配置。</p>
        </div>
      )}
    </div>
  );
}
