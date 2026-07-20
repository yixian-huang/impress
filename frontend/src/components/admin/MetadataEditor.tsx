import { useState } from "react";
import { AdminButton, AdminInput } from "@/components/admin/ui";

interface MetadataEditorProps {
  value: Record<string, unknown>;
  onChange: (value: Record<string, unknown>) => void;
}

export default function MetadataEditor({ value, onChange }: MetadataEditorProps) {
  const [newKey, setNewKey] = useState("");
  const [newValue, setNewValue] = useState("");

  const entries = Object.entries(value || {});

  const handleAdd = () => {
    if (!newKey.trim()) return;
    onChange({ ...value, [newKey.trim()]: newValue });
    setNewKey("");
    setNewValue("");
  };

  const handleRemove = (key: string) => {
    const next = { ...value };
    delete next[key];
    onChange(next);
  };

  const handleValueChange = (key: string, val: string) => {
    onChange({ ...value, [key]: val });
  };

  return (
    <div className="space-y-2">
      {entries.map(([key, val]) => (
        <div key={key} className="flex items-center gap-2">
          <AdminInput
            type="text"
            value={key}
            readOnly
            className="w-1/3 rounded-lg bg-slate-50 py-1.5"
          />
          <AdminInput
            type="text"
            value={String(val ?? "")}
            onChange={(e) => handleValueChange(key, e.target.value)}
            className="flex-1 rounded-lg py-1.5"
          />
          <button
            type="button"
            onClick={() => handleRemove(key)}
            className="px-2 py-1.5 text-sm text-red-500 hover:text-red-700"
          >
            ×
          </button>
        </div>
      ))}
      <div className="flex items-center gap-2">
        <AdminInput
          type="text"
          value={newKey}
          onChange={(e) => setNewKey(e.target.value)}
          placeholder="Key"
          className="w-1/3 rounded-lg py-1.5"
          onKeyDown={(e) => e.key === "Enter" && handleAdd()}
        />
        <AdminInput
          type="text"
          value={newValue}
          onChange={(e) => setNewValue(e.target.value)}
          placeholder="Value"
          className="flex-1 rounded-lg py-1.5"
          onKeyDown={(e) => e.key === "Enter" && handleAdd()}
        />
        <AdminButton type="button" size="sm" variant="secondary" onClick={handleAdd}>
          添加
        </AdminButton>
      </div>
    </div>
  );
}
