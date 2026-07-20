import { describe, expect, it } from "vitest";
import { sanitizePublicHtml } from "./sanitizePublicHtml";

describe("sanitizePublicHtml", () => {
  it("strips script tags and event handlers from public article HTML", () => {
    const dirty =
      `<p>Hello</p><script>alert(1)</script><img src="/x.png" onerror="alert(2)" alt="ok">` +
      `<a href="javascript:alert(3)">bad</a><p onclick="evil()">click</p>`;

    const clean = sanitizePublicHtml(dirty);

    expect(clean).toContain("Hello");
    expect(clean.toLowerCase()).not.toContain("<script");
    expect(clean.toLowerCase()).not.toContain("onerror");
    expect(clean.toLowerCase()).not.toContain("onclick");
    expect(clean.toLowerCase()).not.toContain("javascript:");
    expect(clean).toMatch(/img[^>]*src=["']\/x\.png["']/i);
  });

  it("removes iframe and form injection", () => {
    const dirty = `<p>Safe</p><iframe src="https://evil.example"></iframe><form action="/x"><input></form>`;
    const clean = sanitizePublicHtml(dirty);
    expect(clean).toContain("Safe");
    expect(clean.toLowerCase()).not.toContain("iframe");
    expect(clean.toLowerCase()).not.toContain("<form");
    expect(clean.toLowerCase()).not.toContain("<input");
  });

  it("keeps mermaid source nodes for client render", () => {
    const html =
      `<div class="mermaid" data-type="mermaid" data-mermaid-source="graph TD; A-->B">graph TD; A-->B</div>` +
      `<pre><code class="language-mermaid">sequenceDiagram</code></pre>`;
    const clean = sanitizePublicHtml(html);
    expect(clean).toContain("mermaid");
    expect(clean).toContain("data-type");
    expect(clean).toContain("graph TD");
    expect(clean).toContain("language-mermaid");
  });

  it("returns empty string for empty input", () => {
    expect(sanitizePublicHtml("")).toBe("");
  });
});
