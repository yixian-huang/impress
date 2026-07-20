import { useLayoutEffect, useRef, type RefObject } from "react";
import { useTranslation } from "react-i18next";

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
        securityLevel: "loose",
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
    div.textContent = code.textContent ?? "";
    // If pre was wrapped for copy UI, replace the wrapper
    const wrap = pre.parentElement?.classList.contains("code-block-wrap") ? pre.parentElement : pre;
    wrap.replaceWith(div);
    out.push(div);
  });

  return out;
}

function sourceOf(node: HTMLElement): string {
  return (
    node.getAttribute("data-mermaid-source") ||
    node.getAttribute("data-source") ||
    node.textContent ||
    ""
  ).trim();
}

async function renderMermaidNodes(nodes: HTMLElement[]) {
  if (nodes.length === 0) return;
  const mermaid = await loadMermaid();

  for (const node of nodes) {
    const source = sourceOf(node);
    if (!source) continue;

    node.setAttribute("data-mermaid-source", source);
    node.setAttribute("data-type", "mermaid");
    node.classList.add("mermaid");
    node.removeAttribute("data-processed");
    node.innerHTML = source;

    try {
      await mermaid.run({ nodes: [node], suppressErrors: false });
    } catch {
      try {
        const id = `mermaid-pub-${++mermaidSeq}`;
        const { svg } = await mermaid.render(id, source);
        node.innerHTML = svg;
        node.setAttribute("data-processed", "true");
      } catch (err) {
        console.warn("Mermaid render failed:", err);
        node.innerHTML = `<pre class="mermaid-error">${source.replace(/</g, "&lt;")}</pre>`;
      }
    }
  }
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

  useLayoutEffect(() => {
    const el = contentRef?.current ?? localRef.current;
    if (!el || !html) return;

    const gen = ++renderGen.current;
    let cancelled = false;

    const raf = requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        if (cancelled || gen !== renderGen.current) return;

        enhanceCodeBlocks(el, {
          copy: t("blog.copyCode"),
          copied: t("blog.copiedCode"),
        });

        const nodes = collectMermaidNodes(el);
        void renderMermaidNodes(nodes).catch((err) => {
          if (!cancelled) console.warn("Mermaid batch failed:", err);
        });
      });
    });

    return () => {
      cancelled = true;
      cancelAnimationFrame(raf);
    };
  }, [html, contentRef, t]);

  return (
    <article
      ref={(node) => {
        localRef.current = node;
        if (contentRef) {
          (contentRef as React.MutableRefObject<HTMLElement | null>).current = node;
        }
      }}
      className="tiptap ProseMirror max-w-none article-public-view"
      dangerouslySetInnerHTML={{ __html: html }}
      onClick={onClick}
    />
  );
}
