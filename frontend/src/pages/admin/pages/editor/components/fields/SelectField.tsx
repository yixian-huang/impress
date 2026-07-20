import { AdminField, AdminSelect } from "@/components/admin/ui";
import type { FieldProps } from "./types";

export default function SelectField({ schema, value, onChange }: FieldProps) {
  return (
    <AdminField label={schema.label}>
      <AdminSelect
        className="w-full rounded-lg py-1.5"
        value={value != null ? String(value) : ""}
        onChange={(e) => {
          const v = e.target.value;
          if (v === "") {
            onChange(undefined);
          } else {
            const selectedOpt = schema.options?.find((opt) => String(opt.value) === v);
            onChange(selectedOpt ? selectedOpt.value : v);
          }
        }}
      >
        <option value="">请选择</option>
        {schema.options?.map((opt) => (
          <option key={String(opt.value)} value={String(opt.value)}>
            {opt.label}
          </option>
        ))}
      </AdminSelect>
    </AdminField>
  );
}
