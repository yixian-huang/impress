import { Extension } from "@tiptap/core";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import {
  DEFAULT_IMAGE_MAX_BYTES,
  uploadAndInsertImage,
} from "@/lib/mediaUploadTracked";

export interface ImagePasteOptions {
  /** Max file size in bytes (default: 20MB) */
  maxSize?: number;
}

function dataUrlToFile(dataUrl: string, filename: string): File | null {
  try {
    const [header, base64] = dataUrl.split(",");
    if (!header || !base64) return null;
    const mimeMatch = header.match(/data:image\/([\w+.-]+)/i);
    if (!mimeMatch) return null;
    const mime = `image/${mimeMatch[1]}`;
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return new File([bytes], filename, { type: mime });
  } catch {
    return null;
  }
}

/**
 * Paste / drop images into TipTap with progress tray + retry
 * (via mediaUploadTracked bus).
 */
export const ImagePaste = Extension.create<ImagePasteOptions>({
  name: "imagePaste",

  addOptions() {
    return {
      maxSize: DEFAULT_IMAGE_MAX_BYTES,
    };
  },

  addProseMirrorPlugins() {
    const maxSize = this.options.maxSize ?? DEFAULT_IMAGE_MAX_BYTES;
    const editor = this.editor;

    const doUpload = (file: File) => {
      uploadAndInsertImage(
        file,
        (url, filename) => {
          if (!url || editor.isDestroyed) return;
          editor.chain().focus().setImage({ src: url, alt: filename }).run();
        },
        { maxSize },
      );
    };

    return [
      new Plugin({
        key: new PluginKey("imagePaste"),
        props: {
          handlePaste(_view, event) {
            const clipboardData = event.clipboardData;
            if (!clipboardData) return false;

            const imageFiles = Array.from(clipboardData.items || [])
              .filter((i) => i.type.startsWith("image/"))
              .map((i) => i.getAsFile())
              .filter((f): f is File => !!f);

            if (imageFiles.length > 0) {
              event.preventDefault();
              imageFiles.forEach(doUpload);
              return true;
            }

            // Large base64 images in HTML freeze ProseMirror — upload instead
            const html = clipboardData.getData("text/html");
            if (html && /src=["']data:image\/[^"']{1000,}["']/i.test(html)) {
              event.preventDefault();
              const matches = [
                ...html.matchAll(/src=["'](data:image\/[^"']+)["']/gi),
              ];
              for (const match of matches.slice(0, 5)) {
                const dataUrl = match[1];
                const ext = dataUrl.match(/data:image\/([\w+.-]+)/i)?.[1] || "png";
                const file = dataUrlToFile(dataUrl, `pasted-image.${ext}`);
                if (file) doUpload(file);
              }
              return true;
            }

            return false;
          },

          handleDrop(_view, event) {
            const imageFiles = Array.from(event.dataTransfer?.files || []).filter((f) =>
              f.type.startsWith("image/"),
            );
            if (imageFiles.length === 0) return false;
            event.preventDefault();
            imageFiles.forEach(doUpload);
            return true;
          },
        },
      }),
    ];
  },
});
