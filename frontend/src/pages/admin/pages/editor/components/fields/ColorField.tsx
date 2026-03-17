import type { FieldProps } from "./types";

export default function ColorField({ schema, value, onChange }: FieldProps) {
  const colorValue = (value as string) ?? "";

  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">
        {schema.label}
      </label>
      <div className="flex items-center gap-2">
        <input
          type="text"
          className="w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm"
          value={colorValue}
          placeholder={schema.placeholder ?? "#000000"}
          onChange={(e) => onChange(e.target.value)}
        />
        <div
          className="w-6 h-6 rounded border border-gray-300 shrink-0"
          style={{ backgroundColor: colorValue || "transparent" }}
        />
      </div>
    </div>
  );
}
