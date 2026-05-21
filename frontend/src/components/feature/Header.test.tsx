import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

function mockCtx(localeMode: "mono-zh" | "mono-en" | "bilingual") {
  const lang = localeMode === "mono-en" ? "en" : "zh";
  vi.doMock("react-i18next", () => ({
    useTranslation: () => ({
      i18n: {
        language: lang,
        changeLanguage: vi.fn(),
      },
      t: (k: string) => k,
    }),
    Trans: ({ children }: { children: React.ReactNode }) => children,
  }));
  vi.doMock("@/contexts/GlobalConfigContext", () => ({
    useGlobalConfig: () => ({
      config: {
        siteConfig: {
          identity: { name: { zh: "测试站", en: "Test Site" }, localeMode, defaultLocale: localeMode === "mono-en" ? "en" : "zh" },
          brand: { logo: { light: "/logo.png" }, favicon: "", ogImage: "", primaryColor: "#000" },
          author: { name: "", socials: [] },
          footer: {},
          seo: {},
        },
      },
      locale: localeMode === "mono-en" ? "en" : "zh",
      loading: false,
      features: {},
      refetch: vi.fn(),
    }),
  }));
  vi.doMock("@/contexts/ThemePagesContext", () => ({
    useThemePages: () => ({ headerNavItems: [], menuNavItems: [] }),
  }));
}

describe("Header locale-mode behavior", () => {
  it.each(["mono-zh", "mono-en"] as const)("hides language switcher in %s", async (mode) => {
    vi.resetModules();
    mockCtx(mode);
    const { default: Header } = await import("./Header");
    const { queryByText } = render(<MemoryRouter><Header /></MemoryRouter>);
    expect(queryByText("English")).toBeNull();
    expect(queryByText("中文")).toBeNull();
  });

  it("shows language switcher in bilingual", async () => {
    vi.resetModules();
    mockCtx("bilingual");
    const { default: Header } = await import("./Header");
    const { getByRole } = render(<MemoryRouter><Header /></MemoryRouter>);
    // bilingual + locale=zh shows the "English" toggle text
    expect(getByRole("button", { name: /English|中文/ })).toBeTruthy();
  });
});
