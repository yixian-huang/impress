import { Suspense, lazy, useMemo, useState, memo } from "react";
import type { Editor } from "@tiptap/react";
import { parseMarkdownOutline, type OutlineItem } from "./utils/outline";
import type { RichTextOutlineStats } from "./components/RichTextOutlineSource";

const LazyRichTextOutlineSource = lazy(() =>
  import("./components/RichTextOutlineSource").then((m) => ({
    default: m.RichTextOutlineSource,
  })),
);

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between gap-2">
      <span className="text-slate-400 flex-shrink-0">{label}</span>
      <span className="text-slate-700 font-medium text-right break-all">{value}</span>
    </div>
  );
}

function OutlineNav({
  headings,
  editorMode,
  onSelect,
}: {
  headings: OutlineItem[];
  editorMode: "richtext" | "markdown";
  onSelect: (item: OutlineItem) => void;
}) {
  if (headings.length === 0) {
    return (
      <p className="text-xs text-slate-400 italic">
        暂无标题
        <span className="block mt-1 text-[10px] not-italic">
          {editorMode === "markdown" ? "使用 # / ## / ### 添加标题" : "使用 H1–H3 添加标题"}
        </span>
      </p>
    );
  }
  return (
    <nav className="space-y-0.5" aria-label="文章大纲">
      {headings.map((h, i) => (
        <button
          key={`${h.level}-${h.line ?? h.pos ?? i}-${h.text}`}
          type="button"
          onClick={() => onSelect(h)}
          className="block w-full text-left text-xs py-1 px-1.5 rounded hover:bg-slate-200 text-slate-700 truncate transition-colors"
          style={{ paddingLeft: `${(Math.min(h.level, 4) - 1) * 10 + 6}px` }}
          title={h.text}
        >
          <span className="text-slate-400 mr-1">H{h.level}</span>
          {h.text || <span className="text-slate-300 italic">空标题</span>}
        </button>
      ))}
    </nav>
  );
}

function SidebarShell({
  compact,
  activeTab,
  setActiveTab,
  editorMode,
  headings,
  charCount,
  wordCount,
  article,
  onSelectHeading,
}: {
  compact?: boolean;
  activeTab: "outline" | "details";
  setActiveTab: (t: "outline" | "details") => void;
  editorMode: "richtext" | "markdown";
  headings: OutlineItem[];
  charCount: number;
  wordCount: number;
  article: { slug: string; author: string; createdAt: string | null; publishedAt: string | null } | null;
  onSelectHeading: (item: OutlineItem) => void;
}) {
  const formatDate = (d: string | null) => {
    if (!d) return "—";
    try {
      return new Date(d).toLocaleString("zh-CN");
    } catch {
      return d;
    }
  };
  const showDetails = !compact;

  return (
    <div className="w-56 flex-shrink-0 border-l border-slate-200 bg-slate-50 flex flex-col min-h-0">
      {showDetails ? (
        <div className="flex border-b border-slate-200 flex-shrink-0">
          <button
            type="button"
            onClick={() => setActiveTab("outline")}
            className={`flex-1 px-3 py-2 text-xs font-medium transition-colors ${
              activeTab === "outline"
                ? "text-blue-700 border-b-2 border-blue-600 bg-white"
                : "text-slate-500 hover:text-slate-700"
            }`}
          >
            大纲
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("details")}
            className={`flex-1 px-3 py-2 text-xs font-medium transition-colors ${
              activeTab === "details"
                ? "text-blue-700 border-b-2 border-blue-600 bg-white"
                : "text-slate-500 hover:text-slate-700"
            }`}
          >
            详情
          </button>
        </div>
      ) : (
        <div className="px-3 py-2 text-xs font-medium text-slate-600 border-b border-slate-200 bg-white">
          大纲
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-3">
        {activeTab === "outline" || compact ? (
          <OutlineNav headings={headings} editorMode={editorMode} onSelect={onSelectHeading} />
        ) : (
          <div className="space-y-3 text-xs">
            {editorMode === "richtext" && (
              <>
                <DetailRow label="字符数" value={charCount.toLocaleString()} />
                <DetailRow label="词数" value={wordCount.toLocaleString()} />
              </>
            )}
            {article && (
              <>
                <DetailRow label="创建时间" value={formatDate(article.createdAt)} />
                <DetailRow label="发布时间" value={formatDate(article.publishedAt)} />
                {article.author && <DetailRow label="作者" value={article.author} />}
                {article.slug && (
                  <div>
                    <div className="text-slate-400 mb-0.5">访问链接</div>
                    <a
                      href={`/blog/${article.slug}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:underline break-all"
                    >
                      /blog/{article.slug}
                    </a>
                  </div>
                )}
              </>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

const EditorSidebar = memo(function EditorSidebar({
  editorMode,
  editor,
  markdownSource,
  onJumpMarkdownLine,
  article,
  compact,
}: {
  editorMode: "richtext" | "markdown";
  editor: Editor | null;
  markdownSource?: string;
  onJumpMarkdownLine?: (line: number) => void;
  article: { slug: string; author: string; createdAt: string | null; publishedAt: string | null } | null;
  compact?: boolean;
}) {
  const [activeTab, setActiveTab] = useState<"outline" | "details">("outline");

  const mdHeadings = useMemo(
    () => (editorMode === "markdown" ? parseMarkdownOutline(markdownSource) : []),
    [editorMode, markdownSource],
  );

  const scrollToHeading = (item: OutlineItem) => {
    if (editorMode === "markdown" && item.line != null && onJumpMarkdownLine) {
      onJumpMarkdownLine(item.line);
      return;
    }
    if (!editor || item.pos == null) return;
    editor.chain().focus().setTextSelection(item.pos + 1).run();
    try {
      const dom = editor.view.domAtPos(item.pos + 1);
      const el = dom.node instanceof HTMLElement ? dom.node : dom.node.parentElement;
      el?.scrollIntoView({ behavior: "smooth", block: "center" });
    } catch {
      /* ignore */
    }
  };

  if (editorMode === "markdown") {
    return (
      <SidebarShell
        compact={compact}
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        editorMode={editorMode}
        headings={mdHeadings}
        charCount={0}
        wordCount={0}
        article={article}
        onSelectHeading={scrollToHeading}
      />
    );
  }

  // Richtext: outline stats come from a TipTap-only lazy child
  return (
    <Suspense
      fallback={
        <SidebarShell
          compact={compact}
          activeTab={activeTab}
          setActiveTab={setActiveTab}
          editorMode={editorMode}
          headings={[]}
          charCount={0}
          wordCount={0}
          article={article}
          onSelectHeading={scrollToHeading}
        />
      }
    >
      <LazyRichTextOutlineSource editor={editor}>
        {(stats: RichTextOutlineStats) => (
          <SidebarShell
            compact={compact}
            activeTab={activeTab}
            setActiveTab={setActiveTab}
            editorMode={editorMode}
            headings={stats.headings}
            charCount={stats.charCount}
            wordCount={stats.wordCount}
            article={article}
            onSelectHeading={scrollToHeading}
          />
        )}
      </LazyRichTextOutlineSource>
    </Suspense>
  );
});

export default EditorSidebar;
