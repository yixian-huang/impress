import { getAnalyticsSummary, type AnalyticsSummary } from "@/api/analytics";
import {
  AdminButton,
  AdminErrorBanner,
  AdminLoading,
  AdminPageHeader,
  AdminStatCard,
  AdminTable,
  AdminTableBody,
  AdminTableHead,
  AdminTd,
  AdminTh,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useAdminQuery } from "@/lib/adminQuery";
import { adminQueryKeys } from "@/lib/adminQueryKeys";

const PAGE_KEY_LABELS: Record<string, string> = {
  home: "首页",
  about: "关于我们",
  advantages: "核心优势",
  "core-services": "核心服务",
  cases: "成功案例",
  experts: "专家团队",
  contact: "联系我们",
  global: "全局配置",
};

export default function AdminAnalyticsPage() {
  useDocumentTitle("访问统计");
  const { data, error, loading, isFetching, refetch } = useAdminQuery(
    adminQueryKeys.analytics,
    getAnalyticsSummary,
    { staleTime: 20_000 },
  );

  return (
    <div>
      <AdminPageHeader
        title="访问统计"
        description={`按页面汇总的访问趋势${isFetching && !loading ? " · 刷新中" : ""}`}
        actions={
          <AdminButton
            size="sm"
            onClick={() => refetch({ force: true })}
            disabled={isFetching}
          >
            {isFetching ? "加载中…" : "刷新"}
          </AdminButton>
        }
      />

      {error && (
        <AdminErrorBanner
          message={error.message || "获取统计数据失败，请稍后重试"}
        />
      )}

      {loading && !data ? (
        <AdminLoading />
      ) : data ? (
        <AnalyticsBody data={data} />
      ) : null}
    </div>
  );
}

function AnalyticsBody({ data }: { data: AnalyticsSummary }) {
  return (
    <>
      <div className="mb-8 grid grid-cols-1 gap-4 md:grid-cols-3">
        <AdminStatCard
          label="今日访问"
          value={data.totals.today.toLocaleString()}
          colorClass="bg-gradient-to-br from-blue-500 to-blue-700"
          icon={<span className="text-sm font-bold">今</span>}
        />
        <AdminStatCard
          label="近 7 天"
          value={data.totals.last7d.toLocaleString()}
          colorClass="bg-gradient-to-br from-emerald-500 to-emerald-700"
          icon={<span className="text-sm font-bold">7d</span>}
        />
        <AdminStatCard
          label="近 30 天"
          value={data.totals.last30d.toLocaleString()}
          colorClass="bg-gradient-to-br from-violet-500 to-violet-700"
          icon={<span className="text-sm font-bold">30</span>}
        />
      </div>

      <AdminTable>
        <AdminTableHead>
          <tr>
            <AdminTh>页面</AdminTh>
            <AdminTh className="text-right">今日</AdminTh>
            <AdminTh className="text-right">近 7 天</AdminTh>
            <AdminTh className="text-right">近 30 天</AdminTh>
          </tr>
        </AdminTableHead>
        <AdminTableBody>
          {data.pages.map((page) => (
            <tr key={page.pageKey} className="transition-colors hover:bg-slate-50/70">
              <AdminTd className="whitespace-nowrap font-medium text-slate-900">
                {PAGE_KEY_LABELS[page.pageKey] || page.pageKey}
                <span className="ml-2 text-xs text-slate-400">{page.pageKey}</span>
              </AdminTd>
              <AdminTd className="whitespace-nowrap text-right tabular-nums">
                {page.today.toLocaleString()}
              </AdminTd>
              <AdminTd className="whitespace-nowrap text-right tabular-nums">
                {page.last7d.toLocaleString()}
              </AdminTd>
              <AdminTd className="whitespace-nowrap text-right tabular-nums">
                {page.last30d.toLocaleString()}
              </AdminTd>
            </tr>
          ))}
          {data.pages.length === 0 && (
            <tr>
              <AdminTd colSpan={4} className="py-8 text-center text-slate-500">
                暂无访问数据
              </AdminTd>
            </tr>
          )}
          {data.pages.length > 0 && (
            <tr className="bg-slate-50 font-semibold">
              <AdminTd className="whitespace-nowrap text-slate-900">合计</AdminTd>
              <AdminTd className="whitespace-nowrap text-right tabular-nums text-slate-900">
                {data.totals.today.toLocaleString()}
              </AdminTd>
              <AdminTd className="whitespace-nowrap text-right tabular-nums text-slate-900">
                {data.totals.last7d.toLocaleString()}
              </AdminTd>
              <AdminTd className="whitespace-nowrap text-right tabular-nums text-slate-900">
                {data.totals.last30d.toLocaleString()}
              </AdminTd>
            </tr>
          )}
        </AdminTableBody>
      </AdminTable>
    </>
  );
}
