import { EditorContent, type Editor } from "@tiptap/react";
import EditorBubbleMenu from "@/components/admin/editor/EditorBubbleMenu";
import TableBubbleMenu from "@/components/admin/editor/TableBubbleMenu";
import EditorFloatingMenu from "@/components/admin/editor/EditorFloatingMenu";
import ArticleTypographyRoot from "@/components/blog/ArticleTypographyRoot";

/**
 * TipTap canvas + bubble/floating menus.
 * Lazy-loaded with the richtext surface so Markdown-only sessions skip this chunk.
 */
export function RichTextSurface({
  editor,
  showMenus,
  metadata,
}: {
  editor: Editor;
  showMenus: boolean;
  metadata: Record<string, unknown>;
}) {
  return (
    <div className="h-full overflow-y-auto">
      {showMenus && (
        <>
          <EditorBubbleMenu editor={editor} />
          <TableBubbleMenu editor={editor} />
          <EditorFloatingMenu editor={editor} />
        </>
      )}
      <ArticleTypographyRoot
        mode="editor"
        articleMetadata={metadata}
        className="h-full article-editor-content"
      >
        <EditorContent editor={editor} className="h-full" />
      </ArticleTypographyRoot>
    </div>
  );
}

export default RichTextSurface;
