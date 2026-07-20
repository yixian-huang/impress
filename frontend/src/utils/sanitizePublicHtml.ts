/**
 * Sanitize CMS article HTML for untrusted public render.
 * Strips scripts, event handlers, and javascript: URLs while keeping
 * TipTap/blog content structure (including mermaid source nodes).
 */

const ALLOWED_TAGS = new Set([
  "p",
  "br",
  "div",
  "span",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "strong",
  "b",
  "em",
  "i",
  "u",
  "s",
  "del",
  "strike",
  "sub",
  "sup",
  "a",
  "img",
  "ul",
  "ol",
  "li",
  "blockquote",
  "pre",
  "code",
  "table",
  "thead",
  "tbody",
  "tr",
  "th",
  "td",
  "hr",
  "mark",
  "figure",
  "figcaption",
  "section",
  "article",
  "header",
  "footer",
  "nav",
  "aside",
  "audio",
  "video",
  "source",
  "picture",
]);

const ATTR_ALLOW: Record<string, Set<string>> = {
  a: new Set(["href", "title", "target", "rel"]),
  img: new Set(["src", "alt", "title", "width", "height", "loading"]),
  td: new Set(["colspan", "rowspan"]),
  th: new Set(["colspan", "rowspan"]),
  code: new Set(["class"]),
  pre: new Set(["class"]),
  div: new Set(["class", "data-type", "data-mermaid-source"]),
  span: new Set(["class"]),
  audio: new Set(["src", "controls", "preload"]),
  video: new Set(["src", "controls", "preload", "poster", "width", "height"]),
  source: new Set(["src", "type"]),
};

const GLOBAL_CLASS_ALLOW = new Set([
  "mermaid",
  "language-mermaid",
  "lang-mermaid",
  "code-block-wrap",
  "hljs",
]);

function isDangerousURL(value: string): boolean {
  const v = value.trim().toLowerCase();
  return (
    v.startsWith("javascript:") ||
    v.startsWith("vbscript:") ||
    v.startsWith("data:text/html")
  );
}

function unwrapElement(el: Element) {
  const parent = el.parentNode;
  if (!parent) return;
  while (el.firstChild) parent.insertBefore(el.firstChild, el);
  parent.removeChild(el);
}

function stripWithRegex(html: string): string {
  return html
    .replace(/<script[\s\S]*?<\/script>/gi, "")
    .replace(/<style[\s\S]*?<\/style>/gi, "")
    .replace(/on\w+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, "")
    .replace(/\s(href|src)\s*=\s*("|')\s*javascript:[\s\S]*?\2/gi, ' $1="#"')
    .replace(/<\/?(iframe|object|embed|form|input|button|textarea|select|meta|link|base)[^>]*>/gi, "");
}

/**
 * Sanitize HTML for public article rendering.
 * Uses DOMParser when available (browser / happy-dom); regex fallback otherwise.
 */
export function sanitizePublicHtml(html: string): string {
  if (!html || typeof html !== "string") return "";
  if (!/<[a-zA-Z!/?]/.test(html)) return html;

  const pre = stripWithRegex(html);

  if (typeof DOMParser === "undefined") {
    return pre;
  }

  try {
    const doc = new DOMParser().parseFromString(
      `<div id="__public_html_root__">${pre}</div>`,
      "text/html",
    );
    const root = doc.getElementById("__public_html_root__") || doc.body;
    if (!root) return pre;

    const walker = doc.createTreeWalker(root, NodeFilter.SHOW_COMMENT);
    const comments: Comment[] = [];
    while (walker.nextNode()) comments.push(walker.currentNode as Comment);
    for (const c of comments) c.remove();

    for (const el of Array.from(root.querySelectorAll("*"))) {
      const tag = el.tagName.toLowerCase();

      if (tag === "script" || tag === "style" || tag === "iframe" || tag === "object" || tag === "embed") {
        el.remove();
        continue;
      }

      if (!ALLOWED_TAGS.has(tag)) {
        unwrapElement(el);
        continue;
      }

      const allowed = ATTR_ALLOW[tag] || new Set<string>();
      for (const attr of Array.from(el.attributes)) {
        const name = attr.name.toLowerCase();
        if (name.startsWith("on") || name === "style" || name === "srcdoc") {
          el.removeAttribute(attr.name);
          continue;
        }
        if (name === "class") {
          const kept = attr.value
            .split(/\s+/)
            .filter((c) => c && (GLOBAL_CLASS_ALLOW.has(c) || c.startsWith("language-") || c.startsWith("hljs")));
          if (kept.length && (allowed.has("class") || tag === "div" || tag === "pre" || tag === "code")) {
            el.setAttribute("class", kept.join(" "));
          } else {
            el.removeAttribute("class");
          }
          continue;
        }
        if (name.startsWith("data-")) {
          if (allowed.has(name)) continue;
          el.removeAttribute(attr.name);
          continue;
        }
        if (!allowed.has(name)) {
          el.removeAttribute(attr.name);
          continue;
        }
        if ((name === "href" || name === "src") && isDangerousURL(attr.value)) {
          el.removeAttribute(attr.name);
        }
      }
    }

    return root.innerHTML;
  } catch {
    return pre;
  }
}
