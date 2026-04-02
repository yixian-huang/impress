import { http } from "./http";

export interface SiteDTO {
  id: number;
  name: string;
  domain: string;
  sub_path: string;
  locale: string;
  theme_id: number;
  mode: string;
  status: string;
  created_at: string;
}

export interface SiteListResponse {
  items: SiteDTO[];
  total: number;
}

export interface CreateSiteRequest {
  name: string;
  domain: string;
  sub_path: string;
  locale: string;
  mode: string;
  status: string;
}

export interface UpdateSiteRequest {
  name?: string;
  domain?: string;
  sub_path?: string;
  locale?: string;
  mode?: string;
  status?: string;
}

export async function listSites() {
  const res = await http.get<SiteListResponse>("/admin/sites", {

  });
  return res.data;
}

export async function createSite(data: CreateSiteRequest) {
  const res = await http.post<SiteDTO>("/admin/sites", data, {

  });
  return res.data;
}

export async function updateSite(id: number, data: UpdateSiteRequest) {
  const res = await http.put<SiteDTO>(`/admin/sites/${id}`, data, {

  });
  return res.data;
}

export async function deleteSite(id: number) {
  await http.delete(`/admin/sites/${id}`, {

  });
}

export async function exportSite(id: number) {
  const res = await http.get(`/admin/sites/${id}/export`, {

    responseType: "blob",
  });
  return res.data;
}

export async function importSite(file: File) {
  const formData = new FormData();
  formData.append("file", file);
  const res = await http.post("/admin/sites/import", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
  return res.data;
}
