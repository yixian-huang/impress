import { useMemo, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import type { SectionData, SectionSettings } from "../types";
import { useSectionRegistry } from "@/plugins/hooks";
import { resolveLocale } from "@/utils/locale";

interface SectionWrapperProps {
  settings?: SectionSettings;
  children: ReactNode;
}

function SectionWrapper({ settings, children }: SectionWrapperProps) {
  if (settings?.hidden) return null;

  const bgClass =
    settings?.background === "primary"
      ? "bg-primary"
      : settings?.background === "surface-alt"
        ? "bg-surface-alt"
        : "bg-surface";

  const padClass =
    settings?.padding === "lg"
      ? "py-section-lg"
      : settings?.padding === "sm"
        ? "py-section-sm"
        : settings?.padding === "none"
          ? ""
          : "py-section";

  return (
    <section className={`${bgClass} ${padClass}`}>
      {children}
    </section>
  );
}

/**
 * Recursively resolve bilingual {zh, en} objects to plain strings for a locale.
 * Leaves non-bilingual values untouched.
 */
function resolveBilingualValue(value: unknown, locale: string): unknown {
  if (value === null || value === undefined) return value;

  if (Array.isArray(value)) {
    return value.map((item) => resolveBilingualValue(item, locale));
  }

  if (typeof value === "object") {
    const obj = value as Record<string, unknown>;
    // Detect bilingual object: has "zh" key (and optionally "en" key).
    // Treat as bilingual even if extra keys exist (e.g., corrupted data with
    // indexed char keys like {0: "+", 1: "8", ..., zh: "...", en: "..."}).
    if ("zh" in obj && typeof obj["zh"] === "string") {
      const keys = Object.keys(obj);
      if (keys.every((k) => k === "zh" || k === "en") || ("en" in obj && typeof obj["en"] === "string")) {
        return (obj[locale] as string) || (obj["zh"] as string) || "";
      }
    }
    // Recursively resolve nested bilingual fields
    const result: Record<string, unknown> = {};
    for (const key of Object.keys(obj)) {
      result[key] = resolveBilingualValue(obj[key], locale);
    }
    // Collapse MediaRef-like objects ({url, alt}) to plain URL string
    if ("url" in result && typeof result.url === "string") {
      const keys = Object.keys(result);
      if (keys.every((k) => k === "url" || k === "alt")) {
        return result.url;
      }
    }
    return result;
  }

  return value;
}

interface SectionRendererProps {
  section: SectionData;
}

export default function SectionRenderer({ section }: SectionRendererProps) {
  const { registry } = useSectionRegistry();
  const { i18n } = useTranslation("common");
  const locale = resolveLocale(i18n.language);
  const Component = registry[section.type];

  const resolvedData = useMemo(() => {
    const raw = section.data || (section as any).props || {};
    return resolveBilingualValue(raw, locale) as Record<string, unknown>;
  }, [section, locale]);

  if (!Component) {
    if (import.meta.env.DEV) {
      return (
        <div className="p-4 bg-yellow-50 text-yellow-800 border border-yellow-200 rounded mx-4 my-2">
          Unknown section type: <code>{section.type}</code>
        </div>
      );
    }
    return null;
  }

  return (
    <SectionWrapper settings={section.settings}>
      <Component data={resolvedData} settings={section.settings} variant={section.variant} />
    </SectionWrapper>
  );
}
