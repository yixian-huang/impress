import { useLayoutEffect, useMemo, useRef, type RefObject } from "react";
import { useTranslation } from "react-i18next";
import { sanitizePublicHtml } from "@/utils/sanitizePublicHtml";

interface ArticlePostBodyProps {
  html: string;
  contentRef?: RefObject<HTMLElement | null>;
  onClick?: (e: React.MouseEvent) => void;
}

let mermaidReady: Promise<typeof import("mermaid").default> | null = null;
let mermaidSeq = 0;

function loadMermaid() {
  if (!mermaidReady) {
    mermaidReady = import("mermaid").then((mod) => {
      const mermaid = mod.default;
      mermaid.initialize({
        startOnLoad: false,
        // Strict: no click handlers / HTML labels from untrusted diagram text.
        securityLevel: "strict",
        theme: "neutral",
        fontFamily: "inherit",
      });
      return mermaid;
    });
  }
  return mermaidReady;
}

/** Collect mermaid source nodes, including code fences that were not rewritten. */
function collectMermaidNodes(root: HTMLElement): HTMLElement[] {
  const out: HTMLElement[] = [];

  root.querySelectorAll<HTMLElement>(".mermaid, [data-type='mermaid']").forEach((el) => {
    out.push(el);
  });

  root.querySelectorAll<HTMLElement>("pre > code.language-mermaid, pre > code.lang-mermaid").forEach((code) => {
    const pre = code.parentElement as HTMLElement | null;
    if (!pre || pre.closest(".mermaid")) return;
    const div = document.createElement("div");
    div.className = "mermaid";
    div.setAttribute("data-type", "mermaid");
    const src = (code.textContent ?? "").trim();
    div.textContent = src;
    if (src) div.setAttribute("data-mermaid-source", src);
    // If pre was wrapped for copy UI, replace the wrapper
    const wrap = pre.parentElement?.classList.contains("code-block-wrap") ? pre.parentElement : pre;
    wrap.replaceWith(div);
    out.push(div);
  });

  return out;
}

/**
 * Recover mermaid definition text.
 * Never treat rendered SVG markup as source (common after scroll re-entry re-runs).
 */
function sourceOf(node: HTMLElement): string {
  const attr =
    node.getAttribute("data-mermaid-source") ||
    node.getAttribute("data-source") ||
    "";
  if (attr.trim()) return attr.trim();

  // Already has SVG but lost attributes — cannot recover definition from SVG text.
  if (node.querySelector("svg")) return "";

  const text = (node.textContent || "").trim();
  if (!text) return "";
  if (text.startsWith("<svg") || text.includes("<svg") || text.startsWith("<?xml")) return "";
  return text;
}

function isMermaidRendered(node: HTMLElement): boolean {
  return Boolean(node.querySelector("svg")) && node.getAttribute("data-processed") === "true";
}

async function renderMermaidNodes(nodes: HTMLElement[], opts?: { force?: boolean }) {
  if (nodes.length === 0) return;
  const mermaid = await loadMermaid();

  for (const node of nodes) {
    if (!opts?.force && isMermaidRendered(node)) {
      continue;
    }

    const source = sourceOf(node);
    if (!source) continue;

    // Persist source before any DOM wipe so later re-renders can recover.
    node.setAttribute("data-mermaid-source", source);
    node.setAttribute("data-type", "mermaid");
    node.classList.add("mermaid");
    node.removeAttribute("data-processed");
    node.innerHTML = source;

    try {
      await mermaid.run({ nodes: [node], suppressErrors: false });
      if (node.querySelector("svg")) {
        node.setAttribute("data-processed", "true");
        // Ensure attribute survives mermaid DOM rewrites.
        node.setAttribute("data-mermaid-source", source);
      }
    } catch {
      try {
        const id = `mermaid-pub-${++mermaidSeq}`;
        const { svg } = await mermaid.render(id, source);
        node.innerHTML = svg;
        node.setAttribute("data-processed", "true");
        node.setAttribute("data-mermaid-source", source);
      } catch (err) {
        console.warn("Mermaid render failed:", err);
        node.innerHTML = `<pre class="mermaid-error">${source.replace(/</g, "&lt;")}</pre>`;
        node.setAttribute("data-mermaid-source", source);
      }
    }
  }
}

/**
 * When diagrams leave the viewport and return blank (or React rewrote HTML),
 * re-render only nodes that are visible but missing SVG.
 */
function observeMermaidVisibility(
  root: HTMLElement,
  onNeedRender: (nodes: HTMLElement[]) => void,
): () => void {
  if (typeof IntersectionObserver === "undefined") return () => {};

  const observer = new IntersectionObserver(
    (entries) => {
      const need: HTMLElement[] = [];
      for (const entry of entries) {
        if (!entry.isIntersecting) continue;
        const el = entry.target as HTMLElement;
        if (!isMermaidRendered(el) && sourceOf(el)) {
          need.push(el);
        }
      }
      if (need.length) onNeedRender(need);
    },
    { root: null, rootMargin: "80px", threshold: 0.01 },
  );

  root.querySelectorAll<HTMLElement>(".mermaid, [data-type='mermaid']").forEach((el) => {
    observer.observe(el);
  });

  return () => observer.disconnect();
}

function langFromCode(code: HTMLElement | null): string {
  if (!code) return "";
  const cls = code.getAttribute("class") || "";
  const m = cls.match(/(?:language|lang)-([a-z0-9_+-]+)/i);
  return m?.[1] ?? "";
}

/**
 * Enhance each <pre><code> with a one-click copy button.
 * Skips mermaid fences (handled separately).
 */
function enhanceCodeBlocks(
  root: HTMLElement,
  labels: { copy: string; copied: string },
) {
  root.querySelectorAll<HTMLPreElement>("pre").forEach((pre) => {
    if (pre.closest(".mermaid") || pre.classList.contains("mermaid-error")) return;
    const code = pre.querySelector("code");
    const lang = langFromCode(code);
    if (lang === "mermaid") return;
    if (pre.parentElement?.classList.contains("code-block-wrap")) return;

    const wrap = document.createElement("div");
    wrap.className = "code-block-wrap";
    pre.parentNode?.insertBefore(wrap, pre);
    wrap.appendChild(pre);

    if (lang) {
      const badge = document.createElement("span");
      badge.className = "code-block-lang";
      badge.textContent = lang;
      wrap.appendChild(badge);
    }

    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "code-block-copy";
    btn.textContent = labels.copy;
    btn.setAttribute("aria-label", labels.copy);

    btn.addEventListener("click", async (e) => {
      e.preventDefault();
      e.stopPropagation();
      const text = code?.textContent ?? pre.textContent ?? "";
      try {
        if (navigator.clipboard?.writeText) {
          await navigator.clipboard.writeText(text);
        } else {
          const ta = document.createElement("textarea");
          ta.value = text;
          ta.style.position = "fixed";
          ta.style.left = "-9999px";
          document.body.appendChild(ta);
          ta.select();
          document.execCommand("copy");
          document.body.removeChild(ta);
        }
        btn.textContent = labels.copied;
        btn.classList.add("is-copied");
        window.setTimeout(() => {
          btn.textContent = labels.copy;
          btn.classList.remove("is-copied");
        }, 1600);
      } catch {
        btn.textContent = labels.copy;
      }
    });

    wrap.appendChild(btn);
  });
}

/** Article body HTML — must sit inside `ArticleTypographyRoot`. */
export default function ArticlePostBody({ html, contentRef, onClick }: ArticlePostBodyProps) {
  const localRef = useRef<HTMLElement | null>(null);
  const renderGen = useRef(0);
  const { t } = useTranslation("common");
  const safeHtml = useMemo(() => sanitizePublicHtml(html), [html]);

  useLayoutEffect(() => {
    const el = contentRef?.current ?? localRef.current;
    if (!el || !safeHtml) return;

    const gen = ++renderGen.current;
    let cancelled = false;
    let stopObserve: (() => void) | undefined;

    const runBatch = (nodes: HTMLElement[], force = false) => {
      void renderMermaidNodes(nodes, { force }).catch((err) => {
        if (!cancelled) console.warn("Mermaid batch failed:", err);
      });
    };

    const raf = requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        if (cancelled || gen !== renderGen.current) return;

        enhanceCodeBlocks(el, {
          copy: t("blog.copyCode"),
          copied: t("blog.copiedCode"),
        });

        // Persist source attrs on raw divs before first render.
        el.querySelectorAll<HTMLElement>(".mermaid, [data-type='mermaid']").forEach((node) => {
          if (!node.getAttribute("data-mermaid-source")) {
            const src = sourceOf(node);
            if (src) node.setAttribute("data-mermaid-source", src);
          }
        });

        const nodes = collectMermaidNodes(el);
        runBatch(nodes, false);

        stopObserve = observeMermaidVisibility(el, (need) => {
          if (cancelled || gen !== renderGen.current) return;
          runBatch(need, true);
        });
      });
    });

    return () => {
      cancelled = true;
      cancelAnimationFrame(raf);
      stopObserve?.();
    };
  }, [safeHtml, contentRef, t]);

  return (
    <article
      ref={(node) => {
        localRef.current = node;
        if (contentRef) {
          (contentRef as React.MutableRefObject<HTMLElement | null>).current = node;
        }
      }}
      className="tiptap ProseMirror max-w-none article-public-view"
      dangerouslySetInnerHTML={{ __html: safeHtml }}
      onClick={onClick}
    />
  );
}
