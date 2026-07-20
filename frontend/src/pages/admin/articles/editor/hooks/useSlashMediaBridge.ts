import { useEffect } from "react";

type MediaState = {
  setShowImagePicker: (v: boolean) => void;
  setShowVideoPicker: (v: boolean) => void;
  setShowAudioPicker: (v: boolean) => void;
  setShowEmbedUrl: (v: boolean) => void;
  setShowGalleryPicker: (v: boolean) => void;
};

/** Bridge TipTap slash / replace media events to the active language modal state. */
export function useSlashMediaBridge(state: MediaState | undefined) {
  useEffect(() => {
    if (!state) return;
    const onSlashMedia = (e: Event) => {
      const type = (e as CustomEvent).detail?.type;
      if (type === "image") state.setShowImagePicker(true);
      else if (type === "video") state.setShowVideoPicker(true);
      else if (type === "audio") state.setShowAudioPicker(true);
      else if (type === "embed") state.setShowEmbedUrl(true);
      else if (type === "gallery") state.setShowGalleryPicker(true);
    };
    const onReplaceMedia = (e: Event) => {
      const type = (e as CustomEvent).detail?.type;
      if (type === "image") state.setShowImagePicker(true);
      else if (type === "video") state.setShowVideoPicker(true);
    };
    document.addEventListener("slash-command-media", onSlashMedia);
    document.addEventListener("editor-replace-media", onReplaceMedia);
    return () => {
      document.removeEventListener("slash-command-media", onSlashMedia);
      document.removeEventListener("editor-replace-media", onReplaceMedia);
    };
  }, [state]);
}
