import { http } from "@/api/http";
import type { CommentContentType } from "../api";

export interface AdminComment {
  id: number;
  content: string;
  authorName: string;
  authorEmail?: string;
  authorRole?: string;
  contentType: string;
  contentId: number;
  parentId?: number | null;
  status: string;
  createdAt: string;
  /** Legacy snake_case from older API shapes */
  author_name?: string;
  author_email?: string;
  created_at?: string;
  article_id?: number;
  article_title?: string;
}

export interface AdminCommentListResponse {
  comments: AdminComment[];
  total: number;
  page: number;
  pageSize: number;
}

export async function getAdminComments(
  page: number,
  pageSize: number,
  status?: string,
): Promise<AdminCommentListResponse> {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(pageSize),
  });
  if (status) params.set("status", status);
  const { data } = await http.get<AdminCommentListResponse>(
    `/admin/comments?${params.toString()}`,
  );
  return data;
}

export async function approveComment(id: number): Promise<void> {
  await http.patch(`/admin/comments/${id}/status`, { status: "approved" });
}

export async function rejectComment(id: number): Promise<void> {
  await http.patch(`/admin/comments/${id}/status`, { status: "rejected" });
}

export async function deleteComment(id: number): Promise<void> {
  await http.delete(`/admin/comments/${id}`);
}

export async function adminAuthorComment(input: {
  content: string;
  parentId?: number;
  contentType?: CommentContentType;
  contentId?: number;
}): Promise<void> {
  await http.post("/admin/comments/reply", input);
}

/** @deprecated Use adminAuthorComment */
export async function adminReplyComment(input: {
  content: string;
  parentId: number;
}): Promise<void> {
  await adminAuthorComment(input);
}

export function adminCommentAuthorName(item: AdminComment): string {
  return item.authorName || item.author_name || "—";
}

export function adminCommentCreatedAt(item: AdminComment): string {
  return item.createdAt || item.created_at || "";
}
