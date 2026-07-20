import { AdminField, AdminInput } from "@/components/admin/ui";
import type { FieldProps } from "./types";

export default function ColorField({ schema, value, onChange }: FieldProps) {
  const colorValue = (value as string) ?? "";

  return (
    <AdminField label={schema.label}>
      <div className="flex items-center gap-2">
        <AdminInput
          type="text"
          className="rounded-lg py-1.5 font-mono text-xs"
          value={colorValue}
          placeholder={schema.placeholder ?? "#000000"}
          onChange={(e) => onChange(e.target.value)}
        />
        <input
          type="color"
          value={colorValue && /^#/.test(colorValue) ? colorValue : "#000000"}
          onChange={(e) => onChange(e.target.value)}
          className="h-9 w-9 shrink-0 cursor-pointer rounded-lg border border-slate-200"
        />
      </div>
    </AdminField>
  );
}
