/**
 * Shared design tokens for the admin console.
 * Prefer these over ad-hoc slate/gray classes so pages stay visually consistent.
 */
export const adminTheme = {
  /* ── Surfaces ─────────────────────────────────────────────── */
  pageBg: "bg-[#f4f6f9]",
  shellBg: "bg-[#f4f6f9]",
  surface: "bg-white",
  surfaceMuted: "bg-slate-50/80",
  surfaceSubtle: "bg-slate-50",

  /* ── Sidebar (dark rail) ──────────────────────────────────── */
  sidebarBg: "bg-[#0b1220]",
  sidebarBorder: "border-white/[0.06]",
  sidebarText: "text-slate-300",
  sidebarTextMuted: "text-slate-500",
  sidebarHover: "hover:bg-white/[0.06] hover:text-white",
  sidebarActive: "bg-blue-500/15 text-blue-200 shadow-[inset_3px_0_0_0] shadow-blue-400",
  sidebarActiveIcon: "text-blue-300",
  sidebarIcon: "text-slate-400 group-hover:text-slate-200",
  sidebarSearch:
    "w-full rounded-xl border border-white/10 bg-white/[0.06] py-2 pl-8 pr-8 text-xs text-slate-200 placeholder:text-slate-500 outline-none transition focus:border-blue-400/40 focus:bg-white/[0.08] focus:ring-2 focus:ring-blue-500/20",

  /* ── Typography ───────────────────────────────────────────── */
  pageTitle: "text-[1.375rem] font-semibold tracking-tight text-slate-900 sm:text-2xl",
  pageDesc: "mt-1 text-sm leading-relaxed text-slate-500",
  sectionTitle: "text-base font-semibold tracking-tight text-slate-900",
  sectionDesc: "mt-0.5 text-sm text-slate-500",
  label: "text-sm font-medium text-slate-700",
  muted: "text-slate-500",
  caption: "text-xs text-slate-500",
  body: "text-sm text-slate-700",

  /* ── Borders & elevation ──────────────────────────────────── */
  border: "border-slate-200/90",
  borderSubtle: "border-slate-100",
  card: "bg-white rounded-2xl border border-slate-200/80 shadow-[0_1px_2px_rgba(15,23,42,0.04),0_4px_16px_rgba(15,23,42,0.03)]",
  cardPad: "p-5 sm:p-6",
  cardHeader: "flex items-start justify-between gap-3 border-b border-slate-100/90 px-5 py-4 sm:px-6",
  panel: "rounded-2xl border border-slate-200/80 bg-white",
  dropdown:
    "rounded-xl border border-slate-200/90 bg-white py-1 shadow-[0_12px_40px_rgba(15,23,42,0.12)]",

  /* ── Interactive ──────────────────────────────────────────── */
  focusRing:
    "focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/35 focus-visible:ring-offset-2 focus-visible:ring-offset-[#f4f6f9]",
  focusRingInset:
    "focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/35 focus-visible:ring-offset-0",
  transition: "transition-all duration-150 ease-out",

  /* ── Form controls ────────────────────────────────────────── */
  input:
    "w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 shadow-sm transition hover:border-slate-300 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20 disabled:cursor-not-allowed disabled:bg-slate-50 disabled:text-slate-400",
  select:
    "rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm transition hover:border-slate-300 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20 disabled:cursor-not-allowed disabled:bg-slate-50",
  textarea:
    "w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 shadow-sm transition hover:border-slate-300 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/20 disabled:cursor-not-allowed disabled:bg-slate-50",
  checkbox:
    "h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500/30",

  /* ── Toolbar / filter bars ────────────────────────────────── */
  toolbar:
    "flex flex-wrap items-center gap-2 rounded-2xl border border-slate-200/80 bg-white p-3 shadow-[0_1px_2px_rgba(15,23,42,0.04)] sm:gap-3 sm:p-3.5",
  filterChip:
    "inline-flex items-center gap-1.5 rounded-full border border-slate-200 bg-slate-50 px-2.5 py-1 text-xs font-medium text-slate-600",

  /* ── Table ────────────────────────────────────────────────── */
  tableHead:
    "bg-slate-50/90 text-left text-[11px] font-semibold uppercase tracking-[0.06em] text-slate-500",
  tableRow: "transition-colors hover:bg-slate-50/70",
  tableCell: "px-4 py-3 text-slate-700",
  tableCellHead: "px-4 py-3 whitespace-nowrap",

  /* ── Status ───────────────────────────────────────────────── */
  dangerSoft: "bg-red-50 text-red-800 border-red-200",
  successSoft: "bg-emerald-50 text-emerald-800 border-emerald-200",
  warningSoft: "bg-amber-50 text-amber-800 border-amber-200",
  infoSoft: "bg-blue-50 text-blue-800 border-blue-200",

  /* ── Primary accent ───────────────────────────────────────── */
  primary: "bg-blue-600 text-white hover:bg-blue-700 active:bg-blue-800",
  primarySoft: "bg-blue-50 text-blue-700 hover:bg-blue-100 border border-blue-200/80",
  link: "text-blue-600 hover:text-blue-700 font-medium",
} as const;

export type AdminTheme = typeof adminTheme;
