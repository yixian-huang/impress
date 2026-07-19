import { useEffect } from "react";
import { PRODUCT_BRAND, PRODUCT_DEFAULT_FAVICON, PRODUCT_DEFAULT_OG_IMAGE } from "@/config/productBrand";
import { useBranding } from "@/hooks/useBranding";
import { useSEODefaults } from "@/hooks/useSEODefaults";

function setMetaTag(attr: string, key: string, content: string): HTMLMetaElement {
  let el = document.querySelector<HTMLMetaElement>(`meta[${attr}="${key}"]`);
  if (!el) {
    el = document.createElement("meta");
    el.setAttribute(attr, key);
    document.head.appendChild(el);
  }
  el.setAttribute("content", content);
  return el;
}

export default function DocumentBranding() {
  const branding = useBranding();
  const { defaultDescription, defaultOgImage } = useSEODefaults();

  useEffect(() => {
    const favicon = branding.favicon?.trim() || PRODUCT_DEFAULT_FAVICON;
    let icon = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
    if (!icon) {
      icon = document.createElement("link");
      icon.setAttribute("rel", "icon");
      document.head.appendChild(icon);
    }
    icon.setAttribute("href", favicon);
    icon.setAttribute("type", favicon.endsWith(".svg") ? "image/svg+xml" : "image/png");

    const description = defaultDescription || PRODUCT_BRAND.description;
    const ogImage = defaultOgImage || PRODUCT_DEFAULT_OG_IMAGE;

    setMetaTag("name", "description", description);
    setMetaTag("property", "og:image", ogImage);
    setMetaTag("name", "twitter:card", "summary_large_image");
    setMetaTag("name", "twitter:description", description);
    setMetaTag("name", "twitter:image", ogImage);
  }, [branding.favicon, defaultDescription, defaultOgImage]);

  return null;
}
