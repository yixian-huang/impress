/**
 * Regression tests for migrated config-driven public pages
 * Tests cover all pages migrated in FE-106 and FE-107
 */
import { describe, it, expect, vi, beforeEach } from "vitest";
import { screen, waitFor } from "@testing-library/react";
import { render } from "@/test/test-utils";
import {
  mockHomeConfig,
  mockAboutConfig,
  mockAdvantagesConfig,
  mockCoreServicesConfig,
  mockCasesConfig,
  mockExpertsConfig,
  mockContactConfig,
} from "@/test/mock-data";

// Import page components
import HomePage from "@/pages/home/page";
import AboutPage from "@/pages/about/page";
import AdvantagesPage from "@/pages/advantages/page";
import CoreServicesPage from "@/pages/core-services/page";
import CasesPage from "@/pages/cases/page";
import ExpertsPage from "@/pages/experts/page";
import ContactPage from "@/pages/contact/page";

// Mock the usePublicContent hook
vi.mock("@/hooks/usePublicContent", () => ({
  usePublicContent: vi.fn(),
}));

import { usePublicContent, type UsePublicContentResult } from "@/hooks/usePublicContent";

// Helper to create properly typed mock return values
const createMockResult = (
  config: Record<string, unknown> | null,
  loading = false,
  error: { message: string; code: string } | null = null
): UsePublicContentResult => ({
  loading,
  error,
  config,
  response: config
    ? {
        pageKey: "home" as const,
        locale: "zh" as const,
        version: 1,
        config: config as Record<string, unknown>,
      }
    : null,
  refetch: vi.fn(),
});

describe("Public Pages Regression Suite - Chinese Locale", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Home Page - zh locale", () => {
    it("should render hero section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockHomeConfig));

      render(<HomePage />);

      await waitFor(() => {
        expect(screen.getByText(mockHomeConfig.hero.title)).toBeInTheDocument();
      });

      expect(screen.getByText(mockHomeConfig.hero.subtitle)).toBeInTheDocument();
    });

    it("should render about section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockHomeConfig));

      render(<HomePage />);

      await waitFor(() => {
        expect(screen.getByText(mockHomeConfig.about.title)).toBeInTheDocument();
      });

      expect(screen.getByText(mockHomeConfig.about.descriptions[0])).toBeInTheDocument();
    });

    it("should render advantages section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockHomeConfig));

      render(<HomePage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockHomeConfig.advantages.title)
        ).toBeInTheDocument();
      });

      expect(
        screen.getAllByText(mockHomeConfig.advantages.cards[0].title).length
      ).toBeGreaterThan(0);
    });

    it("should render core services section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockHomeConfig));

      render(<HomePage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockHomeConfig.coreServices.title)
        ).toBeInTheDocument();
      });

      expect(
        screen.getByText(mockHomeConfig.coreServices.items[0].title)
      ).toBeInTheDocument();
    });

    it("should show loading state", () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(null, true));

      render(<HomePage />);

      expect(screen.getByText("Loading...")).toBeInTheDocument();
    });

    it("should show error state", () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(null, false, { message: "Network error", code: "NETWORK_ERROR" }));

      render(<HomePage />);

      expect(
        screen.getByText("Failed to load page content")
      ).toBeInTheDocument();
    });
  });

  describe("About Page - zh locale", () => {
    it("should render hero section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockAboutConfig));

      render(<AboutPage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockAboutConfig.hero.title)
        ).toBeInTheDocument();
      });
    });

    it("should render company profile from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockAboutConfig));

      render(<AboutPage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockAboutConfig.companyProfile.title!)
        ).toBeInTheDocument();
      });
    });

    it("should show loading state", () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(null, true));

      render(<AboutPage />);

      expect(screen.getByText("Loading...")).toBeInTheDocument();
    });
  });

  describe("Advantages Page - zh locale", () => {
    it("should render hero section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockAdvantagesConfig));

      render(<AdvantagesPage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockAdvantagesConfig.hero.title!)
        ).toBeInTheDocument();
      });
    });

    it("should render advantage blocks from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockAdvantagesConfig));

      render(<AdvantagesPage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockAdvantagesConfig.blocks![0].title!)
        ).toBeInTheDocument();
      });
    });
  });

  describe("Core Services Page - zh locale", () => {
    it("should render hero section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockCoreServicesConfig));

      render(<CoreServicesPage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockCoreServicesConfig.hero.title)
        ).toBeInTheDocument();
      });
    });

    it("should render all services from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockCoreServicesConfig));

      render(<CoreServicesPage />);

      await waitFor(() => {
        mockCoreServicesConfig.services.forEach((service) => {
          expect(screen.getByText(service.title)).toBeInTheDocument();
        });
      });
    });
  });

  describe("Cases Page - zh locale", () => {
    it("should render hero section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockCasesConfig));

      render(<CasesPage />);

      await waitFor(() => {
        expect(screen.getByText(mockCasesConfig.hero.title!)).toBeInTheDocument();
      });
    });

    it("should render case categories from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockCasesConfig));

      render(<CasesPage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockCasesConfig.cases![0].title!)
        ).toBeInTheDocument();
      });
    });
  });

  describe("Experts Page - zh locale", () => {
    it("should render hero section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockExpertsConfig));

      render(<ExpertsPage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockExpertsConfig.hero.title!)
        ).toBeInTheDocument();
      });
    });

    it("should render experts from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockExpertsConfig));

      render(<ExpertsPage />);

      await waitFor(() => {
        const expertNames = screen.getAllByText(mockExpertsConfig.experts![0].name!);
        expect(expertNames.length).toBeGreaterThan(0);
      });
    });
  });

  describe("Contact Page - zh locale", () => {
    it("should render hero section from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockContactConfig));

      render(<ContactPage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockContactConfig.hero.title)
        ).toBeInTheDocument();
      });
    });

    it("should render contact information from config", async () => {
      vi.mocked(usePublicContent).mockReturnValue(createMockResult(mockContactConfig));

      render(<ContactPage />);

      await waitFor(() => {
        expect(
          screen.getByText(mockContactConfig.contactInfo.phone!)
        ).toBeInTheDocument();
      });
    });
  });
});

describe("Public Pages Regression Suite - English Locale", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should render Home page in English locale", async () => {
    const enHomeConfig = {
      ...mockHomeConfig,
      hero: { ...mockHomeConfig.hero, title: "English Hero Title" },
    };

    vi.mocked(usePublicContent).mockReturnValue(createMockResult(enHomeConfig));

    render(<HomePage />);

    await waitFor(() => {
      expect(screen.getByText("English Hero Title")).toBeInTheDocument();
    });
  });

  it("should render About page in English locale", async () => {
    const enAboutConfig = {
      ...mockAboutConfig,
      hero: { ...mockAboutConfig.hero, title: "English About Hero" },
    };

    vi.mocked(usePublicContent).mockReturnValue(createMockResult(enAboutConfig));

    render(<AboutPage />);

    await waitFor(() => {
      expect(screen.getByText("English About Hero")).toBeInTheDocument();
    });
  });
});

describe("Graceful Degradation - Missing Optional Fields", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should handle missing hero subtitle gracefully", async () => {
    const configWithoutSubtitle = {
      ...mockHomeConfig,
      hero: { title: mockHomeConfig.hero.title },
    };

    vi.mocked(usePublicContent).mockReturnValue(createMockResult(configWithoutSubtitle));

    render(<HomePage />);

    await waitFor(() => {
      expect(
        screen.getByText(mockHomeConfig.hero.title)
      ).toBeInTheDocument();
    });

    expect(screen.queryByText(mockHomeConfig.hero.subtitle)).not.toBeInTheDocument();
  });

  it("should handle empty advantages cards array", async () => {
    const configWithoutCards = {
      ...mockHomeConfig,
      advantages: { ...mockHomeConfig.advantages, cards: [] },
    };

    vi.mocked(usePublicContent).mockReturnValue(createMockResult(configWithoutCards));

    render(<HomePage />);

    await waitFor(() => {
      expect(
        screen.getByText(mockHomeConfig.advantages.title)
      ).toBeInTheDocument();
    });

    expect(
      screen.queryByText(mockHomeConfig.advantages.cards[0].title)
    ).not.toBeInTheDocument();
  });

  it("should handle completely missing section", async () => {
    const configWithoutAbout = {
      hero: mockHomeConfig.hero,
      advantages: mockHomeConfig.advantages,
      coreServices: mockHomeConfig.coreServices,
    };

    vi.mocked(usePublicContent).mockReturnValue(createMockResult(configWithoutAbout));

    render(<HomePage />);

    await waitFor(() => {
      expect(screen.getByText(mockHomeConfig.hero.title)).toBeInTheDocument();
    });

    expect(
      screen.queryByText(mockHomeConfig.about.title)
    ).not.toBeInTheDocument();
  });
});
