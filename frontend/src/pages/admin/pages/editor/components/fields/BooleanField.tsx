import { AdminCheckbox } from "@/components/admin/ui";
import type { FieldProps } from "./types";

export default function BooleanField({ schema, value, onChange }: FieldProps) {
  return (
    <AdminCheckbox
      checked={!!value}
      onChange={(e) => onChange(e.target.checked)}
      label={schema.label}
    />
  );
}
