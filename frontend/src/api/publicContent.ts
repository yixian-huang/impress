/**
 * Public Content API Client
 * Handles fetching published page configurations from the backend
 */

import axios from "axios";
import { http } from "@/api/http";

export type Locale = "zh" | "en";

export type PageKey =
  | "home"
  | "about"
  | "advantages"
  | "core-services"
  | "cases"
  | "experts"
  | "contact"
  | "global";

export interface PublicPageResponse {
  pageKey: PageKey;
  version: number;
  locale: Locale;
  config: Record<string, unknown>;
}

export interface PublicContentError {
  code: string;
  message: string;
  details?: unknown;
}

/**
 * Fetch published content for a specific page and locale
 * @param pageKey - The page identifier
 * @param locale - The locale to fetch (zh or en)
 * @returns Published page configuration
 * @throws Error with PublicContentError structure on API failure
 */
export async function fetchPublicContent(
  pageKey: PageKey,
  locale: Locale = "zh"
): Promise<PublicPageResponse> {
  try {
    const response = await http.get<PublicPageResponse>(`/public/content/${pageKey}`, {
      params: { locale },
    });
    return response.data;
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const apiError = (error.response?.data as { error?: PublicContentError } | undefined)?.error;
      throw apiError || {
        code: "NETWORK_ERROR",
        message: error.message || "Request failed",
      };
    }

    throw {
      code: "UNKNOWN_ERROR",
      message: "Unknown error occurred",
    } as PublicContentError;
  }
}

/**
 * Fetch draft content for preview purposes (requires admin auth)
 * @param pageKey - The page identifier
 * @param locale - The locale (used for response shaping)
 * @returns Draft config shaped as PublicPageResponse
 * @throws Error with PublicContentError structure on API failure
 */
export async function fetchDraftContent(
  pageKey: PageKey,
  locale: Locale = "zh"
): Promise<PublicPageResponse> {
  try {
    const response = await http.get<{
      pageKey: string;
      config: Record<string, unknown>;
      version: number;
    }>(`/admin/content/${pageKey}/draft`);
    return {
      pageKey: response.data.pageKey as PageKey,
      version: response.data.version,
      locale,
      config: response.data.config,
    };
  } catch (error) {
    if (axios.isAxiosError(error)) {
      const apiError = (error.response?.data as { error?: PublicContentError } | undefined)?.error;
      throw apiError || {
        code: "NETWORK_ERROR",
        message: error.message || "Failed to fetch draft for preview",
      };
    }

    throw {
      code: "UNKNOWN_ERROR",
      message: "Unknown error occurred",
    } as PublicContentError;
  }
}

/**
 * LocalizedText represents a bilingual text field
 */
export interface LocalizedText {
  zh: string;
  en: string;
}

/**
 * Apply locale fallback to localized text
 * Returns the requested locale value, or falls back to zh / en if missing
 * Does NOT mutate the source object
 *
 * Accepts partial shapes such as `{ zh: "…" }` (no `en`) — common after
 * admin saves a single locale or legacy migration.
 *
 * @param text - LocalizedText object or partial / undefined
 * @param locale - Requested locale
 * @returns The text value in the requested locale, or zh/en fallback, or empty string
 */
export function getLocalizedText(
  text: Partial<LocalizedText> | undefined,
  locale: Locale
): string {
  if (!text) return "";

  const preferred = text[locale];
  if (typeof preferred === "string" && preferred.trim().length > 0) {
    return preferred;
  }

  if (typeof text.zh === "string" && text.zh.trim().length > 0) {
    return text.zh;
  }
  if (typeof text.en === "string" && text.en.trim().length > 0) {
    return text.en;
  }

  return "";
}

/** True when value is a leaf bilingual bag: only zh/en keys, string-ish values. */
function isLocalizedTextLike(value: object): boolean {
  const keys = Object.keys(value);
  if (keys.length === 0) return false;
  if (!keys.every((k) => k === "zh" || k === "en")) return false;
  return keys.every((k) => {
    const v = (value as Record<string, unknown>)[k];
    return v === null || v === undefined || typeof v === "string";
  });
}

/**
 * Recursively apply locale selection to a config object
 * For any LocalizedText-like object ({ zh?, en? }), returns the locale-selected value
 * Does NOT mutate the source config
 *
 * Partial bags (only `zh` or only `en`) are flattened too — previously only
 * objects with *both* keys were selected, which left `{ zh: "…" }` as an object
 * and crashed React when components rendered it as a child (error #31).
 *
 * @param config - Page config object
 * @param locale - Target locale
 * @returns Config with locale-selected values
 */
export function normalizeConfigForLocale(
  config: Record<string, unknown>,
  locale: Locale
): Record<string, unknown> {
  const result: Record<string, unknown> = {};

  for (const key in config) {
    const value = config[key];

    if (value === null || value === undefined) {
      result[key] = value;
      continue;
    }

    // LocalizedText-like leaf (both keys, or zh-only / en-only)
    if (
      typeof value === "object" &&
      !Array.isArray(value) &&
      isLocalizedTextLike(value)
    ) {
      result[key] = getLocalizedText(value as Partial<LocalizedText>, locale);
      continue;
    }

    // Recursively process nested objects
    if (typeof value === "object" && !Array.isArray(value)) {
      result[key] = normalizeConfigForLocale(
        value as Record<string, unknown>,
        locale
      );
      continue;
    }

    // Process arrays
    if (Array.isArray(value)) {
      result[key] = value.map((item) => {
        if (typeof item === "object" && item !== null && !Array.isArray(item)) {
          if (isLocalizedTextLike(item)) {
            return getLocalizedText(item as Partial<LocalizedText>, locale);
          }
          return normalizeConfigForLocale(item as Record<string, unknown>, locale);
        }
        return item;
      });
      continue;
    }

    // Primitive values pass through
    result[key] = value;
  }

  return result;
}
