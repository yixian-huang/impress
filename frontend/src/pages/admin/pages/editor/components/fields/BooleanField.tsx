import type { FieldProps } from "./types";

export default function BooleanField({ schema, value, onChange }: FieldProps) {
  return (
    <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
      <input
        type="checkbox"
        checked={!!value}
        onChange={(e) => onChange(e.target.checked)}
      />
      {schema.label}
    </label>
  );
}
