import { useCallback, useEffect, useState } from "react";
import {
  compareArticleVersions,
  getArticleVersion,
  listArticleVersions,
  type ArticleVersionDetail,
  type ArticleVersionListItem,
} from "@/api/articles";
import { VersionHistoryList } from "./components/VersionHistoryList";
import { VersionCompareView } from "./components/VersionCompareView";

/** Snapshot shape used for compare/restore (subset of article fields). */
export type ArticleDraftSnapshot = {
  zhTitle?: string;
  enTitle?: string;
  slug?: string;
  status?: string;
  zhBody?: string;
  enBody?: string;
  coverImage?: string;
  zhSeoTitle?: string;
  enSeoTitle?: string;
  zhMetaDescription?: string;
  enMetaDescription?: string;
  ogImage?: string;
  author?: string;
  [key: string]: unknown;
};

function syntheticCurrentDetail(
  articleId: number,
  draft: ArticleDraftSnapshot,
): ArticleVersionDetail {
  return {
    id: 0,
    articleId,
    version: 0,
    snapshot: { ...draft },
    action: "current",
    summary: "当前编辑器内容",
    createdBy: 0,
    createdAt: new Date().toISOString(),
  };
}

export function ArticleVersionHistoryPanel({
  articleId,
  onClose,
  currentDraft,
  onRestore,
  canRestore = true,
}: {
  articleId: number;
  onClose: () => void;
  currentDraft?: ArticleDraftSnapshot | null;
  onRestore?: (snapshot: ArticleDraftSnapshot) => void;
  canRestore?: boolean;
}) {
  const [versions, setVersions] = useState<ArticleVersionListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [leftVer, setLeftVer] = useState<number | null>(null);
  const [rightVer, setRightVer] = useState<number | "current" | null>(null);
  const [comparing, setComparing] = useState(false);
  const [restoring, setRestoring] = useState(false);
  const [compareResult, setCompareResult] = useState<{
    left: ArticleVersionDetail;
    right: ArticleVersionDetail;
  } | null>(null);
  const [view, setView] = useState<"list" | "compare">("list");

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await listArticleVersions(articleId, 1, 50);
      const items = data.items || [];
      setVersions(items);
      if (items.length >= 1) {
        setLeftVer(items[0].version);
        setRightVer(
          currentDraft
            ? "current"
            : items.length >= 2
              ? items[1].version
              : items[0].version,
        );
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载版本失败");
    } finally {
      setLoading(false);
    }
  }, [articleId, currentDraft]);

  useEffect(() => {
    void load();
  }, [load]);

  const runCompare = async () => {
    if (leftVer == null || rightVer == null) return;
    setComparing(true);
    setError(null);
    try {
      if (rightVer === "current") {
        if (!currentDraft) {
          setError("无法获取当前编辑内容");
          return;
        }
        const left = await getArticleVersion(articleId, leftVer);
        setCompareResult({
          left,
          right: syntheticCurrentDetail(articleId, currentDraft),
        });
      } else {
        setCompareResult(await compareArticleVersions(articleId, leftVer, rightVer));
      }
      setView("compare");
    } catch (err) {
      setError(err instanceof Error ? err.message : "比对失败");
    } finally {
      setComparing(false);
    }
  };

  const handleRestore = async (version: number) => {
    if (!onRestore || !canRestore) return;
    if (
      !window.confirm(
        `确定将编辑器恢复到版本 v${version}？\n当前未保存的修改将被覆盖（恢复后仍需手动保存才会写回服务器）。`,
      )
    ) {
      return;
    }
    setRestoring(true);
    setError(null);
    try {
      const detail = await getArticleVersion(articleId, version);
      onRestore(detail.snapshot as ArticleDraftSnapshot);
    } catch (err) {
      setError(err instanceof Error ? err.message : "恢复失败");
    } finally {
      setRestoring(false);
    }
  };

  const handleQuickCompare = async (version: number) => {
    setLeftVer(version);
    setRightVer(currentDraft ? "current" : version);
    setComparing(true);
    setError(null);
    try {
      if (currentDraft) {
        const left = await getArticleVersion(articleId, version);
        setCompareResult({
          left,
          right: syntheticCurrentDetail(articleId, currentDraft),
        });
      } else {
        const older = versions.find((x) => x.version < version);
        setCompareResult(
          await compareArticleVersions(articleId, older?.version ?? version, version),
        );
      }
      setView("compare");
    } catch (err) {
      setError(err instanceof Error ? err.message : "比对失败");
    } finally {
      setComparing(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-black/30" onClick={onClose}>
      <div
        className={`bg-white h-full shadow-xl flex flex-col ${
          view === "compare" ? "w-full max-w-4xl" : "w-full max-w-md"
        }`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-200 flex-shrink-0">
          <div className="flex items-center gap-2">
            {view === "compare" && (
              <button
                type="button"
                onClick={() => setView("list")}
                className="text-sm text-slate-500 hover:text-slate-800 mr-1"
              >
                ← 返回
              </button>
            )}
            <h3 className="text-base font-semibold text-slate-900">
              {view === "compare" ? "版本比对" : "历史版本"}
            </h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 text-xl leading-none"
          >
            &times;
          </button>
        </div>

        {error && (
          <div className="px-4 py-2 bg-red-50 text-red-700 text-sm border-b border-red-100 flex-shrink-0">
            {error}
          </div>
        )}

        {view === "list" ? (
          <VersionHistoryList
            versions={versions}
            loading={loading}
            currentDraft={currentDraft}
            leftVer={leftVer}
            rightVer={rightVer}
            comparing={comparing}
            restoring={restoring}
            canRestore={canRestore && !!onRestore}
            onLeftChange={setLeftVer}
            onRightChange={setRightVer}
            onCompare={() => void runCompare()}
            onRestore={onRestore ? (v) => void handleRestore(v) : undefined}
            onQuickCompare={(v) => void handleQuickCompare(v)}
          />
        ) : (
          compareResult && (
            <VersionCompareView
              left={compareResult.left}
              right={compareResult.right}
              restoring={restoring}
              canRestore={canRestore && !!onRestore}
              onRestoreLeft={onRestore ? (v) => void handleRestore(v) : undefined}
            />
          )
        )}
      </div>
    </div>
  );
}
