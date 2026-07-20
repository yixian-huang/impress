import { useState } from "react";
import type { FieldProps } from "./types";
import ImagePickerModal from "@/components/admin/ImagePickerModal";
import type { MediaItem } from "@/api/media";
import { AdminButton, AdminField, AdminInput } from "@/components/admin/ui";

function extractUrl(value: unknown): string {
  if (typeof value === "string") return value;
  if (value && typeof value === "object" && "url" in value) {
    return (value as { url: string }).url ?? "";
  }
  return "";
}

export default function MediaField({ schema, value, onChange }: FieldProps) {
  const [showPicker, setShowPicker] = useState(false);
  const url = extractUrl(value);

  return (
    <AdminField label={schema.label}>
      <div className="flex gap-2">
        <AdminInput
          type="text"
          className="rounded-lg py-1.5"
          value={url}
          placeholder={schema.placeholder ?? "图片URL"}
          onChange={(e) => onChange(e.target.value)}
        />
        <AdminButton
          type="button"
          size="sm"
          variant="secondary"
          className="shrink-0"
          onClick={() => setShowPicker(true)}
        >
          选择
        </AdminButton>
      </div>
      {url ? (
        <img
          src={url}
          alt="preview"
          className="mt-2 h-20 w-20 rounded-xl border border-slate-200 object-cover"
        />
      ) : null}
      {showPicker && (
        <ImagePickerModal
          open={showPicker}
          onClose={() => setShowPicker(false)}
          onSelect={(item: MediaItem) => {
            onChange(item.url);
            setShowPicker(false);
          }}
        />
      )}
    </AdminField>
  );
}
