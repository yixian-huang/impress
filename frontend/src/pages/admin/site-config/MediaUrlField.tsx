import { useState } from "react";
import ImagePickerModal from "@/components/admin/ImagePickerModal";

interface MediaUrlFieldProps {
  label: string;
  value: string;
  onChange: (url: string) => void;
  hint?: string;
  /** Square preview vs wide (OG). */
  preview?: "square" | "wide" | "logo";
}

/**
 * Image URL field with media-library picker + direct upload (via ImagePickerModal).
 * Still allows pasting an external CDN URL.
 */
export default function MediaUrlField({
  label,
  value,
  onChange,
  hint,
  preview = "square",
}: MediaUrlFieldProps) {
  const [pickerOpen, setPickerOpen] = useState(false);

  const previewClass =
    preview === "wide"
      ? "mt-2 max-h-28 max-w-full rounded border border-gray-200 object-cover"
      : preview === "logo"
        ? "mt-2 max-h-12 max-w-[12rem] rounded border border-gray-200 object-contain bg-gray-50 p-1"
        : "mt-2 h-16 w-16 rounded border border-gray-200 object-cover";

  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      {hint && <p className="text-xs text-gray-500 mb-1.5">{hint}</p>}
      <div className="flex flex-wrap items-center gap-2">
        <input
          type="text"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="图片 URL，或点击右侧选择/上传"
          className="flex-1 min-w-[12rem] border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-1 focus:ring-blue-500 focus:border-blue-500"
        />
        <button
          type="button"
          onClick={() => setPickerOpen(true)}
          className="shrink-0 px-3 py-2 text-sm border border-gray-300 rounded-lg hover:bg-gray-50"
        >
          图库 / 上传
        </button>
        {value && (
          <button
            type="button"
            onClick={() => onChange("")}
            className="shrink-0 px-2.5 py-2 text-sm text-red-600 border border-red-200 rounded-lg hover:bg-red-50"
          >
            清除
          </button>
        )}
      </div>
      {value && (
        <img
          src={value}
          alt=""
          className={previewClass}
          onError={(e) => {
            (e.target as HTMLImageElement).style.display = "none";
          }}
        />
      )}
      <ImagePickerModal
        open={pickerOpen}
        onClose={() => setPickerOpen(false)}
        onSelect={(item) => {
          onChange(item.url);
          setPickerOpen(false);
        }}
      />
    </div>
  );
}
