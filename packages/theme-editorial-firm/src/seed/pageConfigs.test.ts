import { describe, expect, it } from "vitest";
import {
  EDITORIAL_FIRM_SEED_PAGE_KEYS,
  editorialFirmPageConfigs,
} from "./pageConfigs";

describe("editorialFirmPageConfigs", () => {
  it("covers the four editorial-firm pages", () => {
    for (const key of EDITORIAL_FIRM_SEED_PAGE_KEYS) {
      expect(editorialFirmPageConfigs[key]).toBeDefined();
      expect(Array.isArray(editorialFirmPageConfigs[key].sections)).toBe(true);
      expect(editorialFirmPageConfigs[key].sections.length).toBeGreaterThan(0);
    }
  });

  it("home has at least 4 sections matching design composition", () => {
    const { sections } = editorialFirmPageConfigs.home;
    expect(sections.length).toBeGreaterThanOrEqual(4);
    const types = sections.map((s) => s.type);
    expect(types).toContain("ef-hero-editorial");
    expect(types).toContain("ef-feature-split");
    expect(types).toContain("ef-service-index");
    expect(types).toContain("ef-pull-quote");
    expect(types).toContain("ef-cta-band");
  });

  it("every section type starts with ef-", () => {
    for (const key of EDITORIAL_FIRM_SEED_PAGE_KEYS) {
      for (const section of editorialFirmPageConfigs[key].sections) {
        expect(section.type.startsWith("ef-")).toBe(true);
        expect(section.id).toBeTruthy();
        expect(section.id.startsWith("ef-")).toBe(true);
        expect(section.data).toBeTypeOf("object");
      }
    }
  });

  it("service-index on home has three items linking to /services", () => {
    const svc = editorialFirmPageConfigs.home.sections.find(
      (s) => s.type === "ef-service-index",
    );
    expect(svc).toBeDefined();
    const items = svc!.data.items as Array<{ href?: string }>;
    expect(items).toHaveLength(3);
    for (const item of items) {
      expect(item.href).toBe("/services");
    }
  });

  it("contact page is contact-split only (or quote-optional)", () => {
    const types = editorialFirmPageConfigs.contact.sections.map((s) => s.type);
    expect(types).toContain("ef-contact-split");
    for (const t of types) {
      expect(["ef-contact-split", "ef-pull-quote"].includes(t)).toBe(true);
    }
  });

  it("uses no 印迹 / Blotting brand strings", () => {
    const blob = JSON.stringify(editorialFirmPageConfigs);
    expect(blob).not.toMatch(/印迹|Blotting|blotting/i);
  });
});
