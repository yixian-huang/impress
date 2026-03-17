export type { FieldType, FieldSchema } from "@/theme/types";
import type { FieldSchema } from "@/theme/types";

export interface FieldProps {
  schema: FieldSchema;
  value: unknown;
  onChange: (value: unknown) => void;
}
