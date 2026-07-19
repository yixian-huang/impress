import type { CustomFontRef } from "./types";

const loadedUrls = new Set<string>();

/** Inject @font-face rules for uploaded custom fonts (woff2 preferred). */
export function loadCustomFonts(fonts: CustomFontRef[]): void {
  if (typeof document === "undefined" || fonts.length === 0) return;

  const sheetId = "inkless-custom-fonts";
  let sheet = document.getElementById(sheetId) as HTMLStyleElement | null;
  if (!sheet) {
    sheet = document.createElement("style");
    sheet.id = sheetId;
    document.head.appendChild(sheet);
  }

  const rules: string[] = [];
  for (const font of fonts) {
    if (!font.url || !font.family || loadedUrls.has(font.url)) continue;
    loadedUrls.add(font.url);

    const format = font.url.endsWith(".woff2")
      ? "woff2"
      : font.url.endsWith(".woff")
        ? "woff"
        : "opentype";

    rules.push(`
@font-face {
  font-family: ${JSON.stringify(font.family)};
  src: url(${JSON.stringify(font.url)}) format(${JSON.stringify(format)});
  font-weight: ${font.weight ?? 400};
  font-style: ${font.style ?? "normal"};
  font-display: swap;
}`);
  }

  if (rules.length > 0) {
    sheet.textContent = (sheet.textContent ?? "") + rules.join("\n");
  }
}
