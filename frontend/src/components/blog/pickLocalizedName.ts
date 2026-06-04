import type { Locale } from "@/lib/locale";

export function pickLocalizedName(zh: string, en: string, locale: Locale): string {
  if (locale === "en" && en) return en;
  return zh || en;
}
