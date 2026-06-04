import { useTranslation } from "react-i18next";
import { cr } from "../commentReadingStyles";

export interface CommentFormValues {
  content: string;
  authorName: string;
  authorEmail: string;
}

interface CommentComposerProps {
  values: CommentFormValues;
  onChange: (values: CommentFormValues) => void;
  onSubmit: (e: React.FormEvent) => void;
  submitting: boolean;
  isReading: boolean;
  replyToAuthor?: string | null;
  onCancelReply?: () => void;
  notice?: string | null;
}

const boxedField =
  "w-full border border-border rounded-button bg-surface px-3 py-2 text-sm text-on-surface placeholder:text-on-surface-muted/70 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary/40";

export default function CommentComposer({
  values,
  onChange,
  onSubmit,
  submitting,
  isReading,
  replyToAuthor,
  onCancelReply,
  notice,
}: CommentComposerProps) {
  const { t } = useTranslation("common");
  const fieldClass = isReading ? cr.field : boxedField;

  return (
    <form
      onSubmit={onSubmit}
      className={
        isReading
          ? "mb-8 space-y-4"
          : "mb-8 space-y-3 border border-border rounded-card bg-surface-alt p-4"
      }
    >
      {replyToAuthor && onCancelReply && (
        <div className="flex flex-wrap items-center gap-2 text-sm text-on-surface-muted">
          <span>{t("comments.replyingToAuthor", { name: replyToAuthor })}</span>
          <button type="button" onClick={onCancelReply} className={cr.actionPrimary}>
            {t("comments.cancelReply")}
          </button>
        </div>
      )}

      <textarea
        placeholder={t("comments.write")}
        value={values.content}
        onChange={(e) => onChange({ ...values, content: e.target.value })}
        className={`${fieldClass} ${cr.textarea}`}
        rows={3}
        required
      />

      <div className="flex flex-col sm:flex-row sm:items-end gap-4 sm:gap-6">
        <label className="flex-1 min-w-0">
          <span className="sr-only">{t("comments.name")}</span>
          <input
            type="text"
            placeholder={t("comments.name")}
            value={values.authorName}
            onChange={(e) => onChange({ ...values, authorName: e.target.value })}
            className={fieldClass}
            required
            autoComplete="name"
          />
        </label>
        <label className="flex-1 min-w-0">
          <span className="sr-only">{t("comments.email")}</span>
          <input
            type="email"
            placeholder={t("comments.email")}
            value={values.authorEmail}
            onChange={(e) => onChange({ ...values, authorEmail: e.target.value })}
            className={fieldClass}
            autoComplete="email"
          />
        </label>
      </div>

      {isReading && <p className={cr.hint}>{t("comments.identityHint")}</p>}

      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        {notice && (
          <p className={cr.notice} role="status">
            {notice}
          </p>
        )}
        <button
          type="submit"
          disabled={submitting}
          className={
            isReading
              ? `${cr.submitPrimary} sm:ml-auto`
              : "font-sans text-sm px-4 py-2 rounded-button bg-primary text-on-primary hover:opacity-90 disabled:opacity-50 transition-opacity"
          }
        >
          {submitting ? t("comments.submitting") : t("comments.submit")}
        </button>
      </div>
    </form>
  );
}
