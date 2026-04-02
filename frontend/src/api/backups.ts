import { http } from "@/api/http";

export interface BackupRecord {
  id: number;
  filename: string;
  size: number;
  createdAt: string;
}

export interface ExportRecord {
  filename: string;
  size: number;
  createdAt: string;
}

export interface TableExportInfo {
  count: number;
}

export interface ValidationResult {
  valid: boolean;
  version: string;
  exportedAt: string;
  tables: Record<string, TableExportInfo>;
  mediaFiles: number;
  errors?: string[];
}

export async function getBackups(): Promise<BackupRecord[]> {
  const response = await http.get<{ items: BackupRecord[] }>("/admin/backups", {

  });
  return response.data.items || [];
}

export async function triggerBackup(): Promise<BackupRecord> {
  const response = await http.post<BackupRecord>("/admin/backups", null, {

  });
  return response.data;
}

export async function triggerExport(): Promise<ExportRecord> {
  const response = await http.post<ExportRecord>("/admin/backups/export", null, {

  });
  return response.data;
}

export async function downloadExport(filename: string): Promise<void> {
  const response = await http.get(`/admin/backups/export/${filename}`, {

    responseType: "blob",
  });
  const url = window.URL.createObjectURL(new Blob([response.data]));
  const link = document.createElement("a");
  link.href = url;
  link.setAttribute("download", filename);
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(url);
}

export async function validateImport(file: File): Promise<ValidationResult> {
  const formData = new FormData();
  formData.append("file", file);
  const response = await http.post<ValidationResult>("/admin/backups/import/validate", formData, {

  });
  return response.data;
}

export async function runImport(file: File): Promise<{ message: string }> {
  const formData = new FormData();
  formData.append("file", file);
  const response = await http.post<{ message: string }>("/admin/backups/import", formData, {

    timeout: 600000, // 10 min timeout for large imports
  });
  return response.data;
}
