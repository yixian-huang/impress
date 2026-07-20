import { useState } from "react";
import type { FieldProps } from "./types";
import { AdminField, AdminFilterChip, AdminInput } from "@/components/admin/ui";

function normalize(value: unknown): { zh: string; en: string } {
  if (typeof value === "string") return { zh: value, en: "" };
  if (value == null) return { zh: "", en: "" };
  const obj = value as { zh?: string; en?: string };
  return { zh: obj.zh ?? "", en: obj.en ?? "" };
}

export default function BilingualField({ schema, value, onChange }: FieldProps) {
  const [tab, setTab] = useState<"zh" | "en">("zh");
  const normalized = normalize(value);

  return (
    <AdminField
      label={
        <span className="flex w-full items-center justify-between gap-2">
          <span>{schema.label}</span>
          <span className="flex gap-1">
            <AdminFilterChip active={tab === "zh"} onClick={() => setTab("zh")} className="px-2 py-0.5">
              zh
            </AdminFilterChip>
            <AdminFilterChip active={tab === "en"} onClick={() => setTab("en")} className="px-2 py-0.5">
              en
            </AdminFilterChip>
          </span>
        </span>
      }
    >
      <AdminInput
        type="text"
        className="rounded-lg py-1.5"
        value={normalized[tab]}
        placeholder={schema.placeholder}
        onChange={(e) => onChange({ ...normalized, [tab]: e.target.value })}
      />
    </AdminField>
  );
}
