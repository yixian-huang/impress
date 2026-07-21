import { describe, expect, it } from "vitest";
import { getLocalizedText, normalizeConfigForLocale } from "./publicContent";

describe("getLocalizedText", () => {
  it("picks locale then falls back to zh", () => {
    expect(getLocalizedText({ zh: "中文", en: "EN" }, "en")).toBe("EN");
    expect(getLocalizedText({ zh: "中文", en: "" }, "en")).toBe("中文");
  });

  it("handles zh-only partial bags", () => {
    expect(getLocalizedText({ zh: "仅中文" }, "zh")).toBe("仅中文");
    expect(getLocalizedText({ zh: "仅中文" }, "en")).toBe("仅中文");
  });

  it("handles en-only partial bags", () => {
    expect(getLocalizedText({ en: "EN only" }, "en")).toBe("EN only");
    expect(getLocalizedText({ en: "EN only" }, "zh")).toBe("EN only");
  });
});

describe("normalizeConfigForLocale", () => {
  it("flattens full {zh,en} leaves", () => {
    const out = normalizeConfigForLocale(
      { title: { zh: "标题", en: "Title" } },
      "zh",
    );
    expect(out.title).toBe("标题");
  });

  it("flattens zh-only copyright bags (React #31 regression)", () => {
    const out = normalizeConfigForLocale(
      {
        footer: {
          copyright: { zh: "版权所有 © 2018-2025" },
          icp: "京ICP备",
        },
      },
      "zh",
    );
    expect(out.footer).toEqual({
      copyright: "版权所有 © 2018-2025",
      icp: "京ICP备",
    });
  });

  it("does not treat media refs as LocalizedText", () => {
    const out = normalizeConfigForLocale(
      {
        logo: { url: "/logo.png", alt: { zh: "标志", en: "Logo" } },
      },
      "en",
    );
    expect(out.logo).toEqual({ url: "/logo.png", alt: "Logo" });
  });
});
