import { AdminField, AdminInput } from "@/components/admin/ui";
import type { FieldProps } from "./types";

export default function TextField({ schema, value, onChange }: FieldProps) {
  return (
    <AdminField label={schema.label}>
      <AdminInput
        type="text"
        className="rounded-lg py-1.5"
        value={(value as string) ?? ""}
        placeholder={schema.placeholder}
        onChange={(e) => onChange(e.target.value)}
      />
    </AdminField>
  );
}
