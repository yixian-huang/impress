/** Small character counter for SEO fields. */

export function CharCountMeter({
  length,
  max,
  min,
}: {
  length: number;
  max: number;
  min?: number;
}) {
  let tone = "text-slate-400";
  if (length > max) tone = "text-red-600";
  else if (min != null && length > 0 && length < min) tone = "text-amber-600";
  else if (length > 0) tone = "text-emerald-600";

  return (
    <span className={`text-[10px] tabular-nums font-medium ${tone}`} title={`建议 ${min != null ? `${min}–` : "≤"}${max} 字`}>
      {length}/{max}
    </span>
  );
}
