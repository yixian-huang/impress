import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { getComments, postComment, type Comment, type CommentContentType } from "../api";
import { useCommentIdentity } from "../useCommentIdentity";
import { useCanAuthorReply } from "../useCanAuthorReply";
import AuthorInlineReply from "./AuthorInlineReply";
import CommentComposer from "./CommentComposer";
import { useBranding } from "@/hooks/useBranding";
import { useLocaleMode } from "@/hooks/useLocaleMode";
import { formatArticleDate } from "@/utils/articleLocale";
import { cr } from "../commentReadingStyles";
import { useCommentReadingUi } from "../useCommentReadingUi";

export interface CommentSectionProps {
  contentType: CommentContentType;
  contentId: number;
}

interface ReplyTarget {
  id: number;
  authorName: string;
}

type AuthorReplyMode =
  | null
  | { kind: "article" }
  | { kind: "comment"; id: number; authorName: string };

function CommentItem({
  comment,
  authorLabel,
  onReply,
  onAuthorReply,
  canAuthorReply,
  depth = 0,
  isReading,
}: {
  comment: Comment;
  authorLabel: string;
  onReply: (target: ReplyTarget) => void;
  onAuthorReply?: (target: ReplyTarget) => void;
  canAuthorReply: boolean;
  depth?: number;
  isReading: boolean;
}) {
  const { t } = useTranslation("common");
  const { currentLocale } = useLocaleMode();
  const [collapsed, setCollapsed] = useState(false);
  const replyCount = comment.children?.length ?? 0;
  const isAuthor =
    comment.authorRole === "author" ||
    (authorLabel.length > 0 &&
      comment.authorName.trim().toLowerCase() === authorLabel.trim().toLowerCase());

  const itemClass = isReading
    ? depth > 0
      ? cr.itemReply
      : cr.item
    : depth > 0
      ? "border-l border-border pl-4 ml-1 py-3"
      : "py-5 border-b border-border last:border-b-0";

  return (
    <div className={itemClass}>
      <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1 mb-2">
        <span className={isReading ? cr.metaName : "text-sm font-medium text-on-surface"}>
          {comment.authorName}
        </span>
        {isAuthor && (
          <span className={isReading ? cr.metaBadge : "text-xs text-on-surface-muted"}>
            {t("comments.authorBadge")}
          </span>
        )}
        <time
          className={isReading ? cr.metaDate : "text-xs tabular-nums text-on-surface-muted font-sans"}
          dateTime={comment.createdAt}
        >
          {formatArticleDate(comment.createdAt, currentLocale)}
        </time>
        {comment.pinned && (
          <span className={isReading ? cr.metaBadge : "text-xs font-sans text-on-surface-muted"}>
            {t("comments.pinned")}
          </span>
        )}
      </div>
      <p className={isReading ? cr.body : "text-sm text-on-surface leading-relaxed whitespace-pre-wrap"}>
        {comment.content}
      </p>
      <div className="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1">
        <button
          type="button"
          onClick={() => onReply({ id: comment.id, authorName: comment.authorName })}
          className={isReading ? cr.action : "text-xs font-sans text-on-surface-muted hover:text-primary transition-colors"}
        >
          {t("comments.reply")}
        </button>
        {canAuthorReply && onAuthorReply && (
          <button
            type="button"
            onClick={() => onAuthorReply({ id: comment.id, authorName: comment.authorName })}
            className={isReading ? cr.actionPrimary : "text-xs font-sans text-primary hover:text-accent transition-colors"}
          >
            {t("comments.authorReplyAction")}
          </button>
        )}
      </div>
      {replyCount > 0 && (
        <div className="mt-3">
          <button
            type="button"
            onClick={() => setCollapsed(!collapsed)}
            className={isReading ? cr.action : "text-xs font-sans text-on-surface-muted hover:text-primary transition-colors mb-2"}
          >
            {collapsed ? t("comments.showReplies", { count: replyCount }) : t("comments.hideReplies")}
          </button>
          {!collapsed && (
            <div className={isReading ? "space-y-0" : "space-y-0"}>
              {comment.children!.map((child) => (
                <CommentItem
                  key={child.id}
                  comment={child}
                  authorLabel={authorLabel}
                  onReply={onReply}
                  onAuthorReply={onAuthorReply}
                  canAuthorReply={canAuthorReply}
                  depth={depth + 1}
                  isReading={isReading}
                />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default function CommentSection({ contentType, contentId }: CommentSectionProps) {
  const { t } = useTranslation("common");
  const isReading = useCommentReadingUi(contentType);
  const canAuthorReply = useCanAuthorReply();
  const branding = useBranding();
  const { identity, persist } = useCommentIdentity();
  const authorLabel = branding.author.name?.trim() || branding.siteName;

  const [comments, setComments] = useState<Comment[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [replyTarget, setReplyTarget] = useState<ReplyTarget | null>(null);
  const [authorReplyMode, setAuthorReplyMode] = useState<AuthorReplyMode>(null);
  const [form, setForm] = useState({ content: "", authorName: "", authorEmail: "" });
  const [submitting, setSubmitting] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);

  useEffect(() => {
    setForm((prev) => ({
      ...prev,
      authorName: prev.authorName || identity.authorName,
      authorEmail: prev.authorEmail || identity.authorEmail,
    }));
  }, [identity.authorName, identity.authorEmail]);

  const loadComments = useCallback(async () => {
    const resp = await getComments(contentType, contentId, page);
    setComments(resp.comments ?? []);
    setTotal(resp.total);
  }, [contentType, contentId, page]);

  useEffect(() => {
    loadComments();
  }, [loadComments]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.content.trim() || !form.authorName.trim()) return;
    setSubmitting(true);
    setNotice(null);
    try {
      await postComment({
        content: form.content,
        authorName: form.authorName,
        authorEmail: form.authorEmail,
        contentType,
        contentId,
        parentId: replyTarget?.id,
      });
      persist({ authorName: form.authorName, authorEmail: form.authorEmail });
      setForm((f) => ({ ...f, content: "" }));
      setReplyTarget(null);
      setNotice(t("comments.submittedPending"));
      await loadComments();
    } finally {
      setSubmitting(false);
    }
  };

  const sectionTitle = (
    <h2 className={isReading ? cr.title : "text-lg font-semibold text-on-surface"}>
      {t("comments.title")}
      <span className={isReading ? cr.titleCount : " ml-1 font-normal text-on-surface-muted"}>
        ({total})
      </span>
    </h2>
  );

  const authorReplyParentId =
    authorReplyMode?.kind === "comment" ? authorReplyMode.id : undefined;
  const authorReplyToName =
    authorReplyMode?.kind === "comment" ? authorReplyMode.authorName : undefined;

  return (
    <section
      className={
        isReading ? cr.section : "mt-12 border-t border-border pt-8 font-sans"
      }
      aria-labelledby="comments-heading"
    >
      <div id="comments-heading" className="mb-6 flex flex-wrap items-baseline justify-between gap-3">
        {sectionTitle}
        {canAuthorReply && !authorReplyMode && (
          <button
            type="button"
            onClick={() => setAuthorReplyMode({ kind: "article" })}
            className={isReading ? cr.actionPrimary : "text-xs font-sans text-primary hover:text-accent transition-colors"}
          >
            {t("comments.authorReplyNew")}
          </button>
        )}
      </div>

      {canAuthorReply && authorReplyMode && (
        <AuthorInlineReply
          contentType={contentType}
          contentId={contentId}
          parentId={authorReplyParentId}
          replyToName={authorReplyToName}
          isReading={isReading}
          onCancel={() => setAuthorReplyMode(null)}
          onSent={loadComments}
        />
      )}

      <div className={isReading ? "mb-8" : ""}>
        {isReading && <p className={cr.sublabel}>{t("comments.guestFormLabel")}</p>}
        <CommentComposer
          values={form}
          onChange={setForm}
          onSubmit={handleSubmit}
          submitting={submitting}
          isReading={isReading}
          replyToAuthor={replyTarget?.authorName}
          onCancelReply={() => setReplyTarget(null)}
          notice={notice}
        />
      </div>

      {comments.length === 0 ? (
        <p className={isReading ? cr.hint : "text-sm text-on-surface-muted py-4"}>{t("comments.empty")}</p>
      ) : (
        <div className={isReading ? cr.list : "space-y-2"}>
          {comments.map((c) => (
            <CommentItem
              key={c.id}
              comment={c}
              authorLabel={authorLabel}
              onReply={setReplyTarget}
              onAuthorReply={(target) => {
                setAuthorReplyMode({ kind: "comment", id: target.id, authorName: target.authorName });
                setReplyTarget(null);
              }}
              canAuthorReply={canAuthorReply}
              isReading={isReading}
            />
          ))}
        </div>
      )}

      {total > 20 && (
        <div className="flex justify-center items-center gap-3 mt-8">
          {page > 1 && (
            <button
              type="button"
              onClick={() => setPage(page - 1)}
              className="px-3 py-1.5 text-sm border border-border rounded-button text-on-surface hover:bg-surface-alt transition-colors"
            >
              {t("pagination.prev")}
            </button>
          )}
          <span className="text-sm tabular-nums text-on-surface-muted">
            {t("comments.page", { page })}
          </span>
          {total > page * 20 && (
            <button
              type="button"
              onClick={() => setPage(page + 1)}
              className="px-3 py-1.5 text-sm border border-border rounded-button text-on-surface hover:bg-surface-alt transition-colors"
            >
              {t("pagination.next")}
            </button>
          )}
        </div>
      )}
    </section>
  );
}
