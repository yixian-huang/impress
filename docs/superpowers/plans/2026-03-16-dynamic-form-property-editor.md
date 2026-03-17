# Dynamic Form Property Editor — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the raw JSON textarea in the admin page editor's right panel with a schema-driven dynamic form that renders appropriate controls for each section type.

**Architecture:** A `FieldSchema[]` array is added to each section's metadata in the registry. A generic `DynamicForm` component reads the schema and recursively renders typed field controls (text, bilingual, media, array, etc.). The existing JSON editor is preserved as a fallback toggle.

**Tech Stack:** React 19, TypeScript, Tailwind CSS 3, existing ImagePickerModal component, Vitest + Testing Library

**Spec:** `docs/superpowers/specs/2026-03-16-dynamic-form-property-editor-design.md`

---

## File Structure

### New files

| File | Responsibility |
|------|---------------|
| `frontend/src/pages/admin/pages/editor/components/fields/types.ts` | Re-exports `FieldSchema`/`FieldType` from `@/theme/types`, defines `FieldProps` interface |
| `frontend/src/pages/admin/pages/editor/components/fields/TextField.tsx` | Single-line text input |
| `frontend/src/pages/admin/pages/editor/components/fields/TextareaField.tsx` | Multi-line text input |
| `frontend/src/pages/admin/pages/editor/components/fields/BilingualField.tsx` | zh/en tab-switching single-line input |
| `frontend/src/pages/admin/pages/editor/components/fields/BilingualTextareaField.tsx` | zh/en tab-switching multi-line input |
| `frontend/src/pages/admin/pages/editor/components/fields/SelectField.tsx` | Dropdown with numeric coercion |
| `frontend/src/pages/admin/pages/editor/components/fields/NumberField.tsx` | Number input |
| `frontend/src/pages/admin/pages/editor/components/fields/BooleanField.tsx` | Toggle switch |
| `frontend/src/pages/admin/pages/editor/components/fields/ColorField.tsx` | Text input + color swatch |
| `frontend/src/pages/admin/pages/editor/components/fields/MediaField.tsx` | URL input + ImagePickerModal + thumbnail |
| `frontend/src/pages/admin/pages/editor/components/fields/ArrayField.tsx` | Add/delete/reorder list with recursive sub-forms |
| `frontend/src/pages/admin/pages/editor/components/fields/StringArrayField.tsx` | Add/delete/reorder list of plain string inputs |
| `frontend/src/pages/admin/pages/editor/components/fields/index.ts` | Re-exports all field components |
| `frontend/src/pages/admin/pages/editor/components/FieldRenderer.tsx` | Dispatches schema field to correct component |
| `frontend/src/pages/admin/pages/editor/components/DynamicForm.tsx` | Top-level form: iterates schema, renders FieldRenderers |
| `frontend/src/pages/admin/pages/editor/components/SectionSettings.tsx` | Fixed form for settings (background/padding/maxWidth/hidden) |
| `frontend/src/pages/admin/pages/editor/components/PropertiesPanel.tsx` | Orchestrates header + DynamicForm + SectionSettings + JSON fallback |
| `frontend/src/theme/sectionSchemas.ts` | All 9 section schema definitions + settings schema |
| `frontend/src/pages/admin/pages/editor/components/__tests__/DynamicForm.test.tsx` | Tests for DynamicForm + FieldRenderer |
| `frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx` | Tests for individual field components |
| `frontend/src/pages/admin/pages/editor/components/__tests__/PropertiesPanel.test.tsx` | Integration test for PropertiesPanel with schema lookup + fallback |

### Modified files

| File | Changes |
|------|---------|
| `frontend/src/theme/types.ts` | Add `FieldSchema`, `FieldType` exports (canonical types) |
| `frontend/src/theme/sections/index.ts` | Add `sectionSchemas` export mapping section type → FieldSchema[] |
| `frontend/src/pages/admin/pages/editor/page.tsx` | Replace right-panel JSON textarea with `<PropertiesPanel>`, add `updateSectionSettings` callback |

---

## Chunk 1: Types, Simple Fields, and Tests

### Task 1: Define FieldSchema types

**Files:**
- Modify: `frontend/src/theme/types.ts` (canonical type definitions)
- Create: `frontend/src/pages/admin/pages/editor/components/fields/types.ts` (re-exports + FieldProps)

- [ ] **Step 1: Add FieldSchema and FieldType to theme/types.ts**

Append to the end of `frontend/src/theme/types.ts`:

```ts
// --- Field schema types (used by section schemas and dynamic form) ---

export type FieldType =
  | "text"
  | "textarea"
  | "bilingual"
  | "bilingual-textarea"
  | "media"
  | "color"
  | "select"
  | "number"
  | "boolean"
  | "array"
  | "string-array";

export interface FieldSchema {
  key: string;
  type: FieldType;
  label: string;
  placeholder?: string;
  defaultValue?: unknown;
  hidden?: boolean;
  options?: { label: string; value: string | number }[];
  itemSchema?: FieldSchema[];
}
```

- [ ] **Step 2: Create fields/types.ts (re-export + FieldProps)**

```ts
// frontend/src/pages/admin/pages/editor/components/fields/types.ts

// Re-export canonical types from theme layer
export type { FieldType, FieldSchema } from "@/theme/types";
import type { FieldSchema } from "@/theme/types";

/** Common props passed to every field component */
export interface FieldProps {
  schema: FieldSchema;
  value: unknown;
  onChange: (value: unknown) => void;
}
```

This avoids architectural inversion: `theme/sectionSchemas.ts` imports from `@/theme/types` (same layer), and field components import from their local `types.ts` which re-exports from the canonical source.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/theme/types.ts frontend/src/pages/admin/pages/editor/components/fields/types.ts
git commit -m "feat(editor): add FieldSchema type definitions to theme/types.ts"
```

---

### Task 2: Implement simple field components (Text, Textarea, Number, Boolean, Color, Select)

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/components/fields/TextField.tsx`
- Create: `frontend/src/pages/admin/pages/editor/components/fields/TextareaField.tsx`
- Create: `frontend/src/pages/admin/pages/editor/components/fields/NumberField.tsx`
- Create: `frontend/src/pages/admin/pages/editor/components/fields/BooleanField.tsx`
- Create: `frontend/src/pages/admin/pages/editor/components/fields/ColorField.tsx`
- Create: `frontend/src/pages/admin/pages/editor/components/fields/SelectField.tsx`
- Test: `frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`

- [ ] **Step 1: Write tests for all 6 simple fields**

Test file covers:
- TextField: renders input with value, calls onChange on input
- TextareaField: renders textarea, calls onChange
- NumberField: renders number input, onChange returns number (not string)
- BooleanField: renders checkbox/toggle, onChange returns boolean
- ColorField: renders text input + color swatch div with background matching value
- SelectField: renders select with options, onChange returns number for numeric options and string for string options

Each test uses `render()` + `screen` + `fireEvent`/`userEvent`. Pattern:

```tsx
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
// import each field component...

describe("TextField", () => {
  it("renders value and calls onChange", () => {
    const onChange = vi.fn();
    render(<TextField schema={{ key: "title", type: "text", label: "标题" }} value="hello" onChange={onChange} />);
    expect(screen.getByDisplayValue("hello")).toBeInTheDocument();
    fireEvent.change(screen.getByDisplayValue("hello"), { target: { value: "world" } });
    expect(onChange).toHaveBeenCalledWith("world");
  });
});

// Similar pattern for each...

describe("SelectField", () => {
  it("coerces numeric option values", () => {
    const onChange = vi.fn();
    render(
      <SelectField
        schema={{ key: "columns", type: "select", label: "列数", options: [
          { label: "2列", value: 2 }, { label: "3列", value: 3 },
        ]}}
        value={2}
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByRole("combobox"), { target: { value: "3" } });
    expect(onChange).toHaveBeenCalledWith(3); // number, not string
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: FAIL — components don't exist yet

- [ ] **Step 3: Implement all 6 simple field components**

Each component follows the same pattern — receives `FieldProps`, renders the appropriate HTML control with Tailwind styling, and calls `onChange` with the typed value.

Key implementation details:
- All fields render a `<label>` with `schema.label` and the control below it
- `SelectField`: detect if any option has `typeof value === "number"`, if so coerce `e.target.value` with `Number()`
- `ColorField`: render `<input type="text">` + a 6x6 `<div>` with `style={{ backgroundColor: value }}` as swatch
- `NumberField`: `onChange` calls `onChange(Number(e.target.value))` (or `onChange(undefined)` if empty)
- `BooleanField`: render `<input type="checkbox">`, `onChange(e.target.checked)`
- Tailwind classes: `w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm` for inputs (consistent with existing editor styling)

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/components/fields/ frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx
git commit -m "feat(editor): implement 6 simple field components with tests"
```

---

### Task 3: Implement BilingualField and BilingualTextareaField

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/components/fields/BilingualField.tsx`
- Create: `frontend/src/pages/admin/pages/editor/components/fields/BilingualTextareaField.tsx`
- Test: `frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx` (append)

- [ ] **Step 1: Add tests for bilingual fields**

```tsx
describe("BilingualField", () => {
  it("renders zh tab by default with zh value", () => {
    const onChange = vi.fn();
    render(<BilingualField schema={{ key: "title", type: "bilingual", label: "标题" }} value={{ zh: "你好", en: "Hello" }} onChange={onChange} />);
    expect(screen.getByDisplayValue("你好")).toBeInTheDocument();
  });

  it("switches to en tab and shows en value", () => {
    const onChange = vi.fn();
    render(<BilingualField schema={{ key: "title", type: "bilingual", label: "标题" }} value={{ zh: "你好", en: "Hello" }} onChange={onChange} />);
    fireEvent.click(screen.getByText("en"));
    expect(screen.getByDisplayValue("Hello")).toBeInTheDocument();
  });

  it("handles legacy plain string value", () => {
    const onChange = vi.fn();
    render(<BilingualField schema={{ key: "title", type: "bilingual", label: "标题" }} value="旧数据" onChange={onChange} />);
    expect(screen.getByDisplayValue("旧数据")).toBeInTheDocument();
  });

  it("onChange emits {zh, en} object", () => {
    const onChange = vi.fn();
    render(<BilingualField schema={{ key: "title", type: "bilingual", label: "标题" }} value={{ zh: "你好", en: "Hello" }} onChange={onChange} />);
    fireEvent.change(screen.getByDisplayValue("你好"), { target: { value: "世界" } });
    expect(onChange).toHaveBeenCalledWith({ zh: "世界", en: "Hello" });
  });
});
```

Similar tests for `BilingualTextareaField` (uses `<textarea>` instead of `<input>`).

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement BilingualField and BilingualTextareaField**

Key logic (shared between both, differing only in input element):

```tsx
function BilingualField({ schema, value, onChange }: FieldProps) {
  const [tab, setTab] = useState<"zh" | "en">("zh");

  // Normalize: plain string → { zh: string, en: "" }
  const normalized: { zh: string; en: string } =
    typeof value === "string"
      ? { zh: value, en: "" }
      : (value as { zh: string; en: string }) ?? { zh: "", en: "" };

  const handleChange = (text: string) => {
    onChange({ ...normalized, [tab]: text });
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-1">
        <label className="text-xs font-medium text-gray-600">{schema.label}</label>
        <div className="flex text-xs">
          <button onClick={() => setTab("zh")} className={tab === "zh" ? "font-bold text-blue-600" : "text-gray-400"}>zh</button>
          <span className="mx-1 text-gray-300">|</span>
          <button onClick={() => setTab("en")} className={tab === "en" ? "font-bold text-blue-600" : "text-gray-400"}>en</button>
        </div>
      </div>
      <input value={normalized[tab]} onChange={(e) => handleChange(e.target.value)} ... />
    </div>
  );
}
```

`BilingualTextareaField` is identical but uses `<textarea rows={3}>`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/components/fields/BilingualField.tsx frontend/src/pages/admin/pages/editor/components/fields/BilingualTextareaField.tsx frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx
git commit -m "feat(editor): implement bilingual field components with zh/en tabs"
```

---

### Task 4: Implement MediaField

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/components/fields/MediaField.tsx`
- Test: `frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx` (append)

- [ ] **Step 1: Add tests for MediaField**

```tsx
describe("MediaField", () => {
  it("renders URL input with current value", () => {
    render(<MediaField schema={{ key: "image", type: "media", label: "图片" }} value="/uploads/test.jpg" onChange={vi.fn()} />);
    expect(screen.getByDisplayValue("/uploads/test.jpg")).toBeInTheDocument();
  });

  it("shows thumbnail when value is present", () => {
    render(<MediaField schema={{ key: "image", type: "media", label: "图片" }} value="/uploads/test.jpg" onChange={vi.fn()} />);
    const img = screen.getByRole("img");
    expect(img).toHaveAttribute("src", "/uploads/test.jpg");
  });

  it("shows '选择' button", () => {
    render(<MediaField schema={{ key: "image", type: "media", label: "图片" }} value="" onChange={vi.fn()} />);
    expect(screen.getByText("选择")).toBeInTheDocument();
  });

  it("calls onChange when URL input changes", () => {
    const onChange = vi.fn();
    render(<MediaField schema={{ key: "image", type: "media", label: "图片" }} value="" onChange={onChange} />);
    fireEvent.change(screen.getByRole("textbox"), { target: { value: "/new.jpg" } });
    expect(onChange).toHaveBeenCalledWith("/new.jpg");
  });
});
```

Note: ImagePickerModal integration is tested via the button click opening the modal. Since the modal is a complex component with API calls, we test that clicking "选择" sets `showPicker` state (the modal renders conditionally). Full E2E testing of the picker is out of scope.

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement MediaField**

```tsx
import { useState } from "react";
import ImagePickerModal from "@/components/admin/ImagePickerModal";
import type { FieldProps } from "./types";

export default function MediaField({ schema, value, onChange }: FieldProps) {
  const [showPicker, setShowPicker] = useState(false);
  const url = typeof value === "string" ? value : "";

  return (
    <div>
      <label className="block text-xs font-medium text-gray-600 mb-1">{schema.label}</label>
      <div className="flex gap-2">
        <input
          type="text"
          value={url}
          onChange={(e) => onChange(e.target.value)}
          placeholder={schema.placeholder || "输入图片 URL"}
          className="flex-1 border border-gray-300 rounded-md px-3 py-1.5 text-sm"
        />
        <button
          type="button"
          onClick={() => setShowPicker(true)}
          className="px-3 py-1.5 text-sm border border-gray-300 rounded-md hover:bg-gray-50"
        >
          选择
        </button>
      </div>
      {url && (
        <img src={url} alt="" className="mt-2 w-20 h-20 object-cover rounded border border-gray-200" />
      )}
      <ImagePickerModal
        open={showPicker}
        onClose={() => setShowPicker(false)}
        onSelect={(item) => { onChange(item.url); setShowPicker(false); }}
      />
    </div>
  );
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/components/fields/MediaField.tsx frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx
git commit -m "feat(editor): implement MediaField with ImagePickerModal integration"
```

---

## Chunk 2: Array Fields, FieldRenderer, DynamicForm

### Task 5: Implement StringArrayField

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/components/fields/StringArrayField.tsx`
- Test: `frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx` (append)

- [ ] **Step 1: Add tests for StringArrayField**

```tsx
describe("StringArrayField", () => {
  it("renders each string as an input", () => {
    render(<StringArrayField schema={{ key: "items", type: "string-array", label: "检查项" }} value={["A", "B"]} onChange={vi.fn()} />);
    expect(screen.getByDisplayValue("A")).toBeInTheDocument();
    expect(screen.getByDisplayValue("B")).toBeInTheDocument();
  });

  it("adds a new empty item on + click", () => {
    const onChange = vi.fn();
    render(<StringArrayField schema={{ key: "items", type: "string-array", label: "检查项" }} value={["A"]} onChange={onChange} />);
    fireEvent.click(screen.getByText("+ 添加"));
    expect(onChange).toHaveBeenCalledWith(["A", ""]);
  });

  it("deletes an item", () => {
    const onChange = vi.fn();
    render(<StringArrayField schema={{ key: "items", type: "string-array", label: "检查项" }} value={["A", "B"]} onChange={onChange} />);
    const deleteButtons = screen.getAllByTitle("删除");
    fireEvent.click(deleteButtons[0]);
    expect(onChange).toHaveBeenCalledWith(["B"]);
  });

  it("handles empty/undefined value as empty array", () => {
    render(<StringArrayField schema={{ key: "items", type: "string-array", label: "检查项" }} value={undefined} onChange={vi.fn()} />);
    expect(screen.getByText("+ 添加")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement StringArrayField**

Renders a list of `<input>` elements, each with a delete button. Bottom has "+ 添加" button. Move up/down buttons for reorder. `onChange` emits the full updated `string[]`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/components/fields/StringArrayField.tsx frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx
git commit -m "feat(editor): implement StringArrayField for plain string lists"
```

---

### Task 6: Implement FieldRenderer and field index

FieldRenderer must exist before ArrayField, because ArrayField imports it to render sub-fields.

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/components/fields/index.ts`
- Create: `frontend/src/pages/admin/pages/editor/components/FieldRenderer.tsx`

- [ ] **Step 1: Create fields/index.ts barrel export**

```ts
export { default as TextField } from "./TextField";
export { default as TextareaField } from "./TextareaField";
export { default as BilingualField } from "./BilingualField";
export { default as BilingualTextareaField } from "./BilingualTextareaField";
export { default as MediaField } from "./MediaField";
export { default as ColorField } from "./ColorField";
export { default as SelectField } from "./SelectField";
export { default as NumberField } from "./NumberField";
export { default as BooleanField } from "./BooleanField";
export { default as StringArrayField } from "./StringArrayField";
// ArrayField added after Task 7
export type { FieldSchema, FieldType, FieldProps } from "./types";
```

- [ ] **Step 2: Create FieldRenderer**

Note: ArrayField and StringArrayField are initially included in the map even though ArrayField doesn't exist yet. This is fine — it's a lazy reference that won't fail until actually rendered. ArrayField will be added to the barrel export in Task 7.

```tsx
import type { FieldSchema } from "./fields/types";
import {
  TextField, TextareaField, BilingualField, BilingualTextareaField,
  MediaField, ColorField, SelectField, NumberField, BooleanField,
  StringArrayField,
} from "./fields";

// Lazy-import ArrayField to break circular dependency
// (ArrayField imports FieldRenderer, FieldRenderer imports ArrayField)
import type { FieldProps } from "./fields/types";
import { lazy, Suspense, type ComponentType } from "react";
const ArrayField = lazy(() => import("./fields/ArrayField"));

const FIELD_MAP: Record<string, ComponentType<FieldProps>> = {
  text: TextField,
  textarea: TextareaField,
  bilingual: BilingualField,
  "bilingual-textarea": BilingualTextareaField,
  media: MediaField,
  color: ColorField,
  select: SelectField,
  number: NumberField,
  boolean: BooleanField,
  "string-array": StringArrayField,
};

interface FieldRendererProps {
  schema: FieldSchema;
  value: unknown;
  onChange: (value: unknown) => void;
}

export default function FieldRenderer({ schema, value, onChange }: FieldRendererProps) {
  if (schema.hidden) return null;

  // Handle array type separately due to lazy import
  if (schema.type === "array") {
    return (
      <Suspense fallback={null}>
        <ArrayField schema={schema} value={value} onChange={onChange} />
      </Suspense>
    );
  }

  const Component = FIELD_MAP[schema.type];
  if (!Component) return null;
  return <Component schema={schema} value={value} onChange={onChange} />;
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/components/FieldRenderer.tsx frontend/src/pages/admin/pages/editor/components/fields/index.ts
git commit -m "feat(editor): add FieldRenderer dispatcher and field barrel export"
```

---

### Task 7: Implement ArrayField (recursive object arrays)

Depends on: Task 6 (FieldRenderer must exist for ArrayField to import).

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/components/fields/ArrayField.tsx`
- Modify: `frontend/src/pages/admin/pages/editor/components/fields/index.ts` (add ArrayField export)
- Test: `frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx` (append)

- [ ] **Step 1: Add tests for ArrayField**

```tsx
describe("ArrayField", () => {
  const cardSchema: FieldSchema = {
    key: "cards", type: "array", label: "卡片",
    itemSchema: [
      { key: "title", type: "bilingual", label: "标题" },
      { key: "image", type: "media", label: "图片" },
    ],
  };

  it("renders each item as a card with sub-fields", () => {
    render(
      <ArrayField
        schema={cardSchema}
        value={[{ title: { zh: "卡1", en: "Card1" }, image: "/a.jpg" }]}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("卡1")).toBeInTheDocument();
    expect(screen.getByDisplayValue("/a.jpg")).toBeInTheDocument();
  });

  it("adds a new item with empty defaults", () => {
    const onChange = vi.fn();
    render(<ArrayField schema={cardSchema} value={[]} onChange={onChange} />);
    fireEvent.click(screen.getByText("+ 添加"));
    expect(onChange).toHaveBeenCalledTimes(1);
    const newItems = onChange.mock.calls[0][0];
    expect(newItems).toHaveLength(1);
  });

  it("deletes an item", () => {
    const onChange = vi.fn();
    render(
      <ArrayField
        schema={cardSchema}
        value={[{ title: { zh: "A", en: "" } }, { title: { zh: "B", en: "" } }]}
        onChange={onChange}
      />
    );
    const deleteButtons = screen.getAllByTitle("删除");
    fireEvent.click(deleteButtons[0]);
    const newItems = onChange.mock.calls[0][0];
    expect(newItems).toHaveLength(1);
    expect(newItems[0].title.zh).toBe("B");
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement ArrayField**

Key implementation details:
- Receives `value` as `unknown[]` (normalize to `[]` if falsy)
- Each item renders as a bordered card with: index summary line, move up/down/delete buttons, sub-form fields
- Summary line: `"{label} {i+1}"` + first bilingual/text field value from `itemSchema` as preview
- Sub-fields: import and use `FieldRenderer` for each `itemSchema` field
- `+ 添加` button: creates a new item object from `itemSchema` defaults. For fields with `hidden: true` and key `"id"`, auto-generate via `crypto.randomUUID()`
- Move up/down: swap items in array, emit new array
- Delete: filter item out, emit new array
- Each item's sub-field `onChange`: update that field in the item object immutably, emit new full array

```tsx
import FieldRenderer from "../FieldRenderer";
import type { FieldProps } from "./types";

export default function ArrayField({ schema, value, onChange }: FieldProps) {
  const items = Array.isArray(value) ? value : [];
  const itemSchema = schema.itemSchema || [];

  const createDefaultItem = (): Record<string, unknown> => {
    const item: Record<string, unknown> = {};
    for (const field of itemSchema) {
      if (field.hidden && field.key === "id") {
        item[field.key] = crypto.randomUUID();
      } else if (field.defaultValue !== undefined) {
        item[field.key] = field.defaultValue;
      }
    }
    return item;
  };

  const updateItem = (index: number, key: string, val: unknown) => {
    const next = items.map((item: any, i: number) =>
      i === index ? { ...item, [key]: val } : item
    );
    onChange(next);
  };

  const addItem = () => onChange([...items, createDefaultItem()]);
  const deleteItem = (index: number) => onChange(items.filter((_: any, i: number) => i !== index));
  const moveItem = (from: number, to: number) => {
    const next = [...items];
    const [moved] = next.splice(from, 1);
    next.splice(to, 0, moved);
    onChange(next);
  };
  // ... render cards with FieldRenderer for each sub-field
}
```

- [ ] **Step 4: Add ArrayField to fields/index.ts barrel export**

Add to `frontend/src/pages/admin/pages/editor/components/fields/index.ts`:

```ts
export { default as ArrayField } from "./ArrayField";
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/fields.test.tsx`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/components/fields/ArrayField.tsx frontend/src/pages/admin/pages/editor/components/fields/index.ts frontend/src/pages/admin/pages/editor/components/__tests__/fields.test.tsx
git commit -m "feat(editor): implement ArrayField with recursive sub-forms"
```

---

### Task 8: Implement DynamicForm

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/components/DynamicForm.tsx`
- Test: `frontend/src/pages/admin/pages/editor/components/__tests__/DynamicForm.test.tsx`

- [ ] **Step 1: Write tests for DynamicForm**

```tsx
describe("DynamicForm", () => {
  const heroSchema: FieldSchema[] = [
    { key: "title", type: "bilingual", label: "标题" },
    { key: "backgroundColor", type: "color", label: "背景色" },
  ];

  it("renders a field for each schema entry", () => {
    render(
      <DynamicForm
        schema={heroSchema}
        data={{ title: { zh: "你好", en: "Hi" }, backgroundColor: "#fff" }}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("你好")).toBeInTheDocument();
    expect(screen.getByDisplayValue("#fff")).toBeInTheDocument();
  });

  it("calls onChange with updated data when a field changes", () => {
    const onChange = vi.fn();
    render(
      <DynamicForm
        schema={heroSchema}
        data={{ title: { zh: "你好", en: "Hi" }, backgroundColor: "#fff" }}
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByDisplayValue("#fff"), { target: { value: "#000" } });
    expect(onChange).toHaveBeenCalledWith({
      title: { zh: "你好", en: "Hi" },
      backgroundColor: "#000",
    });
  });

  it("skips hidden fields", () => {
    render(
      <DynamicForm
        schema={[{ key: "id", type: "text", label: "ID", hidden: true }]}
        data={{ id: "abc" }}
        onChange={vi.fn()}
      />
    );
    expect(screen.queryByDisplayValue("abc")).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/DynamicForm.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement DynamicForm**

```tsx
import FieldRenderer from "./FieldRenderer";
import type { FieldSchema } from "./fields/types";

interface DynamicFormProps {
  schema: FieldSchema[];
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
}

export default function DynamicForm({ schema, data, onChange }: DynamicFormProps) {
  const handleFieldChange = (key: string, value: unknown) => {
    onChange({ ...data, [key]: value });
  };

  return (
    <div className="space-y-4">
      {schema.map((field) => (
        <FieldRenderer
          key={field.key}
          schema={field}
          value={data[field.key]}
          onChange={(val) => handleFieldChange(field.key, val)}
        />
      ))}
    </div>
  );
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/DynamicForm.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/components/DynamicForm.tsx frontend/src/pages/admin/pages/editor/components/__tests__/DynamicForm.test.tsx
git commit -m "feat(editor): implement DynamicForm component"
```

---

## Chunk 3: Section Schemas, SectionSettings, PropertiesPanel, Integration

### Task 9: Define all 9 section schemas

**Files:**
- Create: `frontend/src/theme/sectionSchemas.ts`
- Modify: `frontend/src/theme/sections/index.ts`

- [ ] **Step 1: Create sectionSchemas.ts**

Define a `Record<string, FieldSchema[]>` mapping each section type to its schema, based on the spec's Section Schemas tables. Also export `settingsSchema: FieldSchema[]` for the shared settings.

Full listing (abbreviated here, see spec for complete field lists):

```ts
import type { FieldSchema } from "@/theme/types";

export const sectionSchemas: Record<string, FieldSchema[]> = {
  hero: [
    { key: "title", type: "bilingual", label: "标题" },
    { key: "subtitle", type: "bilingual", label: "副标题" },
    { key: "label", type: "bilingual", label: "标签文字" },
    { key: "backgroundImage", type: "media", label: "背景图" },
    { key: "backgroundColor", type: "color", label: "背景色" },
  ],
  "text-image": [ /* ... per spec ... */ ],
  "card-grid": [ /* ... per spec, including cards array with titleEn ... */ ],
  "service-cards": [ /* ... */ ],
  "team-grid": [ /* ... including experts with hidden id field ... */ ],
  checklist: [ /* ... categories array with string-array items ... */ ],
  "contact-form": [ /* ... all 12 fields ... */ ],
  "company-profile": [ /* ... */ ],
  "rich-text": [ /* ... */ ],
};

export const settingsSchema: FieldSchema[] = [
  { key: "background", type: "select", label: "背景", options: [
    { label: "默认", value: "surface" },
    { label: "交替", value: "surface-alt" },
    { label: "主题色", value: "primary" },
  ]},
  { key: "padding", type: "select", label: "内边距", options: [
    { label: "无", value: "none" },
    { label: "小", value: "sm" },
    { label: "中", value: "md" },
    { label: "大", value: "lg" },
  ]},
  { key: "maxWidth", type: "select", label: "最大宽度", options: [
    { label: "标准", value: "layout" },
    { label: "全宽", value: "full" },
  ]},
  { key: "hidden", type: "boolean", label: "隐藏" },
];
```

- [ ] **Step 2: Re-export from sections/index.ts**

Add to `frontend/src/theme/sections/index.ts`:

```ts
export { sectionSchemas, settingsSchema } from "../sectionSchemas";
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/theme/sectionSchemas.ts frontend/src/theme/sections/index.ts
git commit -m "feat(editor): define schemas for all 9 section types + settings"
```

---

### Task 10: Implement SectionSettings component

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/components/SectionSettings.tsx`

- [ ] **Step 1: Implement SectionSettings**

A simple wrapper that renders a collapsible `<details>` element titled "显示设置", with a `DynamicForm` inside using the `settingsSchema`.

```tsx
import DynamicForm from "./DynamicForm";
import { settingsSchema } from "@/theme/sectionSchemas";
import type { SectionSettings as SectionSettingsType } from "@/theme/types";

interface Props {
  settings: SectionSettingsType;
  onChange: (settings: SectionSettingsType) => void;
}

export default function SectionSettings({ settings, onChange }: Props) {
  return (
    <details className="border-t border-gray-200 pt-3 mt-4">
      <summary className="text-xs font-semibold text-gray-600 cursor-pointer select-none">
        显示设置
      </summary>
      <div className="mt-3">
        <DynamicForm
          schema={settingsSchema}
          data={settings as Record<string, unknown>}
          onChange={(data) => onChange(data as SectionSettingsType)}
        />
      </div>
    </details>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/components/SectionSettings.tsx
git commit -m "feat(editor): implement SectionSettings collapsible form"
```

---

### Task 11: Implement PropertiesPanel

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/components/PropertiesPanel.tsx`
- Test: `frontend/src/pages/admin/pages/editor/components/__tests__/PropertiesPanel.test.tsx`

- [ ] **Step 1: Write tests for PropertiesPanel**

```tsx
describe("PropertiesPanel", () => {
  it("renders DynamicForm when section has schema", () => {
    render(
      <PropertiesPanel
        section={{ id: "1", type: "hero", data: { title: { zh: "Hi", en: "" } }, settings: {} }}
        onDataChange={vi.fn()}
        onSettingsChange={vi.fn()}
      />
    );
    // Should show form field, not JSON textarea
    expect(screen.getByDisplayValue("Hi")).toBeInTheDocument();
    expect(screen.queryByText("数据 (JSON)")).not.toBeInTheDocument();
  });

  it("falls back to JSON editor when no schema exists", () => {
    render(
      <PropertiesPanel
        section={{ id: "1", type: "unknown-type", data: { foo: "bar" }, settings: {} }}
        onDataChange={vi.fn()}
        onSettingsChange={vi.fn()}
      />
    );
    // Should show JSON textarea
    expect(screen.getByText("数据 (JSON)")).toBeInTheDocument();
  });

  it("toggles between form and JSON mode", () => {
    render(
      <PropertiesPanel
        section={{ id: "1", type: "hero", data: { title: { zh: "Hi", en: "" } }, settings: {} }}
        onDataChange={vi.fn()}
        onSettingsChange={vi.fn()}
      />
    );
    fireEvent.click(screen.getByText("切换到 JSON 编辑"));
    expect(screen.getByText("切换到表单编辑")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/PropertiesPanel.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement PropertiesPanel**

Key decisions:
- Schema lookup uses the static `sectionSchemas` map (no context needed — schemas are compile-time constants).
- JSON fallback has two modes: (a) no schema exists → always JSON, uncontrolled textarea matching original editor behavior; (b) user toggled to JSON → controlled textarea with `jsonText` state, apply on switch back.

```tsx
import { useState } from "react";
import DynamicForm from "./DynamicForm";
import SectionSettings from "./SectionSettings";
import { sectionSchemas } from "@/theme/sectionSchemas";
import type { SectionData, SectionSettings as SectionSettingsType } from "@/theme/types";

interface Props {
  section: SectionData;
  onDataChange: (data: Record<string, unknown>) => void;
  onSettingsChange: (settings: SectionSettingsType) => void;
}

export default function PropertiesPanel({ section, onDataChange, onSettingsChange }: Props) {
  const schema = sectionSchemas[section.type];
  const [jsonMode, setJsonMode] = useState(false);
  const [jsonText, setJsonText] = useState("");

  const switchToJson = () => {
    setJsonText(JSON.stringify(section.data, null, 2));
    setJsonMode(true);
  };

  const switchToForm = () => {
    try {
      const parsed = JSON.parse(jsonText);
      onDataChange(parsed);
      setJsonMode(false);
    } catch { /* keep in JSON mode if invalid */ }
  };

  // Determine which editor to show
  const showJsonEditor = !schema || jsonMode;

  return (
    <div>
      {/* Header */}
      <div className="mb-3">
        <span className="text-xs text-gray-500">
          类型: {section.type}
          {section.variant ? ` / ${section.variant}` : ""}
          {section.locked ? " (锁定)" : ""}
        </span>
      </div>

      {showJsonEditor ? (
        /* JSON editor — two sub-cases handled by jsonMode flag */
        <div>
          <label className="block text-xs font-medium text-gray-600 mb-1">数据 (JSON)</label>
          {jsonMode ? (
            /* User toggled to JSON: controlled textarea, changes buffered in jsonText */
            <textarea
              value={jsonText}
              onChange={(e) => setJsonText(e.target.value)}
              rows={12}
              className="w-full border border-gray-300 rounded-md px-2 py-1.5 text-xs font-mono resize-none"
              spellCheck={false}
            />
          ) : (
            /* No schema fallback: matches original editor behavior — parse-on-change, revert on invalid */
            <textarea
              value={JSON.stringify(section.data, null, 2)}
              onChange={(e) => {
                try { onDataChange(JSON.parse(e.target.value)); } catch {}
              }}
              rows={12}
              className="w-full border border-gray-300 rounded-md px-2 py-1.5 text-xs font-mono resize-none"
              spellCheck={false}
            />
          )}
        </div>
      ) : (
        /* Dynamic form */
        <DynamicForm schema={schema} data={section.data} onChange={onDataChange} />
      )}

      {/* Settings */}
      <SectionSettings
        settings={section.settings || {}}
        onChange={onSettingsChange}
      />

      {/* Toggle link — only shown when schema exists (otherwise always JSON) */}
      {schema && (
        <div className="mt-3 text-center">
          <button
            onClick={jsonMode ? switchToForm : switchToJson}
            className="text-xs text-gray-400 hover:text-blue-500 underline"
          >
            {jsonMode ? "切换到表单编辑" : "切换到 JSON 编辑"}
          </button>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && pnpm test -- src/pages/admin/pages/editor/components/__tests__/PropertiesPanel.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/components/PropertiesPanel.tsx frontend/src/pages/admin/pages/editor/components/__tests__/PropertiesPanel.test.tsx
git commit -m "feat(editor): implement PropertiesPanel with schema lookup and JSON fallback"
```

---

### Task 12: Integrate PropertiesPanel into page editor

**Files:**
- Modify: `frontend/src/pages/admin/pages/editor/page.tsx`

- [ ] **Step 1: Add updateSectionSettings callback**

In `PageEditorPage`, add alongside existing `updateSectionData`:

```ts
const updateSectionSettings = useCallback((index: number, settings: SectionSettings) => {
  setSections((prev) =>
    prev.map((s, i) => (i === index ? { ...s, settings } : s)),
  );
}, []);
```

Import `SectionSettings` type from `@/theme/types` (already imported in the file).

- [ ] **Step 2: Replace right sidebar JSON textarea with PropertiesPanel**

Replace lines 804-853 of `page.tsx` (the right sidebar content) with:

```tsx
import PropertiesPanel from "./components/PropertiesPanel";

{/* Right sidebar: section data editor */}
<div className="w-80 flex-shrink-0 border-l border-gray-200 flex flex-col">
  <div className="px-3 py-2 border-b border-gray-200">
    <span className="text-xs font-semibold text-gray-600 uppercase">
      {selectedSection ? (selectedMeta?.labelZh || selectedSection.type) : "属性编辑"}
    </span>
  </div>
  <div className="flex-1 overflow-y-auto p-3">
    {selectedSection ? (
      <PropertiesPanel
        section={selectedSection}
        onDataChange={(data) => updateSectionData(selectedIndex!, data)}
        onSettingsChange={(settings) => updateSectionSettings(selectedIndex!, settings)}
      />
    ) : (
      <div className="text-xs text-gray-400 text-center py-8">
        选择左侧区块以编辑属性
      </div>
    )}
  </div>
</div>
```

- [ ] **Step 3: Run lint and type-check**

Run: `pnpm lint && pnpm type-check`
Expected: PASS with no errors

- [ ] **Step 4: Run all tests**

Run: `pnpm test`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/page.tsx
git commit -m "feat(editor): integrate PropertiesPanel replacing JSON textarea"
```

---

### Task 13: Extract useDragSort hook and add drag support to ArrayField

Per spec: "extract a shared `useDragSort` hook from existing `makeDragHandlers` in page.tsx, then use it in both the section list and ArrayField."

**Files:**
- Create: `frontend/src/pages/admin/pages/editor/hooks/useDragSort.ts`
- Modify: `frontend/src/pages/admin/pages/editor/page.tsx` (replace inline `makeDragHandlers` + `dragIndexRef` with `useDragSort`)
- Modify: `frontend/src/pages/admin/pages/editor/components/fields/ArrayField.tsx` (add drag support using `useDragSort`)

- [ ] **Step 1: Create useDragSort hook**

Extract from existing `makeDragHandlers` pattern in `page.tsx` (lines 390-406):

```ts
// frontend/src/pages/admin/pages/editor/hooks/useDragSort.ts
import { useRef, useCallback } from "react";

export function useDragSort(onReorder: (from: number, to: number) => void) {
  const dragIndexRef = useRef<number | null>(null);

  const makeDragHandlers = useCallback((index: number) => ({
    onDragStart: (e: React.DragEvent) => {
      dragIndexRef.current = index;
      e.dataTransfer.effectAllowed = "move";
    },
    onDragOver: (e: React.DragEvent) => {
      e.preventDefault();
      e.dataTransfer.dropEffect = "move";
    },
    onDrop: (e: React.DragEvent) => {
      e.preventDefault();
      const from = dragIndexRef.current;
      if (from !== null && from !== index) onReorder(from, index);
      dragIndexRef.current = null;
    },
    onDragEnd: () => { dragIndexRef.current = null; },
  }), [onReorder]);

  return { makeDragHandlers };
}
```

- [ ] **Step 2: Replace inline drag logic in page.tsx with useDragSort**

Remove `dragIndexRef` and `makeDragHandlers` from `PageEditorPage`, replace with:

```ts
const { makeDragHandlers } = useDragSort(moveSection);
```

- [ ] **Step 3: Add drag support to ArrayField items**

Import `useDragSort` in `ArrayField.tsx`. Wrap each item card with `draggable` and the drag handlers from `useDragSort(moveItem)`.

- [ ] **Step 4: Run lint, type-check, and tests**

Run: `pnpm lint && pnpm type-check && pnpm test`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/pages/editor/hooks/useDragSort.ts frontend/src/pages/admin/pages/editor/page.tsx frontend/src/pages/admin/pages/editor/components/fields/ArrayField.tsx
git commit -m "refactor(editor): extract useDragSort hook, add drag support to ArrayField"
```

---

### Task 14: Final verification

- [ ] **Step 1: Run full quality gate**

```bash
pnpm lint && pnpm type-check && pnpm test
```

Expected: All pass

- [ ] **Step 2: Manual smoke test**

Start dev server with `make dev`. Navigate to admin page editor. Verify:
1. Select a hero section → right panel shows form fields (标题, 副标题, etc.) instead of JSON
2. Edit a bilingual field → zh/en tabs work, preview updates
3. Edit a card-grid section → cards array shows inline cards with add/delete/drag-reorder
4. Click "切换到 JSON 编辑" → shows JSON textarea; click back → shows form
5. Settings collapsible shows background/padding/maxWidth/hidden controls
6. Save draft → no errors, data persists on reload
7. Select an unknown section type (if any) → falls back to JSON editor
8. Drag-reorder works on both section list (left sidebar) and array items (right panel)

- [ ] **Step 3: Commit any fixes from smoke test**
