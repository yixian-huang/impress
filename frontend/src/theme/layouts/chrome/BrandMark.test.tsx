import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import BrandMark from "./BrandMark";

vi.mock("@/hooks/useBranding", () => ({
  useBranding: () => ({
    siteName: "Test Site",
    author: { name: "Author", avatar: "/av.png", bio: "", socials: [] },
    logo: { light: "" },
  }),
}));

describe("BrandMark", () => {
  it("renders nothing for none mode", () => {
    const { container } = render(
      <MemoryRouter>
        <BrandMark brandMode="none" />
      </MemoryRouter>,
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders author name for text mode", () => {
    render(
      <MemoryRouter>
        <BrandMark brandMode="text" />
      </MemoryRouter>,
    );
    expect(screen.getByText("Author")).toBeTruthy();
  });
});
