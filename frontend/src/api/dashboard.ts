import { http } from "@/api/http";

export type AdminDashboardSummary = {
  todayVisits: number;
  pagesCount: number;
  articlesCount: number;
  mediaCount: number;
  errors?: Record<string, boolean>;
};

export async function getAdminDashboardSummary(): Promise<AdminDashboardSummary> {
  const response = await http.get<AdminDashboardSummary>("/admin/dashboard/summary");
  return response.data;
}
