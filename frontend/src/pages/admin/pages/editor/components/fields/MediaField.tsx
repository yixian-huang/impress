import { useState } from "react";
import type { FieldProps } from "./types";
import ImagePickerModal from "@/components/admin/ImagePickerModal";
import type { MediaItem } from "@/api/media";

export default function MediaField({ schema, value, onChange }: FieldProps) {
  const [showPicker, setShowPicker] = useState(false);
  const url = (value as string) ?? "";

  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">
        {schema.label}
      </label>
      <div className="flex gap-2">
        <input
          type="text"
          className="w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm"
          value={url}
          placeholder={schema.placeholder ?? "图片URL"}
          onChange={(e) => onChange(e.target.value)}
        />
        <button
          type="button"
          className="shrink-0 px-3 py-1.5 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
          onClick={() => setShowPicker(true)}
        >
          选择
        </button>
      </div>
      {url && (
        <img
          src={url}
          alt="preview"
          className="mt-2 w-20 h-20 object-cover rounded border border-gray-200"
        />
      )}
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
    </div>
  );
}
