import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";

describe("Footer", () => {
  it("auto-generates copyright when not configured", async () => {
    vi.resetModules();
    vi.doMock("@/contexts/GlobalConfigContext", () => ({
      useGlobalConfig: () => ({
        config: {
          siteConfig: {
            identity: { name: { zh: "我的站" }, localeMode: "mono-zh", defaultLocale: "zh" },
            brand: { logo: { light: "" }, favicon: "", ogImage: "", primaryColor: "#000" },
            author: { name: "", socials: [] },
            footer: {},
            seo: {},
          },
        },
        locale: "zh",
        loading: false,
        features: {},
        refetch: vi.fn(),
      }),
    }));
    const { default: Footer } = await import("./Footer");
    const { container } = render(<Footer />);
    const year = String(new Date().getFullYear());
    expect(container.textContent).toContain(year);
    expect(container.textContent).toContain("我的站");
  });

  it("hides ICP block when icp is empty", async () => {
    vi.resetModules();
    vi.doMock("@/contexts/GlobalConfigContext", () => ({
      useGlobalConfig: () => ({
        config: {
          siteConfig: {
            identity: { name: { zh: "x" }, localeMode: "mono-zh", defaultLocale: "zh" },
            brand: { logo: { light: "" }, favicon: "", ogImage: "", primaryColor: "#000" },
            author: { name: "", socials: [] },
            footer: { icp: "" },
            seo: {},
          },
        },
        locale: "zh",
        loading: false,
        features: {},
        refetch: vi.fn(),
      }),
    }));
    const { default: Footer } = await import("./Footer");
    const { container } = render(<Footer />);
    expect(container.textContent).not.toMatch(/ICP/);
  });
});
