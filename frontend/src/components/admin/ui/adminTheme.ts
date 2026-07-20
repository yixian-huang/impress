/**
 * Inkless admin — monochrome art-print / letterpress.
 *
 * Pure black–white–gray hierarchy (no warm yellow paper, no SaaS blue).
 * Contrast via ink density, border weight, and type — not hue.
 */
export const adminTheme = {
  /* ── Surfaces (cool print stock) ──────────────────────────── */
  pageBg: "bg-[#f4f4f5]",
  shellBg: "bg-[#f4f4f5]",
  surface: "bg-white",
  surfaceMuted: "bg-neutral-100/90",
  surfaceSubtle: "bg-neutral-100",

  /* ── Sidebar (black ink rail) ─────────────────────────────── */
  sidebarBg: "bg-[#0a0a0a]",
  sidebarBorder: "border-white/10",
  sidebarText: "text-neutral-300",
  sidebarTextMuted: "text-neutral-500",
  sidebarHover: "hover:bg-white/[0.06] hover:text-white",
  sidebarActive: "bg-white text-neutral-950 shadow-sm",
  sidebarActiveIcon: "text-neutral-950",
  sidebarIcon: "text-neutral-500 group-hover:text-neutral-200",
  sidebarSearch:
    "w-full rounded-lg border border-white/10 bg-white/[0.04] py-2 pl-8 pr-8 text-xs text-neutral-200 placeholder:text-neutral-500 outline-none transition focus:border-white/25 focus:bg-white/[0.07] focus:ring-1 focus:ring-white/15",

  /* ── Typography ───────────────────────────────────────────── */
  pageTitle:
    "text-[1.375rem] font-semibold tracking-[-0.02em] text-neutral-950 sm:text-[1.65rem]",
  pageDesc: "mt-1.5 text-sm leading-relaxed text-neutral-500",
  sectionTitle: "text-base font-semibold tracking-[-0.01em] text-neutral-950",
  sectionDesc: "mt-0.5 text-sm text-neutral-500",
  label: "text-sm font-medium text-neutral-700",
  muted: "text-neutral-500",
  caption: "text-xs tracking-wide text-neutral-500",
  body: "text-sm text-neutral-700",

  /* ── Borders & elevation ──────────────────────────────────── */
  border: "border-neutral-200",
  borderSubtle: "border-neutral-100",
  card:
    "bg-white rounded-xl border border-neutral-200/90 shadow-[0_1px_0_rgba(0,0,0,0.03),0_8px_24px_rgba(0,0,0,0.03)]",
  cardPad: "p-5 sm:p-6",
  cardHeader:
    "flex items-start justify-between gap-3 border-b border-neutral-100 px-5 py-4 sm:px-6",
  panel: "rounded-xl border border-neutral-200/90 bg-white",
  dropdown:
    "rounded-xl border border-neutral-200 bg-white py-1 shadow-[0_16px_40px_rgba(0,0,0,0.1)]",

  /* ── Interactive ──────────────────────────────────────────── */
  focusRing:
    "focus:outline-none focus-visible:ring-2 focus-visible:ring-neutral-950/20 focus-visible:ring-offset-2 focus-visible:ring-offset-[#f4f4f5]",
  focusRingInset:
    "focus:outline-none focus-visible:ring-2 focus-visible:ring-neutral-950/25 focus-visible:ring-offset-0",
  transition: "transition-all duration-150 ease-out",

  /* ── Form controls ────────────────────────────────────────── */
  input:
    "w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-950 placeholder:text-neutral-400 shadow-[inset_0_1px_2px_rgba(0,0,0,0.02)] transition hover:border-neutral-300 focus:border-neutral-950/35 focus:outline-none focus:ring-2 focus:ring-neutral-950/10 disabled:cursor-not-allowed disabled:bg-neutral-50 disabled:text-neutral-400",
  select:
    "rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-950 shadow-[inset_0_1px_2px_rgba(0,0,0,0.02)] transition hover:border-neutral-300 focus:border-neutral-950/35 focus:outline-none focus:ring-2 focus:ring-neutral-950/10 disabled:cursor-not-allowed disabled:bg-neutral-50",
  textarea:
    "w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-950 placeholder:text-neutral-400 shadow-[inset_0_1px_2px_rgba(0,0,0,0.02)] transition hover:border-neutral-300 focus:border-neutral-950/35 focus:outline-none focus:ring-2 focus:ring-neutral-950/10 disabled:cursor-not-allowed disabled:bg-neutral-50",
  checkbox:
    "h-4 w-4 rounded border-neutral-300 text-neutral-950 focus:ring-neutral-950/20",

  /* ── Toolbar / filter bars ────────────────────────────────── */
  toolbar:
    "flex flex-wrap items-center gap-2 rounded-xl border border-neutral-200/90 bg-white p-3 shadow-[0_1px_0_rgba(0,0,0,0.02)] sm:gap-3 sm:p-3.5",
  filterChip:
    "inline-flex items-center gap-1.5 rounded-full border border-neutral-200 bg-neutral-50 px-2.5 py-1 text-xs font-medium tracking-wide text-neutral-600",

  /* ── Table ────────────────────────────────────────────────── */
  tableHead:
    "bg-neutral-50 text-left text-[11px] font-semibold uppercase tracking-[0.08em] text-neutral-500",
  tableRow: "transition-colors hover:bg-neutral-50/90",
  tableCell: "px-4 py-3 text-neutral-700",
  tableCellHead: "px-4 py-3 whitespace-nowrap",

  /* ── Status (gray density, not rainbow) ───────────────────── */
  dangerSoft: "bg-neutral-100 text-neutral-900 border-neutral-300",
  successSoft: "bg-neutral-50 text-neutral-800 border-neutral-200",
  warningSoft: "bg-neutral-100 text-neutral-700 border-neutral-200",
  infoSoft: "bg-neutral-50 text-neutral-700 border-neutral-200",

  /* ── Primary = black ink ──────────────────────────────────── */
  primary: "bg-neutral-950 text-white hover:bg-neutral-800 active:bg-black",
  primarySoft:
    "bg-neutral-100 text-neutral-950 hover:bg-neutral-200/80 border border-neutral-200",
  link: "text-neutral-950 hover:text-neutral-700 font-medium underline-offset-4 hover:underline decoration-neutral-300",
} as const;

export type AdminTheme = typeof adminTheme;
