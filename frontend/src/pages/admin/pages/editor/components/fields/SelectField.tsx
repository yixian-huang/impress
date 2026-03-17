import type { FieldProps } from "./types";

export default function SelectField({ schema, value, onChange }: FieldProps) {
  const hasNumericOption = schema.options?.some(
    (opt) => typeof opt.value === "number"
  );

  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">
        {schema.label}
      </label>
      <select
        className="w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm"
        value={value != null ? String(value) : ""}
        onChange={(e) => {
          const v = e.target.value;
          if (v === "") {
            onChange(undefined);
          } else {
            onChange(hasNumericOption ? Number(v) : v);
          }
        }}
      >
        <option value="">请选择</option>
        {schema.options?.map((opt) => (
          <option key={String(opt.value)} value={String(opt.value)}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
}
