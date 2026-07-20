import { uploadMedia, type MediaItem } from "@/api/media";

export const DEFAULT_IMAGE_MAX_BYTES = 20 * 1024 * 1024;

export type MediaUploadEvent =
  | { type: "start"; id: string; name: string; size: number }
  | { type: "progress"; id: string; percent: number }
  | { type: "success"; id: string }
  | {
      type: "error";
      id: string;
      name: string;
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
  for (const listener of listeners) {
    try {
      listener(event);
    } catch {
      /* ignore subscriber errors */
    }
  }
}

function makeUploadId(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `up-${crypto.randomUUID()}`;
  }
  return `up-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 9)}`;
}

/** Pull a readable message from axios / Error / unknown. */
export function formatUploadError(err: unknown, fallback = "上传失败"): string {
  if (!err) return fallback;
  if (typeof err === "string" && err.trim()) return err;
  const ax = err as {
    response?: { data?: { error?: { message?: string }; message?: string } };
    message?: string;
  };
  const fromApi =
    ax?.response?.data?.error?.message
    || ax?.response?.data?.message;
  if (typeof fromApi === "string" && fromApi.trim()) return fromApi;
  if (err instanceof Error && err.message.trim()) return err.message;
  return fallback;
}

export type UploadMediaTrackedOptions = {
  filename?: string;
  /** Called after a successful upload (e.g. insert into editor) */
  onInserted?: (result: { url: string; filename: string }) => void;
  maxSize?: number;
};

/**
 * Upload a media file/blob with progress events.
 * Returns the full MediaItem so pickers can select it immediately.
 * On failure, emits `error` with retry that re-runs the same work (same tray id).
 */
export async function uploadMediaTracked(
  file: File | Blob,
  opts?: UploadMediaTrackedOptions,
): Promise<MediaItem> {
  const id = makeUploadId();
  const name =
    (opts?.filename || (file instanceof File ? file.name : "") || "image").trim() || "image";
  const maxSize = opts?.maxSize ?? DEFAULT_IMAGE_MAX_BYTES;
  const size = file.size ?? 0;

  if (size > maxSize) {
    const mb = (size / 1024 / 1024).toFixed(1);
    const maxMb = Math.round(maxSize / 1024 / 1024);
    const message = `「${name}」过大（${mb}MB），上限 ${maxMb}MB`;
    emitMediaUpload({ type: "error", id, name, message });
    throw new Error(message);
  }

  const run = async (): Promise<MediaItem> => {
    emitMediaUpload({ type: "start", id, name, size });
    try {
      const media = await uploadMedia(file, opts?.filename ?? name, {
        onProgress: (percent) => emitMediaUpload({ type: "progress", id, percent }),
      });
      opts?.onInserted?.({
        url: media.url,
        filename: media.filename || name,
      });
      emitMediaUpload({ type: "success", id });
      return media;
    } catch (err) {
      const detail = formatUploadError(err);
      const message = `「${name}」上传失败：${detail}`;
      emitMediaUpload({
        type: "error",
        id,
        name,
        message,
        retry: () => {
          void run().catch(() => {
            /* error re-emitted */
          });
        },
      });
      throw err;
    }
  };

  return run();
}

/** Fire-and-forget upload that inserts on success (paste / drop pipelines). */
export function uploadAndInsertImage(
  file: File,
  insert: (url: string, filename: string) => void,
  opts?: { maxSize?: number; filename?: string },
): void {
  void uploadMediaTracked(file, {
    maxSize: opts?.maxSize,
    filename: opts?.filename,
    onInserted: ({ url, filename }) => {
      if (!url) return;
      insert(url, filename);
    },
  }).catch(() => {
    /* error already emitted on the bus */
  });
}
