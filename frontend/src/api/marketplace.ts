import { http } from "@/api/http";

export interface MarketplaceItem {
  id: number;
  type: "plugin" | "theme";
  name: string;
  slug: string;
  description: string;
  author: string;
  version: string;
  icon_url: string;
  downloads: number;
  category: string;
  status: "available" | "installed" | "update_available";
}

export interface InstalledItem {
  id: number;
  type: "plugin" | "theme";
  name: string;
  slug: string;
  version: string;
  status: "active" | "inactive" | "error";
  installed_at: string;
}

export interface MarketplaceListResponse {
  items: MarketplaceItem[];
  page: number;
  pageSize: number;
  total: number;
}

export interface MarketplaceItemDetail extends MarketplaceItem {
  versions: { version: string; released_at: string; changelog: string }[];
  readme: string;
}

export async function getMarketplaceItems(params: {
  type?: string;
  category?: string;
  search?: string;
  page?: number;
  pageSize?: number;
}): Promise<MarketplaceListResponse> {
  const response = await http.get<MarketplaceListResponse>("/admin/marketplace/items", {

    params,
  });
  return response.data;
}

export async function getInstalledItems(): Promise<InstalledItem[]> {
  const response = await http.get<{ items: InstalledItem[] }>("/admin/marketplace/installed", {

  });
  return response.data.items || [];
}

export async function installItem(slug: string): Promise<{ message: string }> {
  const response = await http.post<{ message: string }>(
    `/admin/marketplace/items/${slug}/install`,
    null,
  );
  return response.data;
}

export async function uninstallItem(slug: string): Promise<{ message: string }> {
  const response = await http.delete<{ message: string }>(
    `/admin/marketplace/items/${slug}`,
  );
  return response.data;
}

export async function getMarketplaceItemDetail(slug: string): Promise<MarketplaceItemDetail> {
  const response = await http.get<MarketplaceItemDetail>(
    `/admin/marketplace/items/${slug}`,
  );
  return response.data;
}
