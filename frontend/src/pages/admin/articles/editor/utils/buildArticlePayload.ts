import { slugifyTitle } from "./slugify";

/** Field slice needed to assemble a create/update article payload. */
export type ArticlePayloadFields = {
  zhTitle: string;
  enTitle: string;
  slug: string;
  coverImage: string;
  zhSeoTitle: string;
  enSeoTitle: string;
  zhMetaDescription: string;
  enMetaDescription: string;
  ogImage: string;
  selectedCategoryIds: number[];
  selectedTagIds: number[];
  author: string;
  autoSummary: boolean;
  allowComments: boolean;
  pinned: boolean;
  visibility: string;
  metadata: Record<string, unknown>;
};

export type ArticleBodies = {
  zhBody: string;
  enBody: string;
};

/**
 * Pure payload builder — no React state. Call only at save/schedule time.
 */
export function buildArticlePayload(
  fields: ArticlePayloadFields,
  bodies: ArticleBodies,
  status: "draft" | "published",
  publishedAt?: string,
): Record<string, unknown> {
  const finalSlug = fields.slug.trim() || slugifyTitle(fields.zhTitle);
  const payload: Record<string, unknown> = {
    zhTitle: fields.zhTitle,
    enTitle: fields.enTitle,
    slug: finalSlug,
    coverImage: fields.coverImage,
    zhBody: bodies.zhBody,
    enBody: bodies.enBody,
    zhSeoTitle: fields.zhSeoTitle,
    enSeoTitle: fields.enSeoTitle,
    zhMetaDescription: fields.zhMetaDescription,
    enMetaDescription: fields.enMetaDescription,
    ogImage: fields.ogImage,
    status,
    categoryIds: fields.selectedCategoryIds,
    tagIds: fields.selectedTagIds,
    author: fields.author,
    autoSummary: fields.autoSummary,
    allowComments: fields.allowComments,
    pinned: fields.pinned,
    visibility: fields.visibility,
    metadata: fields.metadata,
  };
  if (status === "published") {
    payload.publishedAt = publishedAt ?? new Date().toISOString();
  }
  return payload;
}

/** Pick payload fields from the form state object (stable field names). */
export function pickPayloadFields(form: ArticlePayloadFields): ArticlePayloadFields {
  return {
    zhTitle: form.zhTitle,
    enTitle: form.enTitle,
    slug: form.slug,
    coverImage: form.coverImage,
    zhSeoTitle: form.zhSeoTitle,
    enSeoTitle: form.enSeoTitle,
    zhMetaDescription: form.zhMetaDescription,
    enMetaDescription: form.enMetaDescription,
    ogImage: form.ogImage,
    selectedCategoryIds: form.selectedCategoryIds,
    selectedTagIds: form.selectedTagIds,
    author: form.author,
    autoSummary: form.autoSummary,
    allowComments: form.allowComments,
    pinned: form.pinned,
    visibility: form.visibility,
    metadata: form.metadata,
  };
}
