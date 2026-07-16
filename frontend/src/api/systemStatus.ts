import { http } from "@/api/http";

export interface SystemApplicationInfo {
  version: string;
}

export interface SystemRuntimeInfo {
  goVersion: string;
  os: string;
  arch: string;
  cpuCount: number;
  goroutines: number;
  uptime: number;
}

export interface SystemMemoryInfo {
  allocMB: number;
  totalAllocMB: number;
  sysMB: number;
  gcPauseMs: number;
}

export interface SystemDatabaseInfo {
  type: string;
  healthy: boolean;
  status: string;
  error?: string;
  openConnections: number;
  maxOpenConnections: number;
  inUse: number;
  idle: number;
}

export interface SystemStorageInfo {
  type: string;
  healthy: boolean;
  status: string;
  error?: string;
  uploadDirSizeMB: number;
  uploadDirBytes: number;
  mediaCount: number;
}

export interface SystemContentCounts {
  articles: number;
  pages: number;
  media: number;
  users: number;
}

export interface SystemStatusResponse {
  application: SystemApplicationInfo;
  runtime: SystemRuntimeInfo;
  memory: SystemMemoryInfo;
  database: SystemDatabaseInfo;
  storage: SystemStorageInfo;
  content: SystemContentCounts;
}

export async function getSystemStatus(): Promise<SystemStatusResponse> {
  const response = await http.get<SystemStatusResponse>("/admin/system/status");
  return response.data;
}
