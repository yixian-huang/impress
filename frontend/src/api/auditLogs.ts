import { http } from "@/api/http";

export interface AuditEvent {
  id: number;
  action: string;
  actor: string;
  resource: string;
  result: string;
  details: string;
  createdAt: string;
}

export interface AuditLogListResponse {
  items: AuditEvent[];
  total: number;
  page: number;
  pageSize: number;
}

export interface AuditLogFilters {
  action?: string;
  actor?: string;
  from?: string;
  to?: string;
}

export async function getAuditLogs(
  page: number = 1,
  pageSize: number = 20,
  filters: AuditLogFilters = {}
): Promise<AuditLogListResponse> {
  const response = await http.get<AuditLogListResponse>("/admin/audit-logs", {
    params: { page, pageSize, ...filters },

  });
  return response.data;
}
