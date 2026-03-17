import type { FieldProps } from "./types";

export default function TextField({ schema, value, onChange }: FieldProps) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">
        {schema.label}
      </label>
      <input
        type="text"
        className="w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm"
        value={(value as string) ?? ""}
        placeholder={schema.placeholder}
        onChange={(e) => onChange(e.target.value)}
      />
    </div>
  );
}
