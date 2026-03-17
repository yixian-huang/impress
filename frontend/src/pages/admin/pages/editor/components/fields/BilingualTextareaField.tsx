import { useState } from "react";
import type { FieldProps } from "./types";

function normalize(value: unknown): { zh: string; en: string } {
  if (typeof value === "string") return { zh: value, en: "" };
  if (value == null) return { zh: "", en: "" };
  const obj = value as { zh?: string; en?: string };
  return { zh: obj.zh ?? "", en: obj.en ?? "" };
}

export default function BilingualTextareaField({ schema, value, onChange }: FieldProps) {
  const [tab, setTab] = useState<"zh" | "en">("zh");
  const normalized = normalize(value);

  return (
    <div>
      <div className="flex items-center justify-between mb-1">
        <label className="block text-xs font-medium text-gray-600">
          {schema.label}
        </label>
        <div className="flex gap-1 text-xs">
          <button
            type="button"
            className={tab === "zh" ? "font-bold text-blue-600" : "text-gray-400"}
            onClick={() => setTab("zh")}
          >
            zh
          </button>
          <span className="text-gray-300">|</span>
          <button
            type="button"
            className={tab === "en" ? "font-bold text-blue-600" : "text-gray-400"}
            onClick={() => setTab("en")}
          >
            en
          </button>
        </div>
      </div>
      <textarea
        rows={3}
        className="w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm"
        value={normalized[tab]}
        placeholder={schema.placeholder}
        onChange={(e) => onChange({ ...normalized, [tab]: e.target.value })}
      />
    </div>
  );
}
