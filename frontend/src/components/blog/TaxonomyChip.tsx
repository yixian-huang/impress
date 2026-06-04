interface TaxonomyChipProps {
  label: string;
  active?: boolean;
  onClick?: () => void;
  prefix?: string;
}

/** Flat text chip — no bordered box (reading blog archive). */
export default function TaxonomyChip({ label, active, onClick, prefix }: TaxonomyChipProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={[
        "inline-flex items-center gap-0.5 px-2 py-1 rounded-sm text-sm font-sans transition-colors",
        active
          ? "text-primary bg-primary/8"
          : "text-on-surface-muted hover:text-primary hover:bg-surface-alt/80",
      ].join(" ")}
      aria-current={active ? "true" : undefined}
    >
      {prefix && <span className="opacity-70">{prefix}</span>}
      {label}
    </button>
  );
}
