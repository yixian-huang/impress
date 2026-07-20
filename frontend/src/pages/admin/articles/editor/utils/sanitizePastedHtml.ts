/**
 * Clean HTML pasted from Word / Google Docs / web pages so TipTap
 * does not inherit MSO styles, font tags, or noisy spans.
 */

const ALLOWED_TAGS = new Set([
  "p", "br", "div", "span",
  "h1", "h2", "h3", "h4", "h5", "h6",
  "strong", "b", "em", "i", "u", "s", "del", "strike", "sub", "sup",
  "a", "img",
  "ul", "ol", "li",
  "blockquote", "pre", "code",
  "table", "thead", "tbody", "tr", "th", "td",
  "hr",
  "mark",
]);

/** Attributes allowed on specific tags */
const ATTR_ALLOW: Record<string, Set<string>> = {
  a: new Set(["href", "title", "target", "rel"]),
  img: new Set(["src", "alt", "title", "width", "height"]),
  td: new Set(["colspan", "rowspan"]),
  th: new Set(["colspan", "rowspan"]),
  code: new Set(["class"]),
  pre: new Set(["class"]),
};

function unwrapElement(el: Element) {
  const parent = el.parentNode;
  if (!parent) return;
  while (el.firstChild) parent.insertBefore(el.firstChild, el);
  parent.removeChild(el);
}

function isMsoOrJunkClass(className: string): boolean {
  return /mso|Mso|apple-|WordSection|normal0/i.test(className);
}

/**
 * Sanitize pasted HTML string. Safe to call in browser; no-ops on empty.
 */
export function sanitizePastedHtml(html: string): string {
  if (!html || typeof html !== "string") return html || "";
  // Quick path: plain text fragments without tags
  if (!/<[a-zA-Z!/?]/.test(html)) return html;

  const cleaned = html
    // XML / Word namespaces
    .replace(/<\?xml[\s\S]*?\?>/gi, "")
    .replace(/<o:p[^>]*>[\s\S]*?<\/o:p>/gi, "")
    .replace(/<\/?o:[^>]*>/gi, "")
    .replace(/<\/?w:[^>]*>/gi, "")
    .replace(/<\/?m:[^>]*>/gi, "")
    // Conditional comments
    .replace(/<!--\[if[\s\S]*?<!\[endif\]-->/gi, "")
    .replace(/<!--[\s\S]*?-->/g, "")
    // Style / script blocks
    .replace(/<style[\s\S]*?<\/style>/gi, "")
    .replace(/<script[\s\S]*?<\/script>/gi, "")
    .replace(/<meta[^>]*>/gi, "")
    .replace(/<link[^>]*>/gi, "");

  if (typeof DOMParser === "undefined") {
    // Node / test fallback without full DOM — strip style= and class=
    return cleaned
      .replace(/\sstyle\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, "")
      .replace(/\sclass\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, "");
  }

  try {
    const doc = new DOMParser().parseFromString(
      `<div id="__paste_root__">${cleaned}</div>`,
      "text/html",
    );
    const root = doc.getElementById("__paste_root__") || doc.body;
    if (!root) return cleaned;

    // Remove comments
    const walker = doc.createTreeWalker(root, NodeFilter.SHOW_COMMENT);
    const comments: Comment[] = [];
    while (walker.nextNode()) comments.push(walker.currentNode as Comment);
    for (const c of comments) c.remove();

    const all = Array.from(root.querySelectorAll("*"));
    for (const el of all) {
      const tag = el.tagName.toLowerCase();

      // Drop unknown / junk tags but keep children
      if (!ALLOWED_TAGS.has(tag) || tag === "font") {
        unwrapElement(el);
        continue;
      }

      // Strip event handlers and most attributes
      const allowed = ATTR_ALLOW[tag] || new Set<string>();
      for (const attr of Array.from(el.attributes)) {
        const name = attr.name.toLowerCase();
        if (name.startsWith("on") || name === "style" || name.startsWith("data-")) {
          el.removeAttribute(attr.name);
          continue;
        }
        if (name === "class") {
          if (isMsoOrJunkClass(attr.value) || !allowed.has("class")) {
            el.removeAttribute("class");
          }
          continue;
        }
        if (!allowed.has(name)) {
          el.removeAttribute(attr.name);
        }
      }

      // Normalize presentational tags
      if (tag === "b") {
        const strong = doc.createElement("strong");
        while (el.firstChild) strong.appendChild(el.firstChild);
        el.replaceWith(strong);
      } else if (tag === "i") {
        const em = doc.createElement("em");
        while (el.firstChild) em.appendChild(el.firstChild);
        el.replaceWith(em);
      }

      // Empty spans → unwrap
      if (el.tagName.toLowerCase() === "span" && el.attributes.length === 0) {
        unwrapElement(el);
      }
    }

    return root.innerHTML;
  } catch {
    return cleaned
      .replace(/\sstyle\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, "")
      .replace(/\sclass\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)/gi, "");
  }
}
