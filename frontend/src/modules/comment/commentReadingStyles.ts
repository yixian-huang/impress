/** Shared typography for blog reading-layout comment UI. */
export const cr = {
  section: "comment-reading mt-12 pt-10",
  title: "comment-reading__title text-on-surface",
  titleCount: "ml-1 font-normal text-on-surface-muted tabular-nums",
  metaName: "comment-reading__meta font-medium text-on-surface",
  metaDate: "comment-reading__meta text-on-surface-muted tabular-nums",
  metaBadge: "comment-reading__meta text-on-surface-muted",
  body: "comment-reading__body text-on-surface whitespace-pre-wrap",
  action: "comment-reading__action text-on-surface-muted hover:text-primary transition-colors",
  actionPrimary: "comment-reading__action text-primary hover:text-accent transition-colors",
  hint: "comment-reading__hint",
  notice: "comment-reading__hint",
  sublabel: "comment-reading__hint mb-3",
  field:
    "w-full bg-transparent border-0 border-b border-border/60 rounded-none px-0 py-2 text-sm text-on-surface placeholder:text-on-surface-muted/50 focus:outline-none focus:border-primary/50 transition-colors",
  textarea: "min-h-[5rem] resize-y leading-relaxed",
  submit:
    "text-sm text-on-surface hover:text-primary disabled:opacity-50 transition-colors",
  submitPrimary:
    "inline-flex items-center justify-center px-4 py-2 text-sm rounded-sm bg-primary/90 text-on-primary hover:bg-primary disabled:opacity-50 transition-colors",
  item: "py-5",
  itemReply: "pl-4 ml-0 py-4",
  authorForm: "mb-6 space-y-3",
  authorFormLabel: "text-sm text-on-surface-muted",
  list: "space-y-0",
} as const;
