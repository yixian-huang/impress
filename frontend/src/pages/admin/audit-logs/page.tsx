import { useState } from "react";
import { getAuditLogs, type AuditEvent, type AuditLogFilters } from "@/api/auditLogs";
import {
  AdminBadge,
  AdminButton,
  AdminErrorBanner,
  AdminInput,
  AdminLoading,
  AdminPageHeader,
  AdminPagination,
  AdminSelect,
  AdminTable,
  AdminTableBody,
  AdminTableHead,
  AdminTd,
  AdminTh,
  AdminToolbar,
  AdminTr,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useAdminQuery } from "@/lib/adminQuery";
import { adminQueryKeys } from "@/lib/adminQueryKeys";

const ACTION_OPTIONS = [
  { value: "", label: "全部操作" },
  { value: "content.create", label: "创建内容" },
  { value: "content.update", label: "更新内容" },
  { value: "content.save_draft", label: "保存草稿" },
  { value: "content.publish", label: "发布内容" },
  { value: "content.unpublish", label: "下线内容" },
  { value: "content.rollback", label: "回滚内容" },
  { value: "content.delete", label: "删除内容" },
  { value: "content.validate", label: "验证内容" },
  { value: "auth.login", label: "登录" },
  { value: "permissions.assign", label: "分配权限" },
  { value: "permissions.unassign", label: "取消权限" },
  { value: "migration.import", label: "数据迁移" },
  { value: "migration.retry", label: "重试迁移" },
  { value: "media.upload", label: "上传媒体" },
  { value: "media.delete", label: "删除媒体" },
  { value: "backup.create", label: "创建备份" },
  { value: "backup.restore", label: "恢复备份" },
];

const PAGE_SIZE = 20;

function parseDetails(details: string): Record<string, unknown> | null {
  try {
    return JSON.parse(details);
  } catch {
    return null;
  }
}

function formatDetailsSummary(details: string): string {
  const parsed = parseDetails(details);
  if (!parsed) return details || "-";

  const parts: string[] = [];
  if (parsed.pageKey) parts.push(`页面: ${parsed.pageKey}`);
  if (parsed.version) parts.push(`版本: ${parsed.version}`);
  if (parsed.filename) parts.push(`文件: ${parsed.filename}`);
  if (parsed.reason) parts.push(`原因: ${parsed.reason}`);
  if (parsed.error) parts.push(`错误: ${parsed.error}`);
  if (parsed.status) parts.push(`状态码: ${parsed.status}`);
  if (parsed.request_id) parts.push(`请求: ${parsed.request_id}`);

  return parts.length > 0 ? parts.join(", ") : JSON.stringify(parsed).slice(0, 80);
}

export default function AdminAuditLogsPage() {
  useDocumentTitle("审计日志");
  const [page, setPage] = useState(1);
  const [filters, setFilters] = useState<AuditLogFilters>({});

  const { data, error, loading, isFetching, refetch } = useAdminQuery(
    [...adminQueryKeys.auditLogs, page, PAGE_SIZE, filters.action ?? "", filters.actor ?? "", filters.from ?? "", filters.to ?? ""],
    () => getAuditLogs(page, PAGE_SIZE, filters),
    { staleTime: 15_000 },
  );

  const fetchData = () => refetch({ force: true });
  const totalPages = data ? Math.ceil(data.total / PAGE_SIZE) : 0;

  const handleFilterChange = (key: keyof AuditLogFilters, value: string) => {
    setPage(1);
    setFilters((prev) => {
      const next = { ...prev };
      if (value) {
        next[key] = value;
      } else {
        delete next[key];
      }
      return next;
    });
  };

  return (
    <div>
      <AdminPageHeader
        title="审计日志"
        description="追踪后台关键操作记录"
        actions={
          <AdminButton size="sm" onClick={fetchData} disabled={isFetching}>
            {isFetching ? "加载中…" : "刷新"}
          </AdminButton>
        }
      />

      <AdminToolbar className="mb-6">
        <AdminSelect
          value={filters.action || ""}
          onChange={(e) => handleFilterChange("action", e.target.value)}
          aria-label="操作类型"
        >
          {ACTION_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </AdminSelect>
        <AdminInput
          type="text"
          placeholder="操作人"
          value={filters.actor || ""}
          onChange={(e) => handleFilterChange("actor", e.target.value)}
          className="max-w-[10rem]"
        />
        <AdminInput
          type="date"
          value={filters.from || ""}
          onChange={(e) => handleFilterChange("from", e.target.value)}
          className="max-w-[11rem]"
          aria-label="开始日期"
        />
        <AdminInput
          type="date"
          value={filters.to || ""}
          onChange={(e) => handleFilterChange("to", e.target.value)}
          className="max-w-[11rem]"
          aria-label="结束日期"
        />
      </AdminToolbar>

      {error && (
        <AdminErrorBanner message={error.message || "获取审计日志失败，请稍后重试"} />
      )}

      {loading && !data ? (
        <AdminLoading />
      ) : data ? (
        <>
          <AdminTable>
            <AdminTableHead>
              <tr>
                <AdminTh>时间</AdminTh>
                <AdminTh>操作</AdminTh>
                <AdminTh>操作人</AdminTh>
                <AdminTh>资源</AdminTh>
                <AdminTh>结果</AdminTh>
                <AdminTh>详情</AdminTh>
              </tr>
            </AdminTableHead>
            <AdminTableBody>
              {data.items.map((event: AuditEvent) => (
                <AdminTr key={event.id}>
                  <AdminTd className="whitespace-nowrap">
                    {new Date(event.createdAt).toLocaleString("zh-CN")}
                  </AdminTd>
                  <AdminTd className="font-medium text-slate-900">{event.action}</AdminTd>
                  <AdminTd>{event.actor}</AdminTd>
                  <AdminTd>{event.resource}</AdminTd>
                  <AdminTd>
                    <AdminBadge tone={event.result === "success" ? "success" : "danger"}>
                      {event.result}
                    </AdminBadge>
                  </AdminTd>
                  <AdminTd className="max-w-xs truncate text-slate-500" title={event.details}>
                    {formatDetailsSummary(event.details)}
                  </AdminTd>
                </AdminTr>
              ))}
              {data.items.length === 0 && (
                <tr>
                  <AdminTd colSpan={6} className="py-8 text-center text-slate-500">
                    暂无审计日志
                  </AdminTd>
                </tr>
              )}
            </AdminTableBody>
          </AdminTable>

          <AdminPagination
            page={page}
            totalPages={totalPages}
            total={data.total}
            onPageChange={setPage}
          />
        </>
      ) : null}
    </div>
  );
}
