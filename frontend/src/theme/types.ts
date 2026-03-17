export interface SectionData {
  id: string;
  type: string;
  variant?: string;   // layout variant key, defaults to "default"
  locked?: boolean;    // true in template mode — cannot move/delete
  data: Record<string, unknown>;
  settings?: SectionSettings;
}

export interface SectionSettings {
  background?: "surface" | "surface-alt" | "primary" | string;
  padding?: "none" | "sm" | "md" | "lg";
  maxWidth?: "layout" | "full" | string;
  hidden?: boolean;
}

export interface SectionProps<T = Record<string, unknown>> {
  data: T;
  settings?: SectionSettings;
  variant?: string;
}

export interface SectionMeta {
  type: string;
  label: string;
  labelZh: string;
  icon?: string;
}

export interface PageConfig {
  layout?: string;
  sections: SectionData[];
}

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
