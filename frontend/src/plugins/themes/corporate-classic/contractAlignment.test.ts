import { describe, expect, it } from "vitest";
import {
  CORPORATE_CLASSIC_CONTRACT_VERSION,
  CORPORATE_CLASSIC_THEME_ID,
  createCorporateClassicTheme,
} from "@inkless/theme-corporate-classic";
import { THEME_CONTRACT_VERSION } from "@/theme-host/contract";
import { BUILTIN_THEME_IDS } from "@/plugins/builtinThemes";
import { corporateClassicTheme } from "./index";

describe("corporate-classic contract alignment", () => {
  it("theme id matches builtin constant", () => {
    expect(CORPORATE_CLASSIC_THEME_ID).toBe(BUILTIN_THEME_IDS.CORPORATE_CLASSIC);
    expect(corporateClassicTheme.manifest.id).toBe(BUILTIN_THEME_IDS.CORPORATE_CLASSIC);
  });

  it("targets host contract v1", () => {
    expect(CORPORATE_CLASSIC_CONTRACT_VERSION).toBe(THEME_CONTRACT_VERSION);
    expect(corporateClassicTheme.contractVersion).toBe(THEME_CONTRACT_VERSION);
  });

  it("host loaders attach all seven corporate pages", () => {
    expect(corporateClassicTheme.pages).toHaveLength(7);
    for (const page of corporateClassicTheme.pages) {
      expect(page.lazyComponent).toEqual(expect.any(Function));
    }
  });

  it("shell without loaders still has page metadata", () => {
    const shell = createCorporateClassicTheme();
    expect(shell.pages.map((p) => p.contentKey)).toEqual([
      "home",
      "about",
      "advantages",
      "core-services",
      "cases",
      "experts",
      "contact",
    ]);
  });
});
