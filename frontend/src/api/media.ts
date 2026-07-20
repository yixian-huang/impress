import { http } from "@/api/http";

export interface MediaItem {
  id: number;
  url: string;
  filename: string;
  mimeType: string;
  size: number;
  width?: number;
  height?: number;
  createdAt: string;
}

interface MediaListResponse {
  items: MediaItem[];
  total: number;
  page: number;
  pageSize: number;
}

export type UploadMediaOptions = {
  /** 0–100 progress while the request body is uploading */
  onProgress?: (percent: number) => void;
};

export async function uploadMedia(
  file: File | Blob,
  filename?: string,
  opts?: UploadMediaOptions,
): Promise<MediaItem> {
  const formData = new FormData();
  formData.append("file", file, filename || (file instanceof File ? file.name : "upload.jpg"));

  const response = await http.post<MediaItem>("/admin/media/upload", formData, {
    onUploadProgress: (event) => {
      if (!opts?.onProgress) return;
      const total = event.total ?? 0;
      if (total <= 0) {
        opts.onProgress(0);
        return;
      }
      opts.onProgress(Math.min(100, Math.round((event.loaded / total) * 100)));
    },
  });
  return response.data;
}

export async function listMedia(page: number = 1, pageSize: number = 20, type?: string): Promise<MediaListResponse> {
  const params: Record<string, any> = { page, pageSize };
  if (type) params.type = type;
  const response = await http.get<MediaListResponse>("/admin/media", {
    params,

  });
  return response.data;
}

export async function deleteMedia(id: number): Promise<void> {
  await http.delete(`/admin/media/${id}`, {

  });
}

export async function recropMedia(id: number, file: Blob): Promise<MediaItem> {
  const formData = new FormData();
  formData.append("file", file, "recropped.jpg");
  const response = await http.put<MediaItem>(`/admin/media/${id}/crop`, formData, {

  });
  return response.data;
}

export interface MediaUsage {
  type: "article" | "page" | "content_document";
  id: string;
  title: string;
  field: string;
}

export async function getMediaUsages(id: number): Promise<MediaUsage[]> {
  const response = await http.get<{ usages: MediaUsage[] }>(`/admin/media/${id}/usages`, {

  });
  return response.data.usages || [];
}

export async function renameMedia(id: number, filename: string): Promise<MediaItem> {
  const response = await http.put<MediaItem>(`/admin/media/${id}`, { filename }, {

  });
  return response.data;
}
