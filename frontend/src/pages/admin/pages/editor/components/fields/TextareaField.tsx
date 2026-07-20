import { AdminField, AdminTextarea } from "@/components/admin/ui";
import type { FieldProps } from "./types";

export default function TextareaField({ schema, value, onChange }: FieldProps) {
  return (
    <AdminField label={schema.label}>
      <AdminTextarea
        rows={3}
        className="rounded-lg py-1.5"
        value={(value as string) ?? ""}
        placeholder={schema.placeholder}
        onChange={(e) => onChange(e.target.value)}
      />
    </AdminField>
  );
}
