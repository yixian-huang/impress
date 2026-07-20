import { MediaUploadTray } from "@/components/admin/MediaUploadTray";
import { useMediaUploadTray } from "@/hooks/useMediaUploadTray";

/** Global host so picker/crop uploads share the same progress tray across admin. */
export function AdminMediaUploadHost() {
  const tray = useMediaUploadTray();
  return (
    <MediaUploadTray
      items={tray.items}
      onDismiss={tray.dismiss}
      onRetry={tray.retry}
      canRetry={tray.canRetry}
    />
  );
}
