import { useCallback, useEffect, useMemo, useState } from "react";
import {
  compareArticleVersions,
  listArticleVersions,
  type ArticleVersionDetail,
  type ArticleVersionListItem,
} from "@/api/articles";
import { diffLines, htmlToPlainText, type DiffLine } from "@/lib/textDiff";

const ACTION_LABEL: Record<string, string> = {
  create: "创建",
  save: "保存",
  publish: "发布",
  update: "更新",
};

function formatTime(iso: string) {
  try {
    return new Date(iso).toLocaleString("zh-CN");
  } catch {
    return iso;
  }
}

function DiffView({ lines }: { lines: DiffLine[] }) {
  if (lines.length === 0) {
    return <div className="text-sm text-gray-400 py-8 text-center">无差异</div>;
  }
  const changed = lines.filter((l) => l.op !== "equal").length;
  return (
    <div className="font-mono text-xs leading-5 border border-gray-200 rounded-lg overflow-hidden">
      <div className="px-3 py-1.5 bg-gray-50 border-b border-gray-200 text-gray-500 flex justify-between">
        <span>行级差异</span>
        <span>{changed === 0 ? "完全相同" : `${changed} 处变更行`}</span>
      </div>
      <div className="max-h-[50vh] overflow-auto">
        {lines.map((line, idx) => {
          const bg =
            line.op === "add"
              ? "bg-green-50 text-green-900"
              : line.op === "remove"
                ? "bg-red-50 text-red-900"
                : "bg-white text-gray-700";
          const prefix = line.op === "add" ? "+" : line.op === "remove" ? "−" : " ";
          return (
            <div key={idx} className={`flex ${bg} border-b border-gray-50 last:border-0`}>
              <span className="w-10 flex-shrink-0 text-right pr-2 text-gray-400 select-none tabular-nums">
                {line.leftLine ?? ""}
              </span>
              <span className="w-10 flex-shrink-0 text-right pr-2 text-gray-400 select-none tabular-nums">
                {line.rightLine ?? ""}
              </span>
              <span className="w-4 flex-shrink-0 text-center select-none opacity-70">{prefix}</span>
              <pre className="flex-1 whitespace-pre-wrap break-all py-0.5 pr-2 m-0 font-mono text-xs">
                {line.text || " "}
              </pre>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function FieldDiff({
  label,
  left,
  right,
  asHtml,
}: {
  label: string;
  left: string;
  right: string;
  asHtml?: boolean;
}) {
  const leftText = asHtml ? htmlToPlainText(left) : left;
  const rightText = asHtml ? htmlToPlainText(right) : right;
  const same = leftText === rightText;
  const lines = useMemo(
    () => (same ? [] : diffLines(leftText, rightText)),
    [leftText, rightText, same],
  );

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <h4 className="text-sm font-semibold text-gray-800">{label}</h4>
        {same ? (
          <span className="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 text-gray-500">相同</span>
        ) : (
          <span className="text-[10px] px-1.5 py-0.5 rounded bg-amber-100 text-amber-800">有变更</span>
        )}
      </div>
      {same ? (
        <div className="text-xs text-gray-500 bg-gray-50 rounded p-2 max-h-24 overflow-auto whitespace-pre-wrap">
          {leftText || <span className="italic text-gray-400">（空）</span>}
        </div>
      ) : (
        <DiffView lines={lines} />
      )}
    </div>
  );
}

function snapStr(snap: Record<string, unknown> | undefined, key: string): string {
  if (!snap) return "";
  const v = snap[key];
  return typeof v === "string" ? v : v == null ? "" : String(v);
}

export function ArticleVersionHistoryPanel({
  articleId,
  onClose,
}: {
  articleId: number;
  onClose: () => void;
}) {
  const [versions, setVersions] = useState<ArticleVersionListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [leftVer, setLeftVer] = useState<number | null>(null);
  const [rightVer, setRightVer] = useState<number | null>(null);
  const [comparing, setComparing] = useState(false);
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
      if (items.length >= 2) {
        setRightVer(items[0].version);
        setLeftVer(items[1].version);
      } else if (items.length === 1) {
        setRightVer(items[0].version);
        setLeftVer(items[0].version);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载版本失败");
    } finally {
      setLoading(false);
    }
  }, [articleId]);

  useEffect(() => {
    void load();
  }, [load]);

  const runCompare = async () => {
    if (leftVer == null || rightVer == null) return;
    setComparing(true);
    setError(null);
    try {
      const result = await compareArticleVersions(articleId, leftVer, rightVer);
      setCompareResult(result);
      setView("compare");
    } catch (err) {
      setError(err instanceof Error ? err.message : "比对失败");
    } finally {
      setComparing(false);
    }
  };

  const leftSnap = compareResult?.left.snapshot;
  const rightSnap = compareResult?.right.snapshot;

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-black/30" onClick={onClose}>
      <div
        className={`bg-white h-full shadow-xl flex flex-col ${
          view === "compare" ? "w-full max-w-4xl" : "w-full max-w-md"
        }`}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200 flex-shrink-0">
          <div className="flex items-center gap-2">
            {view === "compare" && (
              <button
                type="button"
                onClick={() => setView("list")}
                className="text-sm text-gray-500 hover:text-gray-800 mr-1"
              >
                ← 返回
              </button>
            )}
            <h3 className="text-base font-semibold text-gray-900">
              {view === "compare" ? "版本比对" : "历史版本"}
            </h3>
          </div>
          <button type="button" onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl leading-none">
            &times;
          </button>
        </div>

        {error && (
          <div className="px-4 py-2 bg-red-50 text-red-700 text-sm border-b border-red-100 flex-shrink-0">
            {error}
          </div>
        )}

        {view === "list" ? (
          <>
            <div className="px-4 py-3 border-b border-gray-100 space-y-2 flex-shrink-0 bg-gray-50">
              <p className="text-xs text-gray-500">选择两个版本进行比对（左侧为旧版，右侧为新版）。</p>
              <div className="flex items-center gap-2">
                <select
                  value={leftVer ?? ""}
                  onChange={(e) => setLeftVer(Number(e.target.value))}
                  className="flex-1 text-sm border border-gray-300 rounded-lg px-2 py-1.5"
                  disabled={versions.length === 0}
                >
                  {versions.map((v) => (
                    <option key={`L-${v.version}`} value={v.version}>
                      v{v.version} · {ACTION_LABEL[v.action] || v.action} · {formatTime(v.createdAt)}
                    </option>
                  ))}
                </select>
                <span className="text-xs text-gray-400">vs</span>
                <select
                  value={rightVer ?? ""}
                  onChange={(e) => setRightVer(Number(e.target.value))}
                  className="flex-1 text-sm border border-gray-300 rounded-lg px-2 py-1.5"
                  disabled={versions.length === 0}
                >
                  {versions.map((v) => (
                    <option key={`R-${v.version}`} value={v.version}>
                      v{v.version} · {ACTION_LABEL[v.action] || v.action} · {formatTime(v.createdAt)}
                    </option>
                  ))}
                </select>
              </div>
              <button
                type="button"
                onClick={() => void runCompare()}
                disabled={comparing || leftVer == null || rightVer == null || versions.length === 0}
                className="w-full px-3 py-1.5 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
              >
                {comparing ? "比对中…" : "开始比对"}
              </button>
            </div>

            <div className="flex-1 overflow-y-auto p-4">
              {loading ? (
                <div className="text-center text-gray-500 py-8">加载中…</div>
              ) : versions.length === 0 ? (
                <div className="text-center text-gray-500 py-8 text-sm">
                  暂无版本记录
                  <p className="mt-1 text-xs text-gray-400">保存或发布文章后会自动生成版本快照。</p>
                </div>
              ) : (
                <div className="space-y-2">
                  {versions.map((v) => (
                    <div
                      key={v.id}
                      className="p-3 border border-gray-200 rounded-lg hover:border-blue-200 transition-colors"
                    >
                      <div className="flex items-center justify-between gap-2">
                        <div className="text-sm font-medium text-gray-800">
                          版本 {v.version}
                          <span className="ml-2 text-[10px] px-1.5 py-0.5 rounded bg-gray-100 text-gray-600">
                            {ACTION_LABEL[v.action] || v.action}
                          </span>
                        </div>
                        <button
                          type="button"
                          className="text-xs text-blue-600 hover:text-blue-800"
                          onClick={() => {
                            setRightVer(v.version);
                            const older = versions.find((x) => x.version < v.version);
                            setLeftVer(older?.version ?? v.version);
                            void (async () => {
                              setComparing(true);
                              try {
                                const result = await compareArticleVersions(
                                  articleId,
                                  older?.version ?? v.version,
                                  v.version,
                                );
                                setCompareResult(result);
                                setView("compare");
                              } catch (err) {
                                setError(err instanceof Error ? err.message : "比对失败");
                              } finally {
                                setComparing(false);
                              }
                            })();
                          }}
                        >
                          查看
                        </button>
                      </div>
                      <div className="text-xs text-gray-500 mt-1">{formatTime(v.createdAt)}</div>
                      {(v.zhTitle || v.summary) && (
                        <div className="text-xs text-gray-600 mt-1 truncate">
                          {v.zhTitle || v.summary}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        ) : (
          <div className="flex-1 overflow-y-auto p-4 space-y-6">
            {compareResult && (
              <>
                <div className="flex items-center gap-3 text-sm bg-gray-50 rounded-lg p-3 border border-gray-100">
                  <div className="flex-1">
                    <div className="text-xs text-gray-400">左（旧）</div>
                    <div className="font-medium">v{compareResult.left.version}</div>
                    <div className="text-xs text-gray-500">{formatTime(compareResult.left.createdAt)}</div>
                  </div>
                  <div className="text-gray-300">→</div>
                  <div className="flex-1 text-right">
                    <div className="text-xs text-gray-400">右（新）</div>
                    <div className="font-medium">v{compareResult.right.version}</div>
                    <div className="text-xs text-gray-500">{formatTime(compareResult.right.createdAt)}</div>
                  </div>
                </div>

                <FieldDiff
                  label="中文标题"
                  left={snapStr(leftSnap, "zhTitle")}
                  right={snapStr(rightSnap, "zhTitle")}
                />
                <FieldDiff
                  label="英文标题"
                  left={snapStr(leftSnap, "enTitle")}
                  right={snapStr(rightSnap, "enTitle")}
                />
                <FieldDiff
                  label="Slug"
                  left={snapStr(leftSnap, "slug")}
                  right={snapStr(rightSnap, "slug")}
                />
                <FieldDiff
                  label="状态"
                  left={snapStr(leftSnap, "status")}
                  right={snapStr(rightSnap, "status")}
                />
                <FieldDiff
                  label="中文正文"
                  left={snapStr(leftSnap, "zhBody")}
                  right={snapStr(rightSnap, "zhBody")}
                  asHtml
                />
                <FieldDiff
                  label="英文正文"
                  left={snapStr(leftSnap, "enBody")}
                  right={snapStr(rightSnap, "enBody")}
                  asHtml
                />
              </>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
