import { uploadMedia } from "@/api/media";

export type MediaUploadEvent =
  | { type: "start"; id: string; name: string; size: number }
  | { type: "progress"; id: string; percent: number }
  | { type: "success"; id: string }
  | {
      type: "error";
      id: string;
      message: string;
      /** Re-run the full upload + insert pipeline */
      retry?: () => void;
    }
  | { type: "dismiss"; id: string };

type Listener = (event: MediaUploadEvent) => void;

const listeners = new Set<Listener>();

export function subscribeMediaUpload(listener: Listener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function emitMediaUpload(event: MediaUploadEvent): void {
  listeners.forEach((l) => {
    try {
      l(event);
    } catch {
      /* ignore subscriber errors */
    }
  });
}

function makeUploadId(): string {
  return `up-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`;
}

/**
 * Upload a media file with progress events.
 * On failure, emits `error` with an optional retry that re-invokes the same work.
 */
export async function uploadMediaTracked(
  file: File,
  opts?: {
    filename?: string;
    /** Called after a successful upload (e.g. insert into editor) */
    onInserted?: (result: { url: string; filename: string }) => void;
    maxSize?: number;
  },
): Promise<{ url: string; filename: string }> {
  const id = makeUploadId();
  const name = opts?.filename || (file instanceof File ? file.name : "image") || "image";
  const maxSize = opts?.maxSize ?? 20 * 1024 * 1024;

  if (file.size > maxSize) {
    const mb = (file.size / 1024 / 1024).toFixed(1);
    const maxMb = (maxSize / 1024 / 1024).toFixed(0);
    emitMediaUpload({
      type: "error",
      id,
      message: `「${name}」过大（${mb}MB），上限 ${maxMb}MB`,
    });
    throw new Error(`Image too large: ${mb}MB`);
  }

  const run = async (): Promise<{ url: string; filename: string }> => {
    emitMediaUpload({ type: "start", id, name, size: file.size });
    try {
      const media = await uploadMedia(file, opts?.filename, {
        onProgress: (percent) => emitMediaUpload({ type: "progress", id, percent }),
      });
      const result = { url: media.url, filename: media.filename || name };
      opts?.onInserted?.(result);
      emitMediaUpload({ type: "success", id });
      return result;
    } catch (err) {
      const message =
        err instanceof Error && err.message
          ? `「${name}」上传失败：${err.message}`
          : `「${name}」上传失败`;
      emitMediaUpload({
        type: "error",
        id,
        message,
        retry: () => {
          void run();
        },
      });
      throw err;
    }
  };

  return run();
}

/** Convenience for TipTap ImagePaste / Markdown insert pipelines. */
export function uploadAndInsertImage(
  file: File,
  insert: (url: string, filename: string) => void,
  opts?: { maxSize?: number },
): void {
  void uploadMediaTracked(file, {
    maxSize: opts?.maxSize,
    onInserted: ({ url, filename }) => insert(url, filename),
  }).catch(() => {
    /* error already emitted */
  });
}
