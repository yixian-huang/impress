import { lazy, Suspense } from "react";
import type { Editor } from "@tiptap/react";

const LangEditorMountInner = lazy(() => import("./LangEditorMountInner"));

type Props = {
  enabled: boolean;
  html: string;
  editable: boolean;
  onDirty: () => void;
  onEditor: (editor: Editor | null) => void;
  onFlushBody?: (html: string) => void;
};

/**
 * Lazy gate for TipTap. Unmounted when `enabled` is false (e.g. pure Markdown mode)
 * so the TipTap+extensions chunk is not paid until richtext is needed.
 */
export function LangEditorMount({
  enabled,
  html,
  editable,
  onDirty,
  onEditor,
  onFlushBody,
}: Props) {
  if (!enabled) return null;
  return (
    <Suspense fallback={null}>
      <LangEditorMountInner
        html={html}
        editable={editable}
        onDirty={onDirty}
        onEditor={onEditor}
        onFlushBody={onFlushBody}
      />
    </Suspense>
  );
}
