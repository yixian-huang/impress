/**
 * UMD entry — host loads this script and expects a register callback.
 * Built-in path uses `src/index.ts` via workspace import instead.
 */
import { corporateClassicTheme } from "./index";

declare global {
  interface Window {
    __INKLESS_THEME_REGISTER__?: (theme: typeof corporateClassicTheme) => void;
    __IMPRESS_THEME_REGISTER__?: (theme: typeof corporateClassicTheme) => void;
  }
}

const register =
  typeof window !== "undefined"
    ? window.__INKLESS_THEME_REGISTER__ ?? window.__IMPRESS_THEME_REGISTER__
    : undefined;

if (typeof register === "function") {
  register(corporateClassicTheme);
}

export {
  corporateClassicTheme,
  corporateClassicTokens,
  CORPORATE_CLASSIC_THEME_ID,
  CORPORATE_CLASSIC_CONTRACT_VERSION,
  createCorporateClassicTheme,
} from "./index";
