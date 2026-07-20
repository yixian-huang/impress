import {
  SITE_CONFIG_GLOBAL_DEFAULT,
  type SiteConfigGlobal,
  type SiteConfigSocial,
} from "@/types/siteConfig";

/** Deep-merge API payload with defaults so partial drafts never crash the form. */
export function normalizeSiteConfig(raw: unknown): SiteConfigGlobal {
  const d = SITE_CONFIG_GLOBAL_DEFAULT;
  if (!raw || typeof raw !== "object") return structuredClone(d);

  const r = raw as Partial<SiteConfigGlobal> & Record<string, unknown>;
  const identity = (r.identity ?? {}) as Partial<SiteConfigGlobal["identity"]>;
  const brand = (r.brand ?? {}) as Partial<SiteConfigGlobal["brand"]>;
  const logo = (brand.logo ?? {}) as Partial<SiteConfigGlobal["brand"]["logo"]>;
  const author = (r.author ?? {}) as Partial<SiteConfigGlobal["author"]>;
  const footer = (r.footer ?? {}) as Partial<SiteConfigGlobal["footer"]>;
  const seo = (r.seo ?? {}) as Partial<SiteConfigGlobal["seo"]>;
  const header = (r.header ?? {}) as Partial<NonNullable<SiteConfigGlobal["header"]>>;

  const socials: SiteConfigSocial[] = Array.isArray(author.socials)
    ? author.socials.filter((s): s is SiteConfigSocial => !!s && typeof s === "object")
    : [];

  return {
    identity: {
      name: { zh: identity.name?.zh ?? d.identity.name.zh, en: identity.name?.en },
      tagline: identity.tagline
        ? { zh: identity.tagline.zh, en: identity.tagline.en }
        : undefined,
      localeMode: identity.localeMode ?? d.identity.localeMode,
      defaultLocale: identity.defaultLocale ?? d.identity.defaultLocale,
    },
    brand: {
      logo: {
        light: logo.light ?? d.brand.logo.light,
        dark: logo.dark,
      },
      favicon: brand.favicon ?? d.brand.favicon,
      ogImage: brand.ogImage ?? d.brand.ogImage,
      primaryColor: brand.primaryColor ?? d.brand.primaryColor,
      accentColor: brand.accentColor,
    },
    author: {
      name: author.name ?? d.author.name,
      avatar: author.avatar,
      bio: author.bio ? { zh: author.bio.zh, en: author.bio.en } : undefined,
      location: author.location,
      socials,
    },
    footer: {
      copyright: footer.copyright
        ? { zh: footer.copyright.zh, en: footer.copyright.en }
        : undefined,
      icp: footer.icp,
      extraLinks: footer.extraLinks,
    },
    seo: {
      defaultTitle: seo.defaultTitle
        ? { zh: seo.defaultTitle.zh, en: seo.defaultTitle.en }
        : undefined,
      titleTemplate: seo.titleTemplate,
      defaultDescription: seo.defaultDescription
        ? { zh: seo.defaultDescription.zh, en: seo.defaultDescription.en }
        : undefined,
      twitterHandle: seo.twitterHandle,
    },
    header: {
      brandMode: header.brandMode,
      showRssLink: header.showRssLink,
      showSocials: header.showSocials,
    },
  };
}
