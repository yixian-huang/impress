import { describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { resolveBrandingView } from "./useBranding";

vi.mock("@/contexts/GlobalConfigContext", () => ({
  useGlobalConfig: () => ({
    config: {
      siteConfig: {
        identity: {
          name: { zh: "我的博客", en: "My Blog" },
          localeMode: "bilingual",
          defaultLocale: "zh",
        },
        brand: { logo: { light: "/logo.png" }, favicon: "/fav.ico", ogImage: "", primaryColor: "#000" },
        author: { name: "Isian", socials: [{ kind: "github", url: "https://github.com/isian" }] },
        footer: { copyright: { zh: "© 2026 我的博客" }, icp: "京 ICP 备 123" },
        seo: {},
      },
    },
    loading: false,
    locale: "zh",
    features: {},
    refetch: vi.fn(),
  }),
}));

import { useBranding } from "./useBranding";

const Wrapper = ({ children }: { children: ReactNode }) => <>{children}</>;

describe("useBranding", () => {
  it("returns localized site name based on current locale", () => {
    const { result } = renderHook(() => useBranding(), { wrapper: Wrapper });
    expect(result.current.siteName).toBe("我的博客");
  });

  it("exposes logo and ICP", () => {
    const { result } = renderHook(() => useBranding(), { wrapper: Wrapper });
    expect(result.current.logo.light).toBe("/logo.png");
    expect(result.current.footer.icp).toBe("京 ICP 备 123");
  });
});

describe("resolveBrandingView legacy global config", () => {
  it("maps branding.companyName and branding.logo.url (blotting / impress shape)", () => {
    const view = resolveBrandingView(
      {
        branding: {
          companyName: "印迹安合法规咨询",
          logo: { url: "/uploads/logo.png", alt: "印迹" },
        },
        header: {
          logo: { url: "/images/logo.svg", alt: "fallback" },
        },
        footer: {
          copyright: "版权所有 © 2018-2025",
        },
      } as never,
      "zh",
    );
    expect(view.siteName).toBe("印迹安合法规咨询");
    expect(view.logo.light).toBe("/uploads/logo.png");
    expect(view.footer.copyright).toBe("版权所有 © 2018-2025");
  });

  it("falls back to header.logo when branding.logo is empty", () => {
    const view = resolveBrandingView(
      {
        branding: { companyName: { zh: "印迹咨询", en: "Blotting" } },
        header: { logo: { url: "/images/logo.svg" } },
      } as never,
      "zh",
    );
    expect(view.siteName).toBe("印迹咨询");
    expect(view.logo.light).toBe("/images/logo.svg");
  });

  it("does not use My Site default when legacy companyName is present", () => {
    const view = resolveBrandingView(
      {
        branding: {
          companyName: { zh: "印迹安合法规咨询", en: "Blotting Consultancy" },
          logo: { url: "/uploads/1771768461126636000-logo.png" },
        },
      } as never,
      "zh",
    );
    expect(view.siteName).not.toMatch(/My Site/i);
    expect(view.logo.light).toContain("/uploads/");
  });
});
