import { http } from "./http";

// JSON config type — matches backend JSONMap
type JSONMap = Record<string, unknown>;

export interface UnifiedPageItem {
  id: number;
  slug: string;
  zhTitle: string;
  enTitle: string;
  mode: "template" | "composable";
  templateId?: number;
  status: string;
  sortOrder: number;
  showInNav: boolean;
  parentId?: number;
  publishedVersion: number;
  draftVersion: number;
  createdAt: string;
  updatedAt: string;
}

export interface UnifiedPageDraft {
  id: number;
  slug: string;
  config: JSONMap;
  version: number;
  publishedVersion: number;
  updatedAt: string;
}

export interface CreateUnifiedPageRequest {
  slug: string;
  zhTitle?: string;
  enTitle?: string;
  mode: "template" | "composable";
  templateId?: number;
  draftConfig?: JSONMap;
  sortOrder?: number;
  showInNav?: boolean;
  parentId?: number;
}

// Admin CRUD
export const listUnifiedPages = (status?: string, mode?: string) => {
  const params = new URLSearchParams();
  if (status) params.set("status", status);
  if (mode) params.set("mode", mode);
  return http.get<UnifiedPageItem[]>(`/admin/unified-pages?${params}`).then((r) => r.data);
};

export const getUnifiedPage = (id: number) =>
  http.get<UnifiedPageItem>(`/admin/unified-pages/${id}`).then((r) => r.data);

export const createUnifiedPage = (data: CreateUnifiedPageRequest) =>
  http.post<UnifiedPageItem>("/admin/unified-pages", data).then((r) => r.data);

export const deleteUnifiedPage = (id: number) =>
  http.delete(`/admin/unified-pages/${id}`);

// Draft
export const getUnifiedPageDraft = (id: number) =>
  http.get<UnifiedPageDraft>(`/admin/unified-pages/${id}/draft`).then((r) => r.data);

export const updateUnifiedPageDraft = (id: number, version: number, config: JSONMap) =>
  http.put(`/admin/unified-pages/${id}/draft`, { config }, {
    headers: { "If-Match": String(version) },
  }).then((r) => r.data);

// Publish / Unpublish / Rollback
export const publishUnifiedPage = (id: number, expectedDraftVersion: number) =>
  http.post(`/admin/unified-pages/${id}/publish`, { expectedDraftVersion }).then((r) => r.data);

export const unpublishUnifiedPage = (id: number) =>
  http.post(`/admin/unified-pages/${id}/unpublish`).then((r) => r.data);

export const rollbackUnifiedPage = (id: number, targetVersion: number) =>
  http.post(`/admin/unified-pages/${id}/rollback`, { targetVersion }).then((r) => r.data);

// Version history
export const listUnifiedPageVersions = (id: number, page = 1, pageSize = 20) =>
  http.get(`/admin/unified-pages/${id}/versions?page=${page}&pageSize=${pageSize}`).then((r) => r.data);

export const getUnifiedPageVersion = (id: number, version: number) =>
  http.get(`/admin/unified-pages/${id}/versions/${version}`).then((r) => r.data);
