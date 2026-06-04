import { useState } from "react";
import { adminReplyComment, type AdminComment, adminCommentAuthorName } from "./api";

interface AdminCommentReplyPanelProps {
  comment: AdminComment;
  onClose: () => void;
  onSent: () => void;
}

export default function AdminCommentReplyPanel({
  comment,
  onClose,
  onSent,
}: AdminCommentReplyPanelProps) {
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      await adminReplyComment({ content: content.trim(), parentId: comment.id });
      onSent();
      onClose();
    } catch {
      setError("回复失败，请稍后重试");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div
        className="bg-white rounded-lg shadow-xl w-full max-w-lg p-6"
        role="dialog"
        aria-labelledby="reply-dialog-title"
      >
        <h3 id="reply-dialog-title" className="text-lg font-semibold text-gray-900 mb-2">
          以作者身份回复
        </h3>
        <p className="text-sm text-gray-500 mb-4">
          回复 <strong>{adminCommentAuthorName(comment)}</strong> 的评论，将直接显示在前台。
        </p>
        <form onSubmit={handleSubmit} className="space-y-4">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm min-h-[120px]"
            placeholder="写下回复…"
            required
            autoFocus
          />
          {error && <p className="text-sm text-red-600">{error}</p>}
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-md"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
            >
              {submitting ? "发送中…" : "发送回复"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
