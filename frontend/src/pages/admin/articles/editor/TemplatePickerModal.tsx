import { ARTICLE_TEMPLATES, type ArticleTemplate } from "./articleTemplates";

export default function TemplatePickerModal({
  open,
  onClose,
  onSelect,
}: {
  open: boolean;
  onClose: () => void;
  onSelect: (template: ArticleTemplate) => void;
}) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div
        className="bg-white rounded-xl shadow-xl w-full max-w-lg max-h-[80vh] overflow-hidden flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-200">
          <div>
            <h3 className="text-base font-semibold text-slate-900">选择文章模板</h3>
            <p className="text-xs text-slate-500 mt-0.5">将覆盖当前标题与正文（可撤销需自行版本恢复）</p>
          </div>
          <button type="button" onClick={onClose} className="text-slate-400 hover:text-slate-700 text-xl leading-none">
            &times;
          </button>
        </div>
        <div className="p-4 space-y-2 overflow-y-auto">
          {ARTICLE_TEMPLATES.map((tpl) => (
            <button
              key={tpl.id}
              type="button"
              onClick={() => onSelect(tpl)}
              className="w-full text-left p-3 rounded-lg border border-slate-200 hover:border-blue-300 hover:bg-blue-50/40 transition-colors"
            >
              <div className="text-sm font-semibold text-slate-900">{tpl.name}</div>
              <div className="text-xs text-slate-500 mt-0.5">{tpl.description}</div>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
