import { describe, expect, it, vi } from "vitest";
import type { ArticleMetaResponse } from "@/api/ai";
import {
  applyAIMetaToForm,
  defaultSelectedKeys,
  qualityWarnings,
  resolveApplyValues,
} from "./applyAIMeta";

const sample: ArticleMetaResponse = {
  candidates: {
    zhTitles: ["标题甲", "标题乙"],
    enTitles: ["Title A", "Title B"],
  },
  suggested: {
    zhTitle: "标题甲",
    enTitle: "Title A",
    slug: "title-a",
    zhSeoTitle: "SEO 中",
    enSeoTitle: "SEO EN",
    zhMetaDescription: "中文描述",
    enMetaDescription: "English desc",
  },
  skipped: [],
};

describe("resolveApplyValues", () => {
  it("uses title candidate index", () => {
    const v0 = resolveApplyValues(sample, 0);
    expect(v0.zhTitle).toBe("标题甲");
    const v1 = resolveApplyValues(sample, 1);
    expect(v1.zhTitle).toBe("标题乙");
    expect(v1.enTitle).toBe("Title B");
  });
});

describe("defaultSelectedKeys", () => {
  it("skips slug when locked", () => {
    const values = resolveApplyValues(sample);
    const keys = defaultSelectedKeys(values, true);
    expect(keys.has("slug")).toBe(false);
    expect(keys.has("zhTitle")).toBe(true);
  });
});

describe("applyAIMetaToForm", () => {
  it("applies only selected fields", () => {
    const setters = {
      setZhTitle: vi.fn(),
      setEnTitle: vi.fn(),
      setSlug: vi.fn(),
      setZhSeoTitle: vi.fn(),
      setEnSeoTitle: vi.fn(),
      setZhMetaDescription: vi.fn(),
      setEnMetaDescription: vi.fn(),
    };
    const values = resolveApplyValues(sample);
    const n = applyAIMetaToForm(["zhTitle", "slug"], values, setters);
    expect(n).toBe(2);
    expect(setters.setZhTitle).toHaveBeenCalledWith("标题甲");
    expect(setters.setSlug).toHaveBeenCalledWith("title-a");
    expect(setters.setEnTitle).not.toHaveBeenCalled();
  });
});

describe("qualityWarnings", () => {
  it("flags long SEO title", () => {
    const long = "x".repeat(70);
    const warns = qualityWarnings({ zhSeoTitle: long });
    expect(warns.some((w) => w.includes("SEO"))).toBe(true);
  });
});
