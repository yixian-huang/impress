/**
 * Built-in registration entry for editorial-firm.
 * Implementation lives in monorepo package `@inkless/theme-editorial-firm`.
 *
 * Pages are CMS section-driven (`renderMode: "dynamic"`); custom sections
 * land in later tasks — no host page loaders required.
 */
import { createEditorialFirmTheme } from "@inkless/theme-editorial-firm";

export {
  EDITORIAL_FIRM_THEME_ID,
  EDITORIAL_FIRM_CONTRACT_VERSION,
  EDITORIAL_PAGE_CONTENT_KEYS,
  EDITORIAL_DEFAULT_LAYOUT,
  createEditorialFirmTheme,
  editorialFirmTokens,
  noirGalleryTokens,
} from "@inkless/theme-editorial-firm";

export const editorialFirmTheme = createEditorialFirmTheme();
