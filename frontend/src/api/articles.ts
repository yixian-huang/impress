import { http } from "@/api/http";

// ---------- Types ----------

export interface Article {
  id: number;
  slug: string;
  status: "draft" | "published" | "scheduled";
  zhTitle: string;
  enTitle: string;
  zhBody: string;
  enBody: string;
  coverImage: string;
  zhSeoTitle: string;
  enSeoTitle: string;
  zhMetaDescription: string;
  enMetaDescription: string;
  ogImage: string;
  categoryId: number | null;
  categoryIds?: number[];
  category?: Category;
  categories?: Category[];
  tags?: Tag[];
  author: string;
  autoSummary: boolean;
  allowComments: boolean;
  pinned: boolean;
  visibility: string;
  metadata: Record<string, unknown>;
  publishedAt: string | null;
  scheduledAt?: string | null;
  createdAt: string;
  updatedAt: string;
  /** Total page views (public detail only). */
  viewCount?: number;
}

export interface Category {
  id: number;
  slug: string;
  zhName: string;
  enName: string;
  parentId?: number | null;
  parent?: Category;
  children?: Category[];
  coverImage: string;
  zhDescription: string;
  enDescription: string;
  hideFromList: boolean;
  preventCascade: boolean;
  metadata: Record<string, unknown>;
  sortOrder: number;
}

export interface Tag {
  id: number;
  slug: string;
  zhName: string;
  enName: string;
  color: string;
  coverImage: string;
  metadata: Record<string, unknown>;
}

interface ArticleListResponse {
  items: Article[];
  total: number;
  page: number;
  pageSize: number;
}

interface PublicArticleListResponse {
  items: Article[];
  total: number;
  page: number;
  pageSize: number;
}

// ---------- Public APIs ----------

export async function getPublicArticles(
  page: number = 1,
  pageSize: number = 10,
  category?: string,
  tag?: string
): Promise<PublicArticleListResponse> {
  const params: Record<string, string | number> = { page, pageSize };
  if (category) params.category = category;
  if (tag) params.tag = tag;

  const response = await http.get<PublicArticleListResponse>("/public/articles", {
    params,
  });
  return response.data;
}

export async function getPublicArticle(slug: string): Promise<Article> {
  const response = await http.get<Article>(`/public/articles/${slug}`);
  return response.data;
}

export async function getPublicCategories(): Promise<Category[]> {
  const response = await http.get<{ items: Category[] }>("/public/categories");
  return response.data.items || [];
}

export async function getPublicCategoryBySlug(slug: string, page: number = 1, pageSize: number = 10) {
  const response = await http.get(`/public/categories/${slug}`, {
    params: { page, pageSize },
  });
  return response.data;
}

export async function getPublicTags(): Promise<Tag[]> {
  const response = await http.get<{ items: Tag[] }>("/public/tags");
  return response.data.items || [];
}

export async function getPublicTagBySlug(slug: string, page: number = 1, pageSize: number = 10) {
  const response = await http.get(`/public/tags/${slug}`, {
    params: { page, pageSize },
  });
  return response.data;
}

// ---------- Admin Article APIs ----------

export async function getAdminArticles(
  page: number = 1,
  pageSize: number = 10,
  status?: string
): Promise<ArticleListResponse> {
  const params: Record<string, string | number> = { page, pageSize };
  if (status) params.status = status;

  const response = await http.get<ArticleListResponse>("/admin/articles", {
    params,

  });
  return response.data;
}

export async function getAdminArticle(id: number): Promise<Article> {
  const response = await http.get<Article>(`/admin/articles/${id}`, {

  });
  return response.data;
}

export async function createArticle(data: Partial<Article>): Promise<Article> {
  const response = await http.post<Article>("/admin/articles", data, {

  });
  return response.data;
}

export type UpdateArticleOptions = {
  /** Last-known server updatedAt for optimistic concurrency. */
  baseUpdatedAt?: string | null;
  /** Skip optimistic lock and overwrite. */
  force?: boolean;
};

export async function updateArticle(
  id: number,
  data: Partial<Article>,
  opts?: UpdateArticleOptions,
): Promise<Article> {
  const body: Record<string, unknown> = { ...data };
  if (opts?.baseUpdatedAt) body.baseUpdatedAt = opts.baseUpdatedAt;
  if (opts?.force) body.force = true;
  const response = await http.put<Article>(`/admin/articles/${id}`, body, {});
  return response.data;
}

export function isArticleVersionConflict(err: unknown): {
  conflict: boolean;
  currentUpdatedAt?: string;
  message?: string;
} {
  const ax = err as {
    response?: { status?: number; data?: { error?: { code?: string; currentUpdatedAt?: string; message?: string } } };
  };
  if (ax?.response?.status !== 409) return { conflict: false };
  const e = ax.response?.data?.error;
  if (e?.code && e.code !== "version_conflict") return { conflict: false };
  return {
    conflict: true,
    currentUpdatedAt: e?.currentUpdatedAt,
    message: e?.message,
  };
}

export async function deleteArticle(id: number): Promise<void> {
  await http.delete(`/admin/articles/${id}`, {

  });
}

// ---------- Article Version History ----------

export interface ArticleVersionListItem {
  id: number;
  articleId: number;
  version: number;
  action: string;
  summary: string;
  createdBy: number;
  createdAt: string;
  zhTitle?: string;
  enTitle?: string;
  status?: string;
}

export interface ArticleVersionDetail {
  id: number;
  articleId: number;
  version: number;
  snapshot: Record<string, unknown>;
  action: string;
  summary: string;
  createdBy: number;
  createdAt: string;
}

export async function listArticleVersions(
  id: number,
  page = 1,
  pageSize = 50,
): Promise<{ items: ArticleVersionListItem[]; total: number; page: number; pageSize: number }> {
  const response = await http.get(`/admin/articles/${id}/versions`, {
    params: { page, pageSize },
  });
  return response.data;
}

export async function getArticleVersion(
  id: number,
  version: number,
): Promise<ArticleVersionDetail> {
  const response = await http.get(`/admin/articles/${id}/versions/${version}`);
  return response.data;
}

export async function compareArticleVersions(
  id: number,
  left: number,
  right: number,
): Promise<{ left: ArticleVersionDetail; right: ArticleVersionDetail }> {
  const response = await http.get(`/admin/articles/${id}/versions/compare`, {
    params: { left, right },
  });
  return response.data;
}

// ---------- Category APIs ----------

export async function getCategories(): Promise<Category[]> {
  const response = await http.get<{ items: Category[] }>("/admin/categories", {

  });
  return response.data.items || [];
}

export async function getCategoryTree(): Promise<Category[]> {
  const response = await http.get<{ items: Category[] }>("/admin/categories/tree", {

  });
  return response.data.items || [];
}

export async function createCategory(data: Partial<Category>): Promise<Category> {
  const response = await http.post<Category>("/admin/categories", data, {

  });
  return response.data;
}

export async function updateCategory(id: number, data: Partial<Category>): Promise<Category> {
  const response = await http.put<Category>(`/admin/categories/${id}`, data, {

  });
  return response.data;
}

export async function deleteCategory(id: number): Promise<void> {
  await http.delete(`/admin/categories/${id}`, {

  });
}

// ---------- Tag APIs ----------

export async function getTags(): Promise<Tag[]> {
  const response = await http.get<{ items: Tag[] }>("/admin/tags", {

  });
  return response.data.items || [];
}

export async function createTag(data: Partial<Tag>): Promise<Tag> {
  const response = await http.post<Tag>("/admin/tags", data, {

  });
  return response.data;
}

export async function updateTag(id: number, data: Partial<Tag>): Promise<Tag> {
  const response = await http.put<Tag>(`/admin/tags/${id}`, data, {

  });
  return response.data;
}

export async function deleteTag(id: number): Promise<void> {
  await http.delete(`/admin/tags/${id}`, {

  });
}
