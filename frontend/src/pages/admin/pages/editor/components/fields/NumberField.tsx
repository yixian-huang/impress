import type { FieldProps } from "./types";

export default function NumberField({ schema, value, onChange }: FieldProps) {
  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">
        {schema.label}
      </label>
      <input
        type="number"
        className="w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm"
        value={value != null ? String(value) : ""}
        placeholder={schema.placeholder}
        onChange={(e) => {
          const v = e.target.value;
          onChange(v === "" ? undefined : Number(v));
        }}
      />
    </div>
  );
}
