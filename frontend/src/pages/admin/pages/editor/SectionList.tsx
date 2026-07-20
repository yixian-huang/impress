import { useSectionRegistry } from "@/plugins/hooks";
import type { SectionData } from "@/theme/types";

export default function SectionListItem({
  section,
  index,
  total,
  isSelected,
  isComposable,
  onSelect,
  onMoveUp,
  onMoveDown,
  onDelete,
  dragHandlers,
}: {
  section: SectionData;
  index: number;
  total: number;
  isSelected: boolean;
  isComposable: boolean;
  onSelect: () => void;
  onMoveUp: () => void;
  onMoveDown: () => void;
  onDelete: () => void;
  dragHandlers: {
    onDragStart: (e: React.DragEvent) => void;
    onDragOver: (e: React.DragEvent) => void;
    onDrop: (e: React.DragEvent) => void;
    onDragEnd: () => void;
  };
}) {
  const { metas: sectionMetas } = useSectionRegistry();
  const meta = sectionMetas.find((m) => m.type === section.type);
  const label = meta?.labelZh || section.type;
  const locked = !!section.locked;
  const draggable = isComposable && !locked;

  return (
    <div
      draggable={draggable}
      onDragStart={draggable ? dragHandlers.onDragStart : undefined}
      onDragOver={draggable ? dragHandlers.onDragOver : undefined}
      onDrop={draggable ? dragHandlers.onDrop : undefined}
      onDragEnd={draggable ? dragHandlers.onDragEnd : undefined}
      onClick={onSelect}
      className={`flex cursor-pointer select-none items-center gap-2 rounded-xl border px-3 py-2 transition-colors ${
        isSelected
          ? "border-blue-500 bg-blue-50 shadow-sm shadow-blue-600/10"
          : "border-slate-200 bg-white hover:border-slate-300"
      }`}
    >
      {locked ? (
        <span className="text-xs text-slate-400" title="模板锁定">
          &#128274;
        </span>
      ) : draggable ? (
        <span className="cursor-grab text-xs text-slate-400" title="拖拽排序">
          &#x2630;
        </span>
      ) : (
        <span className="text-xs text-slate-300">&#x2630;</span>
      )}

      <span className="flex-1 truncate text-sm text-slate-800">
        <span className="mr-1 text-slate-400">{index + 1}.</span>
        {label}
      </span>

      {isComposable && !locked && (
        <>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onMoveUp();
            }}
            disabled={index === 0}
            className="px-1 text-xs text-slate-400 hover:text-slate-700 disabled:opacity-30"
            title="上移"
          >
            &#9650;
          </button>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onMoveDown();
            }}
            disabled={index === total - 1}
            className="px-1 text-xs text-slate-400 hover:text-slate-700 disabled:opacity-30"
            title="下移"
          >
            &#9660;
          </button>
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
            className="px-1 text-sm text-slate-400 hover:text-red-600"
            title="删除"
          >
            &times;
          </button>
        </>
      )}
    </div>
  );
}
