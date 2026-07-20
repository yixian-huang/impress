import type { LocalDraftOffer } from "../hooks/useLocalDraft";

export function LocalDraftBanner({
  offer,
  onRestore,
  onDismiss,
}: {
  offer: LocalDraftOffer;
  onRestore: () => void;
  onDismiss: () => void;
}) {
  return (
    <div
      role="status"
      className="flex flex-wrap items-center gap-2 px-4 py-2 bg-amber-50 border-b border-amber-200 text-sm text-amber-950"
    >
      <span className="flex-1 min-w-[12rem]">
        发现本地草稿（{offer.savedAtLabel}
        {offer.draft.zhTitle ? ` · ${offer.draft.zhTitle.slice(0, 40)}` : ""}
        ）— 可能因未保存关闭或网络中断留下。
      </span>
      <button
        type="button"
        onClick={onRestore}
        className="px-3 py-1 rounded-md bg-amber-600 text-white text-xs font-medium hover:bg-amber-700"
      >
        恢复草稿
      </button>
      <button
        type="button"
        onClick={onDismiss}
        className="px-3 py-1 rounded-md border border-amber-300 text-amber-900 text-xs hover:bg-amber-100"
      >
        丢弃
      </button>
    </div>
  );
}
