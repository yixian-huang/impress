/**
 * Built-in registration entry for corporate-classic.
 * Implementation lives in monorepo package `@inkless/theme-corporate-classic`.
 *
 * Page UI components remain under host `frontend/src/pages/*` and are attached
 * as lazy loaders so the theme package never deep-imports `@/…`.
 */
import { createCorporateClassicTheme } from "@inkless/theme-corporate-classic";

export {
  corporateClassicTokens,
  CORPORATE_CLASSIC_THEME_ID,
  CORPORATE_CLASSIC_CONTRACT_VERSION,
  CORPORATE_DEFAULT_LAYOUT,
  CORPORATE_PAGE_CONTENT_KEYS,
  createCorporateClassicTheme,
} from "@inkless/theme-corporate-classic";

export const corporateClassicTheme = createCorporateClassicTheme({
  loaders: {
    home: () => import("@/pages/home/page"),
    about: () => import("@/pages/about/page"),
    advantages: () => import("@/pages/advantages/page"),
    "core-services": () => import("@/pages/core-services/page"),
    cases: () => import("@/pages/cases/page"),
    experts: () => import("@/pages/experts/page"),
    contact: () => import("@/pages/contact/page"),
  },
});
