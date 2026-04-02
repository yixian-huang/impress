import { http } from "./http";

export interface PageTemplate {
  id: number;
  key: string;
  nameZh: string;
  nameEn: string;
  descriptionZh: string;
  descriptionEn: string;
  category: "builtin" | "custom" | "theme";
  config: Record<string, unknown>;
  thumbnail?: string;
  createdAt: string;
  updatedAt: string;
}

export const listTemplates = (category?: string) => {
  const params = category ? `?category=${category}` : "";
  return http.get<PageTemplate[]>(`/admin/templates${params}`).then((r) => r.data);
};

export const createTemplate = (data: Omit<PageTemplate, "id" | "category" | "createdAt" | "updatedAt">) =>
  http.post<PageTemplate>("/admin/templates", data).then((r) => r.data);

export const updateTemplate = (id: number, data: Partial<PageTemplate>) =>
  http.put<PageTemplate>(`/admin/templates/${id}`, data).then((r) => r.data);

export const deleteTemplate = (id: number) =>
  http.delete(`/admin/templates/${id}`);

export const duplicateTemplate = (id: number) =>
  http.post<PageTemplate>(`/admin/templates/${id}/duplicate`, {}).then((r) => r.data);
