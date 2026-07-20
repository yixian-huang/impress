import { Extension } from "@tiptap/core";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { uploadAndInsertImage } from "@/lib/mediaUploadTracked";

export interface ImagePasteOptions {
  /** Max file size in bytes (default: 20MB) */
  maxSize?: number;
}

/**
 * Convert a base64 data URL to a File object.
 */
function dataUrlToFile(dataUrl: string, filename: string): File | null {
  try {
    const [header, base64] = dataUrl.split(",");
    if (!header || !base64) return null;
    const mimeMatch = header.match(/data:([^;]+)/);
    if (!mimeMatch) return null;
    const mime = mimeMatch[1];
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
      maxSize: 20 * 1024 * 1024,
    };
  },

  addProseMirrorPlugins() {
    const maxSize = this.options.maxSize || 20 * 1024 * 1024;
    const editor = this.editor;

    const doUpload = (file: File) => {
      uploadAndInsertImage(
        file,
        (url, filename) => {
          if (!url) return;
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

            const items = Array.from(clipboardData.items || []);
            const imageItems = items.filter((i) => i.type.startsWith("image/"));

            if (imageItems.length > 0) {
              event.preventDefault();
              for (const imageItem of imageItems) {
                const file = imageItem.getAsFile();
                if (file) doUpload(file);
              }
              return true;
            }

            const html = clipboardData.getData("text/html");
            if (html && /src=["']data:image\/[^"']{1000,}["']/i.test(html)) {
              event.preventDefault();
              const base64Regex = /src=["'](data:image\/[^"']+)["']/gi;
              const matches = [...html.matchAll(base64Regex)];

              for (const match of matches.slice(0, 5)) {
                const dataUrl = match[1];
                const ext = dataUrl.match(/data:image\/(\w+)/)?.[1] || "png";
                const file = dataUrlToFile(dataUrl, `pasted-image.${ext}`);
                if (file) doUpload(file);
              }

              const text = clipboardData.getData("text/plain");
              if (text && matches.length === 0) {
                editor.chain().focus().insertContent(text).run();
              }

              return true;
            }

            return false;
          },

          handleDrop(_view, event) {
            const files = Array.from(event.dataTransfer?.files || []);
            const imageFiles = files.filter((f) => f.type.startsWith("image/"));
            if (imageFiles.length === 0) return false;

            event.preventDefault();
            for (const imageFile of imageFiles) {
              doUpload(imageFile);
            }
            return true;
          },
        },
      }),
    ];
  },
});
