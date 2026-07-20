import { useState, memo } from "react";
import { useEditorState } from "@tiptap/react";
import type { Editor } from "@tiptap/react";

interface HeadingItem {
  level: number;
  text: string;
  pos: number;
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between">
      <span className="text-slate-400">{label}</span>
      <span className="text-slate-700 font-medium">{value}</span>
    </div>
  );
}

const EditorSidebar = memo(function EditorSidebar({ editor, article }: {
  editor: Editor | null;
  article: { slug: string; author: string; createdAt: string | null; publishedAt: string | null } | null;
}) {
  const [activeTab, setActiveTab] = useState<"outline" | "details">("outline");

  // Extract headings & stats reactively from editor state
  const { headings, charCount, wordCount } = useEditorState({
    editor,
    selector: ({ editor: e }) => {
      if (!e) return { headings: [] as HeadingItem[], charCount: 0, wordCount: 0 };
      const h: HeadingItem[] = [];
      e.state.doc.descendants((node, pos) => {
        if (node.type.name === "heading") {
          h.push({ level: node.attrs.level as number, text: node.textContent, pos });
        }
      });
      const text = e.state.doc.textContent;
      const chars = text.length;
      // Count words: CJK chars each count as one word; latin words split by spaces
      const cjk = (text.match(/[\u4e00-\u9fff\u3400-\u4dbf\uf900-\ufaff]/g) || []).length;
      const latin = text.replace(/[\u4e00-\u9fff\u3400-\u4dbf\uf900-\ufaff]/g, " ").trim().split(/\s+/).filter(Boolean).length;
      return { headings: h, charCount: chars, wordCount: cjk + latin };
    },
    equalityFn: (a, b) => {
      if (a.charCount !== b.charCount || a.wordCount !== b.wordCount) return false;
      if (a.headings.length !== b.headings.length) return false;
      return a.headings.every((h, i) =>
        h.level === b.headings[i].level && h.text === b.headings[i].text && h.pos === b.headings[i].pos
      );
    },
  });

  const scrollToHeading = (pos: number) => {
    if (!editor) return;
    editor.chain().focus().setTextSelection(pos).run();
    // Scroll the DOM node into view
    try {
      const dom = editor.view.domAtPos(pos);
      const el = dom.node instanceof HTMLElement ? dom.node : dom.node.parentElement;
      el?.scrollIntoView({ behavior: "smooth", block: "center" });
    } catch { /* ignore */ }
  };

  const formatDate = (d: string | null) => {
    if (!d) return "—";
    try { return new Date(d).toLocaleString("zh-CN"); } catch { return d; }
  };

  return (
    <div className="w-60 flex-shrink-0 border-l border-slate-200 bg-slate-50 flex flex-col min-h-0">
      {/* Tab switcher */}
      <div className="flex border-b border-slate-200 flex-shrink-0">
        <button
          onClick={() => setActiveTab("outline")}
          className={`flex-1 px-3 py-2 text-xs font-medium transition-colors ${
            activeTab === "outline" ? "text-blue-700 border-b-2 border-blue-600 bg-white" : "text-slate-500 hover:text-slate-700"
          }`}
        >
          大纲
        </button>
        <button
          onClick={() => setActiveTab("details")}
          className={`flex-1 px-3 py-2 text-xs font-medium transition-colors ${
            activeTab === "details" ? "text-blue-700 border-b-2 border-blue-600 bg-white" : "text-slate-500 hover:text-slate-700"
          }`}
        >
          详情
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-3">
        {activeTab === "outline" ? (
          headings.length === 0 ? (
            <p className="text-xs text-slate-400 italic">暂无标题</p>
          ) : (
            <nav className="space-y-0.5">
              {headings.map((h, i) => (
                <button
                  key={i}
                  onClick={() => scrollToHeading(h.pos)}
                  className="block w-full text-left text-xs py-1 px-1.5 rounded hover:bg-slate-200 text-slate-700 truncate transition-colors"
                  style={{ paddingLeft: `${(h.level - 1) * 12 + 6}px` }}
                  title={h.text}
                >
                  <span className="text-slate-400 mr-1">H{h.level}</span>
                  {h.text || <span className="text-slate-300 italic">空标题</span>}
                </button>
              ))}
            </nav>
          )
        ) : (
          <div className="space-y-3 text-xs">
            <DetailRow label="字符数" value={charCount.toLocaleString()} />
            <DetailRow label="词数" value={wordCount.toLocaleString()} />
            {article && (
              <>
                <DetailRow label="创建时间" value={formatDate(article.createdAt)} />
                <DetailRow label="发布时间" value={formatDate(article.publishedAt)} />
                {article.author && <DetailRow label="作者" value={article.author} />}
                {article.slug && (
                  <div>
                    <div className="text-slate-400 mb-0.5">访问链接</div>
                    <a
                      href={`/articles/${article.slug}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:underline break-all"
                    >
                      /articles/{article.slug}
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
});

export default EditorSidebar;
