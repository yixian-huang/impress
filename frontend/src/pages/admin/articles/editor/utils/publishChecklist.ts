/** Pre-publish readiness checks for the article editor. */

export type ChecklistSeverity = "block" | "warn";

export type ChecklistItem = {
  id: string;
  severity: ChecklistSeverity;
  message: string;
  hint?: string;
};

export type PublishChecklistInput = {
  zhTitle: string;
  enTitle: string;
  slug: string;
  coverImage: string;
  zhBody: string;
  enBody: string;
  zhMetaDescription: string;
  enMetaDescription: string;
  zhSeoTitle: string;
  enSeoTitle: string;
  enabledLangs: string[];
  author?: string;
};

/** Strip HTML to plain text length (rough, for empty-body checks). */
export function plainTextFromHtml(html: string | undefined | null): string {
  if (!html) return "";
  return html
    .replace(/<script[\s\S]*?<\/script>/gi, "")
    .replace(/<style[\s\S]*?<\/style>/gi, "")
    .replace(/<[^>]+>/g, " ")
    .replace(/&nbsp;/gi, " ")
    .replace(/\s+/g, " ")
    .trim();
}

export function isBodyEmpty(html: string | undefined | null): boolean {
  return plainTextFromHtml(html).length === 0;
}

/** SEO title soft max (characters). */
export const SEO_TITLE_MAX = 60;
/** Meta description ideal max. */
export const SEO_DESC_MAX = 160;
/** Meta description soft minimum for warnings. */
export const SEO_DESC_MIN = 50;

/**
 * Evaluate publish readiness. `block` items should stop publish unless forced
 * is not offered for blocks — only warns can be skipped.
 */
export function evaluatePublishChecklist(input: PublishChecklistInput): ChecklistItem[] {
  const items: ChecklistItem[] = [];
  const zhTitle = (input.zhTitle || "").trim();
  const enTitle = (input.enTitle || "").trim();
  const slug = (input.slug || "").trim();
  const cover = (input.coverImage || "").trim();
  const zhMeta = (input.zhMetaDescription || "").trim();
  const enMeta = (input.enMetaDescription || "").trim();
  const zhSeo = (input.zhSeoTitle || "").trim();
  const enSeo = (input.enSeoTitle || "").trim();
  const enEnabled = input.enabledLangs.includes("en");

  if (!zhTitle) {
    items.push({
      id: "zh-title",
      severity: "block",
      message: "缺少中文标题",
      hint: "标题是发布与列表展示的必填项。",
    });
  }

  if (isBodyEmpty(input.zhBody)) {
    items.push({
      id: "zh-body",
      severity: "block",
      message: "中文正文为空",
      hint: "请至少写一段中文正文再发布。",
    });
  }

  if (!cover) {
    items.push({
      id: "cover",
      severity: "warn",
      message: "未设置封面图",
      hint: "列表与社交分享会更完整。",
    });
  }

  if (!slug) {
    items.push({
      id: "slug",
      severity: "warn",
      message: "未自定义 URL slug",
      hint: "将根据中文标题自动生成；发布后修改可能影响外链。",
    });
  }

  if (!zhMeta) {
    items.push({
      id: "zh-meta",
      severity: "warn",
      message: "缺少中文 Meta 描述",
      hint: "搜索结果摘要会更易点击。",
    });
  } else if (zhMeta.length < SEO_DESC_MIN) {
    items.push({
      id: "zh-meta-short",
      severity: "warn",
      message: `中文 Meta 描述偏短（${zhMeta.length} 字）`,
      hint: `建议 ${SEO_DESC_MIN}–${SEO_DESC_MAX} 字。`,
    });
  } else if (zhMeta.length > SEO_DESC_MAX) {
    items.push({
      id: "zh-meta-long",
      severity: "warn",
      message: `中文 Meta 描述偏长（${zhMeta.length} 字）`,
      hint: `建议不超过 ${SEO_DESC_MAX} 字，避免搜索结果被截断。`,
    });
  }

  if (zhSeo.length > SEO_TITLE_MAX) {
    items.push({
      id: "zh-seo-long",
      severity: "warn",
      message: `中文 SEO 标题偏长（${zhSeo.length} 字）`,
      hint: `建议不超过 ${SEO_TITLE_MAX} 字。`,
    });
  }

  if (enEnabled) {
    if (!enTitle && !isBodyEmpty(input.enBody)) {
      items.push({
        id: "en-title-missing",
        severity: "warn",
        message: "已启用英文但缺少英文标题",
      });
    }
    if (enTitle && isBodyEmpty(input.enBody)) {
      items.push({
        id: "en-body-missing",
        severity: "warn",
        message: "已有英文标题但英文正文为空",
      });
    }
    if (!enTitle && isBodyEmpty(input.enBody)) {
      items.push({
        id: "en-empty",
        severity: "warn",
        message: "已启用英文语言但未填写英文内容",
        hint: "可删除英文语言，或补全标题与正文。",
      });
    }
    if (enMeta && enMeta.length > SEO_DESC_MAX) {
      items.push({
        id: "en-meta-long",
        severity: "warn",
        message: `英文 Meta 描述偏长（${enMeta.length} 字）`,
      });
    }
    if (enSeo.length > SEO_TITLE_MAX) {
      items.push({
        id: "en-seo-long",
        severity: "warn",
        message: `英文 SEO 标题偏长（${enSeo.length} 字）`,
      });
    }
  }

  if (!(input.author || "").trim()) {
    items.push({
      id: "author",
      severity: "warn",
      message: "未填写作者",
    });
  }

  return items;
}

export function checklistHasBlocks(items: ChecklistItem[]): boolean {
  return items.some((i) => i.severity === "block");
}

export function checklistHasWarnings(items: ChecklistItem[]): boolean {
  return items.some((i) => i.severity === "warn");
}
