/**
 * After `vite build`, inject Go html/template SEO tags into out/index.html
 * for backend server-side rendering. Dev uses static placeholders in index.html.
 */
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const outPath = join(root, "out", "index.html");
const tmplPath = join(root, "index.seo.tmpl");

const START = "<!-- inkless-seo-meta-start -->";
const END = "<!-- inkless-seo-meta-end -->";

const html = readFileSync(outPath, "utf8");
const tmpl = readFileSync(tmplPath, "utf8");

const startIdx = html.indexOf(START);
const endIdx = html.indexOf(END);
if (startIdx === -1 || endIdx === -1 || endIdx <= startIdx) {
  console.error("inject-seo-template: markers not found in out/index.html");
  process.exit(1);
}

const before = html.slice(0, startIdx + START.length);
const after = html.slice(endIdx);
const next = `${before}\n${tmpl}\n${after}`;

const withLocale = next.replace(
  /<html\s+lang="zh"/,
  '<html lang="{{.Locale}}"',
);

writeFileSync(outPath, withLocale);
console.log("inject-seo-template: injected Go SEO template into out/index.html");
