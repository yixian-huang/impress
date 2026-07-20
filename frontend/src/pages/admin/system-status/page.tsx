import { useCallback, useEffect, useState } from "react";
import type { ReactNode } from "react";
import { getSystemStatus, type SystemStatusResponse } from "@/api/systemStatus";
import {
  AdminButton,
  AdminErrorBanner,
  AdminLoading,
  AdminPageHeader,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

interface Metric {
  label: string;
  value: string | number;
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat("zh-CN").format(value);
}

function formatMB(value: number): string {
  return `${value.toFixed(2)} MB`;
}

function formatBytes(value: number): string {
  if (value >= 1024 * 1024 * 1024) return `${(value / 1024 / 1024 / 1024).toFixed(2)} GB`;
  if (value >= 1024 * 1024) return `${(value / 1024 / 1024).toFixed(2)} MB`;
  if (value >= 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${value} B`;
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (days > 0) return `${days} 天 ${hours} 小时`;
  if (hours > 0) return `${hours} 小时 ${minutes} 分钟`;
  return `${minutes} 分钟`;
}

function errorMessage(error: unknown): string {
  if (typeof error === "object" && error && "response" in error) {
    const response = (error as { response?: { data?: { error?: { message?: string } } } }).response;
    return response?.data?.error?.message || "系统状态加载失败";
  }
  return "系统状态加载失败";
}

function StatusBadge({ healthy, status }: { healthy: boolean; status: string }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium ${
        healthy ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"
      }`}
    >
      {healthy ? "正常" : status || "异常"}
    </span>
  );
}

function InfoCard({
  title,
  children,
  action,
}: {
  title: string;
  children: ReactNode;
  action?: ReactNode;
}) {
  return (
    <section className="bg-white rounded-lg border border-gray-200 p-5">
      <div className="flex items-center justify-between gap-3 mb-4">
        <h2 className="text-base font-semibold text-gray-900">{title}</h2>
        {action}
      </div>
      {children}
    </section>
  );
}

function MetricGrid({ metrics }: { metrics: Metric[] }) {
  return (
    <dl className="grid grid-cols-1 sm:grid-cols-2 gap-4">
      {metrics.map((metric) => (
        <div key={metric.label} className="rounded-lg bg-gray-50 px-4 py-3">
          <dt className="text-xs font-medium text-gray-500">{metric.label}</dt>
          <dd className="mt-1 text-sm font-semibold text-gray-900 break-words">{metric.value}</dd>
        </div>
      ))}
    </dl>
  );
}

function HealthMessage({ message }: { message?: string }) {
  if (!message) return null;
  return <p className="mt-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{message}</p>;
}

export default function AdminSystemStatusPage() {
  useDocumentTitle("系统状态");
  const [status, setStatus] = useState<SystemStatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState("");

  const fetchStatus = useCallback(async (manual = false) => {
    if (manual) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError("");
    try {
      setStatus(await getSystemStatus());
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  return (
    <div className="space-y-6">
      <AdminPageHeader
        title="系统状态"
        description="查看应用版本、数据库、存储、运行时和内容统计。"
        actions={
          <AdminButton
            size="sm"
            onClick={() => fetchStatus(true)}
            disabled={loading || refreshing}
          >
            {refreshing ? "刷新中…" : "刷新"}
          </AdminButton>
        }
      />

      {error && <AdminErrorBanner message={error} onDismiss={() => setError("")} />}

      {loading ? (
        <AdminLoading />
      ) : status ? (
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
          <InfoCard title="版本">
            <MetricGrid
              metrics={[
                { label: "应用版本", value: status.application.version || "dev" },
                { label: "Go 版本", value: status.runtime.goVersion },
              ]}
            />
          </InfoCard>

          <InfoCard title="数据库" action={<StatusBadge healthy={status.database.healthy} status={status.database.status} />}>
            <MetricGrid
              metrics={[
                { label: "类型", value: status.database.type },
                { label: "打开连接", value: formatNumber(status.database.openConnections) },
                { label: "使用中", value: formatNumber(status.database.inUse) },
                { label: "空闲", value: formatNumber(status.database.idle) },
                { label: "最大打开连接", value: formatNumber(status.database.maxOpenConnections) },
              ]}
            />
            <HealthMessage message={status.database.error} />
          </InfoCard>

          <InfoCard title="存储" action={<StatusBadge healthy={status.storage.healthy} status={status.storage.status} />}>
            <MetricGrid
              metrics={[
                { label: "类型", value: status.storage.type === "local" ? "本地存储" : status.storage.type },
                { label: "媒体文件", value: formatNumber(status.storage.mediaCount) },
                { label: "使用空间", value: formatBytes(status.storage.uploadDirBytes) },
                { label: "使用空间 (MB)", value: formatMB(status.storage.uploadDirSizeMB) },
              ]}
            />
            <HealthMessage message={status.storage.error} />
          </InfoCard>

          <InfoCard title="运行时">
            <MetricGrid
              metrics={[
                { label: "系统", value: `${status.runtime.os}/${status.runtime.arch}` },
                { label: "CPU", value: formatNumber(status.runtime.cpuCount) },
                { label: "协程", value: formatNumber(status.runtime.goroutines) },
                { label: "运行时间", value: formatUptime(status.runtime.uptime) },
                { label: "当前内存", value: formatMB(status.memory.allocMB) },
                { label: "系统内存", value: formatMB(status.memory.sysMB) },
                { label: "累计分配", value: formatMB(status.memory.totalAllocMB) },
                { label: "最近 GC 暂停", value: `${status.memory.gcPauseMs.toFixed(2)} ms` },
              ]}
            />
          </InfoCard>

          <InfoCard title="内容">
            <MetricGrid
              metrics={[
                { label: "统一页面", value: formatNumber(status.content.pages) },
                { label: "文章", value: formatNumber(status.content.articles) },
                { label: "媒体", value: formatNumber(status.content.media) },
                { label: "用户", value: formatNumber(status.content.users) },
              ]}
            />
          </InfoCard>
        </div>
      ) : (
        <div className="bg-white rounded-lg border border-gray-200 p-8 text-center text-gray-500">暂无状态数据</div>
      )}
    </div>
  );
}
