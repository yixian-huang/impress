import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";

function setMock(publicPages: Record<string, boolean>) {
  vi.doMock("@/contexts/GlobalConfigContext", () => ({
    useGlobalConfig: () => ({
      config: {},
      features: { publicPages, blog: { comments: true, rss: true } },
      locale: "zh",
      loading: false,
      refetch: vi.fn(),
    }),
  }));
}

describe("FeatureGate", () => {
  it("renders children when feature true", async () => {
    vi.resetModules();
    setMock({ about: true, home: true, blog: true, contact: true, experts: false, coreServices: false, advantages: false, cases: false });
    const { FeatureGate } = await import("./FeatureGate");
    const { getByText } = render(<FeatureGate feature="about"><span>visible</span></FeatureGate>);
    expect(getByText("visible")).toBeTruthy();
  });

  it("renders NotFound when feature false", async () => {
    vi.resetModules();
    setMock({ about: false, home: true, blog: true, contact: true, experts: false, coreServices: false, advantages: false, cases: false });
    const { FeatureGate } = await import("./FeatureGate");
    const { container } = render(<FeatureGate feature="about"><span data-testid="should-not-render">x</span></FeatureGate>);
    expect(container.querySelector('[data-testid="should-not-render"]')).toBeNull();
  });

  it("old-deploy compat: missing features record → renders as enabled", async () => {
    vi.resetModules();
    vi.doMock("@/contexts/GlobalConfigContext", () => ({
      useGlobalConfig: () => ({
        config: {},
        features: {},
        locale: "zh",
        loading: false,
        refetch: vi.fn(),
      }),
    }));
    const { FeatureGate } = await import("./FeatureGate");
    const { getByText } = render(<FeatureGate feature="about"><span>visible</span></FeatureGate>);
    expect(getByText("visible")).toBeTruthy();
  });
});
