/** Lightweight placeholder while TipTap / Markdown chunks download. */
export function EditorChunkFallback({ label = "加载编辑器…" }: { label?: string }) {
  return (
    <div className="flex h-full min-h-[12rem] items-center justify-center text-sm text-slate-500">
      <div className="flex items-center gap-2">
        <span
          className="inline-block h-3.5 w-3.5 animate-spin rounded-full border-2 border-slate-300 border-t-blue-500"
          aria-hidden
        />
        <span>{label}</span>
      </div>
    </div>
  );
}
