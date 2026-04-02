import type { Locale } from "@/api/publicContent";

/**
 * Resolves the current i18n language to a supported locale.
 * Defaults to "zh" for any language starting with "zh", otherwise "en".
 */
export function resolveLocale(language: string): Locale {
  return language === "zh" || language.startsWith("zh-") ? "zh" : "en";
}
