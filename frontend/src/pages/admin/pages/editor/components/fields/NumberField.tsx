import { AdminField, AdminInput } from "@/components/admin/ui";
import type { FieldProps } from "./types";

export default function NumberField({ schema, value, onChange }: FieldProps) {
  return (
    <AdminField label={schema.label}>
      <AdminInput
        type="number"
        className="rounded-lg py-1.5"
        value={value != null ? String(value) : ""}
        placeholder={schema.placeholder}
        onChange={(e) => {
          const v = e.target.value;
          onChange(v === "" ? undefined : Number(v));
        }}
      />
    </AdminField>
  );
}
