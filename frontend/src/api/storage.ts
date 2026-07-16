import { http } from "./http";

export interface StorageConfig {
  strategy: string;
  bucket: string;
  region: string;
  endpoint: string;
  accessKey: string;
  hasSecretKey: boolean;
  basePath: string;
  updatedAt: string;
}

export interface UpdateStorageConfigRequest {
  strategy: string;
  bucket?: string;
  region?: string;
  endpoint?: string;
  accessKey?: string;
  secretKey?: string;
  basePath?: string;
}

export interface StorageTestResult {
  success: boolean;
  message: string;
}

export async function getStorageConfig() {
  const res = await http.get<StorageConfig>("/admin/storage/config", {

  });
  return res.data;
}

export async function updateStorageConfig(data: UpdateStorageConfigRequest) {
  const res = await http.put<StorageConfig>("/admin/storage/config", data, {

  });
  return res.data;
}

export async function testStorageConnection() {
  const res = await http.post<StorageTestResult>("/admin/storage/test", {}, {

  });
  return res.data;
}
