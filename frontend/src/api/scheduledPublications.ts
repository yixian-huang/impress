import { http } from "@/api/http";

export type ScheduledPublicationResourceType = "article" | "page";

export type ScheduledPublicationStatus =
  | "pending"
  | "running"
  | "succeeded"
  | "failed"
  | "cancelled";

export interface ScheduledPublication {
  id: number;
  resourceType: ScheduledPublicationResourceType;
  resourceId: number;
  title?: string;
  slug?: string;
  status: ScheduledPublicationStatus;
  scheduledAt: string;
  attempts?: number;
  lastError?: string | null;
  expectedVersion?: number | null;
  createdAt: string;
  updatedAt: string;
  completedAt?: string | null;
}

export interface ScheduledPublicationListResponse {
  items: ScheduledPublication[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ListScheduledPublicationsParams {
  page?: number;
  pageSize?: number;
  status?: ScheduledPublicationStatus | "";
  resourceType?: ScheduledPublicationResourceType | "";
  resourceId?: number;
}

export interface CreateScheduledPublicationRequest {
  resourceType: ScheduledPublicationResourceType;
  resourceId: number;
  scheduledAt: string;
  expectedVersion?: number;
  publishPayload?: Record<string, unknown>;
}

export interface UpdateScheduledPublicationRequest {
  scheduledAt?: string;
  expectedVersion?: number;
  publishPayload?: Record<string, unknown>;
}

function buildParams(params: ListScheduledPublicationsParams = {}) {
  const searchParams = new URLSearchParams();
  if (params.page) searchParams.set("page", String(params.page));
  if (params.pageSize) searchParams.set("pageSize", String(params.pageSize));
  if (params.status) searchParams.set("status", params.status);
  if (params.resourceType) searchParams.set("resourceType", params.resourceType);
  if (params.resourceId) searchParams.set("resourceId", String(params.resourceId));
  return searchParams;
}

export async function listScheduledPublications(
  params: ListScheduledPublicationsParams = {},
): Promise<ScheduledPublicationListResponse> {
  const searchParams = buildParams(params);
  const query = searchParams.toString();
  const response = await http.get<ScheduledPublicationListResponse>(
    `/admin/scheduled-publications${query ? `?${query}` : ""}`,
  );
  return response.data;
}

export async function getResourceScheduledPublication(
  resourceType: ScheduledPublicationResourceType,
  resourceId: number,
): Promise<ScheduledPublication | null> {
  const [pending, running, failed] = await Promise.all(
    (["pending", "running", "failed"] as const).map((status) =>
      listScheduledPublications({
        resourceType,
        resourceId,
        status,
        page: 1,
        pageSize: 1,
      })
    ),
  );
  return pending.items[0] ?? running.items[0] ?? failed.items[0] ?? null;
}

export async function createScheduledPublication(
  data: CreateScheduledPublicationRequest,
): Promise<ScheduledPublication> {
  const response = await http.post<ScheduledPublication>("/admin/scheduled-publications", data);
  return response.data;
}

export async function updateScheduledPublication(
  id: number,
  data: UpdateScheduledPublicationRequest,
): Promise<ScheduledPublication> {
  const response = await http.put<ScheduledPublication>(`/admin/scheduled-publications/${id}`, data);
  return response.data;
}

export async function cancelScheduledPublication(id: number): Promise<ScheduledPublication> {
  const response = await http.delete<ScheduledPublication>(`/admin/scheduled-publications/${id}`);
  return response.data;
}

export async function retryScheduledPublication(id: number): Promise<ScheduledPublication> {
  const response = await http.post<ScheduledPublication>(`/admin/scheduled-publications/${id}/retry`);
  return response.data;
}

export function toDateTimeLocalValue(value: string | null | undefined): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value.slice(0, 16);
  const offsetMs = date.getTimezoneOffset() * 60000;
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
}

export function dateTimeLocalToISOString(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    throw new Error("无效的发布时间");
  }
  return date.toISOString();
}

export function futureDateTimeLocalValue(minutesFromNow = 5): string {
  return toDateTimeLocalValue(new Date(Date.now() + minutesFromNow * 60000).toISOString());
}
