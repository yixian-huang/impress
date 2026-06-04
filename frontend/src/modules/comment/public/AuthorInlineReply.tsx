import { useState } from "react";
import { useTranslation } from "react-i18next";
import type { CommentContentType } from "../api";
import { adminAuthorComment } from "../admin/api";
import { cr } from "../commentReadingStyles";

interface AuthorInlineReplyProps {
  contentType: CommentContentType;
  contentId: number;
  parentId?: number;
  replyToName?: string;
  isReading: boolean;
  onCancel: () => void;
  onSent: () => void;
}

export default function AuthorInlineReply({
  contentType,
  contentId,
  parentId,
  replyToName,
  isReading,
  onCancel,
  onSent,
}: AuthorInlineReplyProps) {
  const { t } = useTranslation("common");
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      await adminAuthorComment({
        content: content.trim(),
        contentType,
        contentId,
        parentId,
      });
      setContent("");
      onSent();
      onCancel();
    } catch {
      setError(t("comments.authorReplyFailed"));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className={
        isReading
          ? cr.authorForm
          : "mb-6 p-4 rounded-card bg-blue-50/80 border border-blue-100"
      }
    >
      <p className={isReading ? cr.authorFormLabel : "text-sm font-medium text-on-surface-muted mb-3"}>
        {replyToName
          ? t("comments.authorReplyingTo", { name: replyToName })
          : t("comments.authorReplyNew")}
      </p>
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        className={
          isReading
            ? `${cr.field} ${cr.textarea}`
            : "w-full border border-border rounded-button px-3 py-2 text-sm min-h-[5rem]"
        }
        placeholder={t("comments.authorReplyPlaceholder")}
        required
        autoFocus
      />
      {error && <p className="text-xs text-red-600">{error}</p>}
      <div className="flex items-center gap-4">
        <button type="submit" disabled={submitting} className={cr.submitPrimary}>
          {submitting ? t("comments.submitting") : t("comments.authorReplySubmit")}
        </button>
        <button type="button" onClick={onCancel} className={cr.action}>
          {t("comments.cancelReply")}
        </button>
      </div>
    </form>
  );
}
