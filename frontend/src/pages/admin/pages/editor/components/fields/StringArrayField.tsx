import type { FieldProps } from "./types";

export default function StringArrayField({ schema, value, onChange }: FieldProps) {
  const items: string[] = Array.isArray(value) ? (value as string[]) : [];

  const update = (index: number, newVal: string) => {
    const next = [...items];
    next[index] = newVal;
    onChange(next);
  };

  const remove = (index: number) => {
    const next = items.filter((_, i) => i !== index);
    onChange(next);
  };

  const moveUp = (index: number) => {
    if (index <= 0) return;
    const next = [...items];
    [next[index - 1], next[index]] = [next[index], next[index - 1]];
    onChange(next);
  };

  const moveDown = (index: number) => {
    if (index >= items.length - 1) return;
    const next = [...items];
    [next[index], next[index + 1]] = [next[index + 1], next[index]];
    onChange(next);
  };

  const add = () => {
    onChange([...items, ""]);
  };

  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">
        {schema.label}
      </label>
      <div className="space-y-1">
        {items.map((item, index) => (
          <div key={index} className="flex gap-1 items-center">
            <input
              type="text"
              className="w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm"
              value={item}
              onChange={(e) => update(index, e.target.value)}
            />
            <button
              type="button"
              title="上移"
              className="px-1 text-gray-400 hover:text-gray-600 text-sm"
              onClick={() => moveUp(index)}
            >
              ▲
            </button>
            <button
              type="button"
              title="下移"
              className="px-1 text-gray-400 hover:text-gray-600 text-sm"
              onClick={() => moveDown(index)}
            >
              ▼
            </button>
            <button
              type="button"
              title="删除"
              className="px-1 text-red-400 hover:text-red-600 text-sm"
              onClick={() => remove(index)}
            >
              ×
            </button>
          </div>
        ))}
      </div>
      <button
        type="button"
        className="mt-2 text-sm text-blue-600 hover:text-blue-800"
        onClick={add}
      >
        + 添加
      </button>
    </div>
  );
}
