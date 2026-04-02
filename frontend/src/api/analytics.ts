import { http } from "@/api/http";

export interface PageViewStats {
  pageKey: string;
  today: number;
  last7d: number;
  last30d: number;
}

export interface AnalyticsSummary {
  pages: PageViewStats[];
  totals: {
    today: number;
    last7d: number;
    last30d: number;
  };
}

export async function getAnalyticsSummary(): Promise<AnalyticsSummary> {
  const response = await http.get<AnalyticsSummary>("/admin/analytics/summary", {

  });
  return response.data;
}
