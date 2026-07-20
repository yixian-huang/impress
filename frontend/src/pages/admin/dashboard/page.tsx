import { Link } from "react-router-dom";
import {
  BarChart3,
  Eye,
  FileStack,
  FileText,
  Image as ImageIcon,
  Pencil,
  Plus,
  Settings2,
  Upload,
} from "lucide-react";
import { getAdminDashboardSummary } from "@/api/dashboard";
import {
  AdminCard,
  AdminPageHeader,
  AdminStatCard,
} from "@/components/admin/ui";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";
import { useAdminQuery } from "@/lib/adminQuery";
import { adminQueryKeys } from "@/lib/adminQueryKeys";
import { ADMIN_PAGES_PATH } from "@/router/adminAccess";

interface QuickAction {
  label: string;
  path: string;
  icon: React.ReactNode;
  /** Monochrome print weights — not rainbow chips. */
  tone: "ink" | "dark" | "mid" | "line" | "soft" | "ghost";
}

const toneClass: Record<QuickAction["tone"], string> = {
  ink: "bg-neutral-950 text-white border border-neutral-950 hover:bg-neutral-800",
  dark: "bg-neutral-800 text-white border border-neutral-800 hover:bg-neutral-700",
  mid: "bg-neutral-100 text-neutral-900 border border-neutral-200 hover:bg-neutral-200/80",
  line: "bg-white text-neutral-800 border border-neutral-300 hover:border-neutral-950 hover:bg-neutral-50",
  soft: "bg-neutral-50 text-neutral-700 border border-neutral-200 hover:bg-neutral-100",
  ghost: "bg-white text-neutral-600 border border-dashed border-neutral-300 hover:border-neutral-400 hover:text-neutral-900",
};

export default function AdminDashboardPage() {
  useDocumentTitle("仪表盘");
  const { data, loading } = useAdminQuery(
    adminQueryKeys.dashboardStats,
    getAdminDashboardSummary,
    { staleTime: 20_000 },
  );

  const stats = data ?? {
    todayVisits: 0,
    pagesCount: 0,
    articlesCount: 0,
    mediaCount: 0,
    errors: {} as Record<string, boolean>,
  };
  const errors = stats.errors ?? {};

  const quickActions: QuickAction[] = [
    {
      label: "新建文章",
      path: "/admin/articles/new",
      tone: "ink",
      icon: <Plus className="h-5 w-5" />,
    },
    {
      label: "上传媒体",
      path: "/admin/media",
      tone: "dark",
      icon: <Upload className="h-5 w-5" />,
    },
    {
      label: "编辑页面",
      path: ADMIN_PAGES_PATH,
      tone: "mid",
      icon: <Pencil className="h-5 w-5" />,
    },
    {
      label: "设置中心",
      path: "/admin/settings",
      tone: "line",
      icon: <Settings2 className="h-5 w-5" />,
    },
    {
      label: "访问统计",
      path: "/admin/analytics",
      tone: "soft",
      icon: <BarChart3 className="h-5 w-5" />,
    },
    {
      label: "站点配置",
      path: "/admin/site-config",
      tone: "ghost",
      icon: <FileStack className="h-5 w-5" />,
    },
  ];

  return (
    <div>
      <AdminPageHeader title="仪表盘" description="站点运行概览与常用操作" />

      <div className="mb-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <AdminStatCard
          label="今日访问"
          value={errors.todayVisits ? "—" : stats.todayVisits}
          colorClass="bg-neutral-950"
          loading={loading}
          icon={<Eye className="h-6 w-6" />}
        />
        <AdminStatCard
          label="内容页数"
          value={errors.pagesCount ? "—" : stats.pagesCount}
          colorClass="bg-neutral-800"
          loading={loading}
          icon={<FileStack className="h-6 w-6" />}
        />
        <AdminStatCard
          label="文章数"
          value={errors.articlesCount ? "—" : stats.articlesCount}
          colorClass="bg-neutral-600"
          loading={loading}
          icon={<FileText className="h-6 w-6" />}
        />
        <AdminStatCard
          label="媒体文件"
          value={errors.mediaCount ? "—" : stats.mediaCount}
          colorClass="bg-neutral-400"
          loading={loading}
          icon={<ImageIcon className="h-6 w-6 text-neutral-950" />}
        />
      </div>

      <AdminCard title="快捷操作" description="从这里进入最常用的管理入口">
        <div className="grid grid-cols-2 gap-3 lg:grid-cols-3">
          {quickActions.map((action) => (
            <Link
              key={action.label}
              to={action.path}
              className={`flex items-center gap-2.5 rounded-xl px-4 py-3.5 text-sm font-medium transition-all hover:-translate-y-0.5 ${toneClass[action.tone]}`}
            >
              {action.icon}
              {action.label}
            </Link>
          ))}
        </div>
      </AdminCard>
    </div>
  );
}
