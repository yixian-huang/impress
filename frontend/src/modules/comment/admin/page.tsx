import { useState, useEffect, useCallback, Fragment } from "react";
import {
  AdminButton,
  AdminErrorBanner,
  AdminLoading,
  AdminPageHeader,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import AdminCommentReplyPanel from "./AdminCommentReplyPanel";
import {
  adminCommentAuthorName,
  adminCommentCreatedAt,
  approveComment,
  deleteComment,
  getAdminComments,
  rejectComment,
  type AdminComment,
} from "./api";

type StatusFilter = "" | "pending" | "approved" | "rejected";

const STATUS_TABS: { label: string; value: StatusFilter }[] = [
  { label: "全部", value: "" },
  { label: "待审核", value: "pending" },
  { label: "已通过", value: "approved" },
  { label: "已拒绝", value: "rejected" },
];

const STATUS_BADGE: Record<string, { label: string; className: string }> = {
  pending: { label: "待审核", className: "bg-yellow-100 text-yellow-800" },
  approved: { label: "已通过", className: "bg-green-100 text-green-800" },
  rejected: { label: "已拒绝", className: "bg-red-100 text-red-800" },
};

const PAGE_SIZE = 20;

export default function AdminCommentsPage() {
  useDocumentTitle("评论管理");
  const [data, setData] = useState<Awaited<ReturnType<typeof getAdminComments>> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("");
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [actionLoading, setActionLoading] = useState(false);
  const [replyTarget, setReplyTarget] = useState<AdminComment | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await getAdminComments(page, PAGE_SIZE, statusFilter || undefined);
      setData(result);
    } catch {
      setError("获取评论列表失败，请稍后重试");
    } finally {
      setLoading(false);
    }
  }, [page, statusFilter]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const totalPages = data ? Math.ceil(data.total / PAGE_SIZE) : 0;

  const handleTabChange = (value: StatusFilter) => {
    setStatusFilter(value);
    setPage(1);
    setSelectedIds(new Set());
  };

  const handleToggleSelect = (id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const handleSelectAll = () => {
    if (!data) return;
    if (selectedIds.size === data.comments.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(data.comments.map((item) => item.id)));
    }
  };

  const handleApprove = async (id: number) => {
    setActionLoading(true);
    try {
      await approveComment(id);
      await fetchData();
    } catch {
      setError("操作失败，请稍后重试");
    } finally {
      setActionLoading(false);
    }
  };

  const handleReject = async (id: number) => {
    setActionLoading(true);
    try {
      await rejectComment(id);
      await fetchData();
    } catch {
      setError("操作失败，请稍后重试");
    } finally {
      setActionLoading(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm("确认删除此评论？此操作不可撤销。")) return;
    setActionLoading(true);
    try {
      await deleteComment(id);
      setSelectedIds((prev) => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
      await fetchData();
    } catch {
      setError("删除失败，请稍后重试");
    } finally {
      setActionLoading(false);
    }
  };

  const handleBulkApprove = async () => {
    if (selectedIds.size === 0) return;
    setActionLoading(true);
    try {
      await Promise.all(Array.from(selectedIds).map((id) => approveComment(id)));
      setSelectedIds(new Set());
      await fetchData();
    } catch {
      setError("批量操作失败，请稍后重试");
    } finally {
      setActionLoading(false);
    }
  };

  const handleBulkReject = async () => {
    if (selectedIds.size === 0) return;
    setActionLoading(true);
    try {
      await Promise.all(Array.from(selectedIds).map((id) => rejectComment(id)));
      setSelectedIds(new Set());
      await fetchData();
    } catch {
      setError("批量操作失败，请稍后重试");
    } finally {
      setActionLoading(false);
    }
  };

  const handleBulkDelete = async () => {
    if (selectedIds.size === 0) return;
    if (!window.confirm(`确认删除选中的 ${selectedIds.size} 条评论？此操作不可撤销。`)) return;
    setActionLoading(true);
    try {
      await Promise.all(Array.from(selectedIds).map((id) => deleteComment(id)));
      setSelectedIds(new Set());
      await fetchData();
    } catch {
      setError("批量删除失败，请稍后重试");
    } finally {
      setActionLoading(false);
    }
  };

  const formatTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString("zh-CN");
  };

  const truncate = (text: string, maxLen = 80) => {
    if (text.length <= maxLen) return text;
    return text.slice(0, maxLen) + "…";
  };

  return (
    <div>
      <AdminPageHeader
        title="评论管理"
        description="审核与回复访客评论"
        actions={
          <AdminButton size="sm" onClick={fetchData} disabled={loading}>
            {loading ? "加载中…" : "刷新"}
          </AdminButton>
        }
      />

      <div className="mb-6 flex gap-1 border-b border-gray-200">
        {STATUS_TABS.map((tab) => {
          const isActive = statusFilter === tab.value;
          return (
            <button
              key={tab.value}
              type="button"
              onClick={() => handleTabChange(tab.value)}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                isActive
                  ? "border-blue-600 text-blue-600"
                  : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
              }`}
            >
              {tab.label}
            </button>
          );
        })}
      </div>

      {selectedIds.size > 0 && (
        <div className="mb-4 flex items-center gap-3 p-3 bg-blue-50 border border-blue-200 rounded-md">
          <span className="text-sm text-blue-700">已选 {selectedIds.size} 项</span>
          <button
            type="button"
            onClick={handleBulkApprove}
            disabled={actionLoading}
            className="px-3 py-1 text-xs font-medium text-green-700 bg-white border border-green-300 rounded-md hover:bg-green-50 disabled:opacity-50"
          >
            批量通过
          </button>
          <button
            type="button"
            onClick={handleBulkReject}
            disabled={actionLoading}
            className="px-3 py-1 text-xs font-medium text-yellow-700 bg-white border border-yellow-300 rounded-md hover:bg-yellow-50 disabled:opacity-50"
          >
            批量拒绝
          </button>
          <button
            type="button"
            onClick={handleBulkDelete}
            disabled={actionLoading}
            className="px-3 py-1 text-xs font-medium text-red-700 bg-white border border-red-300 rounded-md hover:bg-red-50 disabled:opacity-50"
          >
            批量删除
          </button>
        </div>
      )}

      {error && <AdminErrorBanner message={error} onDismiss={() => setError(null)} />}

      {loading && !data ? (
        <AdminLoading />
      ) : data ? (
        <>
          <div className="bg-white shadow rounded-lg overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 py-3 text-left">
                    <input
                      type="checkbox"
                      checked={data.comments.length > 0 && selectedIds.size === data.comments.length}
                      onChange={handleSelectAll}
                      className="rounded border-gray-300"
                    />
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    作者
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    内容
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    文章
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    状态
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    时间
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {data.comments.map((item: AdminComment) => {
                  const badge = STATUS_BADGE[item.status] || { label: item.status, className: "bg-gray-100 text-gray-800" };
                  return (
                    <Fragment key={item.id}>
                      <tr className="hover:bg-gray-50">
                        <td className="px-4 py-4">
                          <input
                            type="checkbox"
                            checked={selectedIds.has(item.id)}
                            onChange={() => handleToggleSelect(item.id)}
                            className="rounded border-gray-300"
                          />
                        </td>
                        <td className="px-4 py-4 text-sm text-gray-900">
                          <div className="font-medium">
                            {adminCommentAuthorName(item)}
                            {item.authorRole === "author" && (
                              <span className="ml-1 text-xs text-blue-600">作者</span>
                            )}
                          </div>
                          {(item.authorEmail || item.author_email) && (
                            <div className="text-xs text-gray-500 mt-0.5">
                              {item.authorEmail || item.author_email}
                            </div>
                          )}
                        </td>
                        <td className="px-4 py-4 text-sm text-gray-700 max-w-xs">
                          {truncate(item.content)}
                        </td>
                        <td className="px-4 py-4 text-sm text-gray-500">
                          {item.article_title ? (
                            <span className="text-xs">{truncate(item.article_title, 40)}</span>
                          ) : item.article_id ? (
                            <span className="text-xs text-gray-400">文章 #{item.article_id}</span>
                          ) : (
                            <span className="text-xs text-gray-400">—</span>
                          )}
                        </td>
                        <td className="px-4 py-4 text-sm">
                          <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${badge.className}`}>
                            {badge.label}
                          </span>
                        </td>
                        <td className="px-4 py-4 text-sm text-gray-500 whitespace-nowrap">
                          {formatTime(adminCommentCreatedAt(item))}
                        </td>
                        <td className="px-4 py-4 text-sm">
                          <div className="flex items-center gap-2 flex-wrap">
                            <button
                              type="button"
                              onClick={() => setReplyTarget(item)}
                              disabled={actionLoading}
                              className="text-blue-600 hover:text-blue-800 text-xs font-medium disabled:opacity-50"
                            >
                              回复
                            </button>
                            {item.status !== "approved" && (
                              <button
                                type="button"
                                onClick={() => handleApprove(item.id)}
                                disabled={actionLoading}
                                className="text-green-600 hover:text-green-800 text-xs font-medium disabled:opacity-50"
                              >
                                通过
                              </button>
                            )}
                            {item.status !== "rejected" && (
                              <button
                                type="button"
                                onClick={() => handleReject(item.id)}
                                disabled={actionLoading}
                                className="text-yellow-600 hover:text-yellow-800 text-xs font-medium disabled:opacity-50"
                              >
                                拒绝
                              </button>
                            )}
                            <button
                              type="button"
                              onClick={() => handleDelete(item.id)}
                              disabled={actionLoading}
                              className="text-red-600 hover:text-red-800 text-xs font-medium disabled:opacity-50"
                            >
                              删除
                            </button>
                          </div>
                        </td>
                      </tr>
                    </Fragment>
                  );
                })}
                {data.comments.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-6 py-8 text-center text-sm text-gray-500">
                      暂无评论记录
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="mt-4 flex items-center justify-between">
              <p className="text-sm text-gray-500">
                共 {data.total} 条，第 {page}/{totalPages} 页
              </p>
              <div className="flex gap-2">
                <button
                  type="button"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  上一页
                </button>
                <button
                  type="button"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page >= totalPages}
                  className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  下一页
                </button>
              </div>
            </div>
          )}
        </>
      ) : null}

      {replyTarget && (
        <AdminCommentReplyPanel
          comment={replyTarget}
          onClose={() => setReplyTarget(null)}
          onSent={fetchData}
        />
      )}
    </div>
  );
}
