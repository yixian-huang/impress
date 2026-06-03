import { describe, it, expect, vi } from "vitest";
import { renderHook } from "@testing-library/react";
import { useHeaderSettings } from "./useHeaderSettings";

vi.mock("@/contexts/GlobalConfigContext", () => ({
  useGlobalConfig: () => ({
    config: {
      siteConfig: {
        header: { brandMode: "avatar" as const },
      },
    },
  }),
}));

vi.mock("@/plugins/hooks", () => ({
  useThemeSettings: () => ({
    "header.brandMode": "text",
    "header.showRssLink": true,
  }),
}));

describe("useHeaderSettings", () => {
  it("site config overrides theme defaults", () => {
    const { result } = renderHook(() => useHeaderSettings());
    expect(result.current.brandMode).toBe("avatar");
    expect(result.current.showRssLink).toBe(true);
  });
});
