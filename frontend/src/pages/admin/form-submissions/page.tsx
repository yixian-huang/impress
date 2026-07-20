import { useState, useEffect, useCallback, Fragment } from "react";
import {
  getFormSubmissions,
  getSubmissionCounts,
  updateSubmissionStatus,
  bulkUpdateStatus,
  deleteFormSubmission,
} from "@/api/formSubmissions";
import type { FormSubmission, FormSubmissionListResponse } from "@/api/formSubmissions";
import {
  AdminButton,
  AdminErrorBanner,
  AdminLoading,
  AdminPageHeader,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

type StatusFilter = "" | "unread" | "read" | "archived";

const STATUS_TABS: { label: string; value: StatusFilter }[] = [
  { label: "全部", value: "" },
  { label: "未读", value: "unread" },
  { label: "已读", value: "read" },
  { label: "已归档", value: "archived" },
];

const STATUS_BADGE: Record<string, { label: string; className: string }> = {
  unread: { label: "未读", className: "bg-blue-100 text-blue-800" },
  read: { label: "已读", className: "bg-gray-100 text-gray-800" },
  archived: { label: "已归档", className: "bg-yellow-100 text-yellow-800" },
};

const PAGE_SIZE = 20;

export default function AdminFormSubmissionsPage() {
  useDocumentTitle("表单提交");
  const [data, setData] = useState<FormSubmissionListResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("");
  const [counts, setCounts] = useState<Record<string, number>>({});
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [expandedId, setExpandedId] = useState<number | null>(null);

  const fetchCounts = useCallback(async () => {
    try {
      const result = await getSubmissionCounts();
      setCounts(result.counts);
    } catch {
      // silently ignore count errors
    }
  }, []);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await getFormSubmissions(
        page,
        PAGE_SIZE,
        undefined,
        statusFilter || undefined
      );
      setData(result);
    } catch {
      setError("获取表单提交列表失败，请稍后重试");
    } finally {
      setLoading(false);
    }
  }, [page, statusFilter]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    fetchCounts();
  }, [fetchCounts]);

  const totalPages = data ? Math.ceil(data.total / PAGE_SIZE) : 0;

  const handleTabChange = (value: StatusFilter) => {
    setStatusFilter(value);
    setPage(1);
    setSelectedIds(new Set());
    setExpandedId(null);
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
    if (selectedIds.size === data.items.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(data.items.map((item) => item.id)));
    }
  };

  const handleStatusChange = async (id: number, status: "unread" | "read" | "archived") => {
    try {
      await updateSubmissionStatus(id, status);
      await fetchData();
      await fetchCounts();
    } catch {
      setError("更新状态失败");
    }
  };

  const handleBulkStatus = async (status: "unread" | "read" | "archived") => {
    if (selectedIds.size === 0) return;
    try {
      await bulkUpdateStatus(Array.from(selectedIds), status);
      setSelectedIds(new Set());
      await fetchData();
      await fetchCounts();
    } catch {
      setError("批量更新状态失败");
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm("确认删除此提交记录？此操作不可撤销。")) return;
    try {
      await deleteFormSubmission(id);
      await fetchData();
      await fetchCounts();
    } catch {
      setError("删除失败");
    }
  };

  const handleToggleExpand = (id: number) => {
    setExpandedId((prev) => (prev === id ? null : id));
  };

  const formatTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleString("zh-CN");
  };

  return (
    <div>
      <AdminPageHeader
        title="表单提交"
        description="查看与处理站点表单线索"
        actions={
          <AdminButton
            size="sm"
            onClick={() => {
              fetchData();
              fetchCounts();
            }}
            disabled={loading}
          >
            {loading ? "加载中…" : "刷新"}
          </AdminButton>
        }
      />

      {/* Status Tabs */}
      <div className="mb-6 flex gap-1 border-b border-gray-200">
        {STATUS_TABS.map((tab) => {
          const isActive = statusFilter === tab.value;
          const unreadCount = tab.value === "unread" ? (counts.unread || 0) : 0;
          return (
            <button
              key={tab.value}
              onClick={() => handleTabChange(tab.value)}
              className={`relative px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                isActive
                  ? "border-blue-600 text-blue-600"
                  : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
              }`}
            >
              {tab.label}
              {tab.value === "unread" && unreadCount > 0 && (
                <span className="ml-1.5 inline-flex items-center justify-center px-1.5 py-0.5 text-xs font-bold leading-none text-white bg-red-500 rounded-full">
                  {unreadCount}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {/* Bulk Actions */}
      {selectedIds.size > 0 && (
        <div className="mb-4 flex items-center gap-3 p-3 bg-blue-50 border border-blue-200 rounded-md">
          <span className="text-sm text-blue-700">
            已选 {selectedIds.size} 项
          </span>
          <button
            onClick={() => handleBulkStatus("read")}
            className="px-3 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            标为已读
          </button>
          <button
            onClick={() => handleBulkStatus("unread")}
            className="px-3 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            标为未读
          </button>
          <button
            onClick={() => handleBulkStatus("archived")}
            className="px-3 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            归档
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
                      checked={data.items.length > 0 && selectedIds.size === data.items.length}
                      onChange={handleSelectAll}
                      className="rounded border-gray-300"
                    />
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    姓名
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    邮箱
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    类型
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    状态
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    提交时间
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {data.items.map((item: FormSubmission) => {
                  const badge = STATUS_BADGE[item.status] || STATUS_BADGE.unread;
                  const isExpanded = expandedId === item.id;
                  return (
                    <Fragment key={item.id}>
                      <tr
                        className={`hover:bg-gray-50 cursor-pointer ${
                          item.status === "unread" ? "font-medium" : ""
                        }`}
                        onClick={() => handleToggleExpand(item.id)}
                      >
                        <td className="px-4 py-4" onClick={(e) => e.stopPropagation()}>
                          <input
                            type="checkbox"
                            checked={selectedIds.has(item.id)}
                            onChange={() => handleToggleSelect(item.id)}
                            className="rounded border-gray-300"
                          />
                        </td>
                        <td className="px-4 py-4 text-sm text-gray-900">
                          {item.name}
                        </td>
                        <td className="px-4 py-4 text-sm text-gray-700">
                          {item.email}
                        </td>
                        <td className="px-4 py-4 text-sm text-gray-700">
                          {item.formType}
                        </td>
                        <td className="px-4 py-4 text-sm">
                          <span className={`inline-flex px-2 py-0.5 rounded-full text-xs font-medium ${badge.className}`}>
                            {badge.label}
                          </span>
                        </td>
                        <td className="px-4 py-4 text-sm text-gray-500 whitespace-nowrap">
                          {formatTime(item.createdAt)}
                        </td>
                        <td className="px-4 py-4 text-sm" onClick={(e) => e.stopPropagation()}>
                          <div className="flex items-center gap-2">
                            {item.status === "unread" ? (
                              <button
                                onClick={() => handleStatusChange(item.id, "read")}
                                className="text-blue-600 hover:text-blue-800 text-xs"
                                title="标为已读"
                              >
                                已读
                              </button>
                            ) : item.status === "read" ? (
                              <button
                                onClick={() => handleStatusChange(item.id, "unread")}
                                className="text-blue-600 hover:text-blue-800 text-xs"
                                title="标为未读"
                              >
                                未读
                              </button>
                            ) : null}
                            {item.status !== "archived" && (
                              <button
                                onClick={() => handleStatusChange(item.id, "archived")}
                                className="text-yellow-600 hover:text-yellow-800 text-xs"
                                title="归档"
                              >
                                归档
                              </button>
                            )}
                            <button
                              onClick={() => handleDelete(item.id)}
                              className="text-red-600 hover:text-red-800 text-xs"
                              title="删除"
                            >
                              删除
                            </button>
                          </div>
                        </td>
                      </tr>
                      {isExpanded && (
                        <tr>
                          <td colSpan={7} className="px-4 py-4 bg-gray-50">
                            <div className="space-y-2 text-sm">
                              {item.phone && (
                                <div>
                                  <span className="font-medium text-gray-700">电话：</span>
                                  <span className="text-gray-600">{item.phone}</span>
                                </div>
                              )}
                              {item.company && (
                                <div>
                                  <span className="font-medium text-gray-700">公司：</span>
                                  <span className="text-gray-600">{item.company}</span>
                                </div>
                              )}
                              {item.message && (
                                <div>
                                  <span className="font-medium text-gray-700">留言：</span>
                                  <p className="text-gray-600 mt-1 whitespace-pre-wrap">{item.message}</p>
                                </div>
                              )}
                              {item.sourceUrl && (
                                <div>
                                  <span className="font-medium text-gray-700">来源页面：</span>
                                  <span className="text-gray-600">{item.sourceUrl}</span>
                                </div>
                              )}
                              {item.locale && (
                                <div>
                                  <span className="font-medium text-gray-700">语言：</span>
                                  <span className="text-gray-600">{item.locale}</span>
                                </div>
                              )}
                              {item.ipAddress && (
                                <div>
                                  <span className="font-medium text-gray-700">IP：</span>
                                  <span className="text-gray-600">{item.ipAddress}</span>
                                </div>
                              )}
                              {item.metadata && Object.keys(item.metadata).length > 0 && (
                                <div>
                                  <span className="font-medium text-gray-700">元数据：</span>
                                  <pre className="mt-1 p-2 bg-white border border-gray-200 rounded text-xs text-gray-600 overflow-x-auto">
                                    {JSON.stringify(item.metadata, null, 2)}
                                  </pre>
                                </div>
                              )}
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
                {data.items.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-6 py-8 text-center text-sm text-gray-500">
                      暂无表单提交记录
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="mt-4 flex items-center justify-between">
              <p className="text-sm text-gray-500">
                共 {data.total} 条，第 {page}/{totalPages} 页
              </p>
              <div className="flex gap-2">
                <button
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  上一页
                </button>
                <button
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
    </div>
  );
}
