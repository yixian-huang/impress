import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";

function setMock(blog: { comments: boolean; rss: boolean }, features?: Record<string, unknown>) {
  vi.doMock("@/contexts/GlobalConfigContext", () => ({
    useGlobalConfig: () => ({
      config: {},
      features: features ?? { blog },
      locale: "zh",
      loading: false,
      refetch: vi.fn(),
    }),
  }));
}

describe("BlogFeatureGate", () => {
  it("renders children when feature enabled", async () => {
    vi.resetModules();
    setMock({ comments: true, rss: true });
    const { BlogFeatureGate } = await import("./BlogFeatureGate");
    const { getByText } = render(
      <BlogFeatureGate feature="comments">
        <span>visible</span>
      </BlogFeatureGate>,
    );
    expect(getByText("visible")).toBeTruthy();
  });

  it("hides children when rss disabled", async () => {
    vi.resetModules();
    setMock({ comments: true, rss: false });
    const { BlogFeatureGate } = await import("./BlogFeatureGate");
    const { container } = render(
      <BlogFeatureGate feature="rss">
        <span data-testid="hidden">x</span>
      </BlogFeatureGate>,
    );
    expect(container.querySelector('[data-testid="hidden"]')).toBeNull();
  });
});
