import { describe, expect, it } from "vitest";
import { sanitizePastedHtml } from "./sanitizePastedHtml";

describe("sanitizePastedHtml", () => {
  it("strips mso classes and inline styles", () => {
    const html =
      '<p class="MsoNormal" style="margin:0;color:red"><span style="font-size:12pt">Hello</span></p>';
    const out = sanitizePastedHtml(html);
    expect(out).not.toMatch(/style=/i);
    expect(out).not.toMatch(/MsoNormal/i);
    expect(out).toContain("Hello");
  });

  it("removes word xml leftovers and comments", () => {
    const html =
      '<!--StartFragment--><p>A</p><o:p>&nbsp;</o:p><!--EndFragment-->';
    const out = sanitizePastedHtml(html);
    expect(out).not.toMatch(/o:p/i);
    expect(out).not.toMatch(/StartFragment/);
    expect(out).toContain("A");
  });

  it("keeps links and images with safe attrs", () => {
    const html = '<p><a href="https://ex.com" onclick="alert(1)" style="color:red">x</a>'
      + '<img src="/a.png" alt="pic" onerror="x" style="width:1px"></p>';
    const out = sanitizePastedHtml(html);
    expect(out).toContain('href="https://ex.com"');
    expect(out).toContain('src="/a.png"');
    expect(out).toContain('alt="pic"');
    expect(out).not.toMatch(/onclick|onerror|style=/i);
  });

  it("preserves plain text without tags", () => {
    expect(sanitizePastedHtml("just text")).toBe("just text");
  });
});
