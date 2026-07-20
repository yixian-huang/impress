import { useSectionRegistry } from "@/plugins/hooks";
import { AdminModal } from "@/components/admin/ui";

export default function SectionPicker({
  onSelect,
  onClose,
}: {
  onSelect: (type: string) => void;
  onClose: () => void;
}) {
  const { metas: sectionMetas } = useSectionRegistry();

  return (
    <AdminModal open title="添加区块" onClose={onClose} widthClass="max-w-lg">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        {sectionMetas.map((meta) => (
          <button
            key={meta.type}
            type="button"
            onClick={() => onSelect(meta.type)}
            className="flex flex-col items-start rounded-xl border border-slate-200 p-3 text-left transition-colors hover:border-blue-400 hover:bg-blue-50"
          >
            <span className="text-sm font-medium text-slate-900">{meta.labelZh}</span>
            <span className="text-xs text-slate-500">{meta.label}</span>
          </button>
        ))}
      </div>
    </AdminModal>
  );
}
