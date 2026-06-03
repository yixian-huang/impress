import type { Locale, LocaleMode, LocalizedString } from "@/lib/locale";

export interface SiteConfigIdentity {
  name: LocalizedString;
  tagline?: LocalizedString;
  localeMode: LocaleMode;
  defaultLocale: Locale;
}

export interface SiteConfigBrand {
  logo: { light: string; dark?: string };
  favicon: string;
  ogImage: string;
  primaryColor: string;
  accentColor?: string;
}

export type SocialKind =
  | "github"
  | "twitter"
  | "email"
  | "rss"
  | "linkedin"
  | "custom";

export interface SiteConfigSocial {
  kind: SocialKind;
  url: string;
  label?: string;
}

export interface SiteConfigAuthor {
  name: string;
  avatar?: string;
  bio?: LocalizedString;
  location?: string;
  socials: SiteConfigSocial[];
}

export interface SiteConfigFooterLink {
  label: LocalizedString;
  url: string;
}

export interface SiteConfigFooter {
  copyright?: LocalizedString;
  icp?: string;
  extraLinks?: SiteConfigFooterLink[];
}

export interface SiteConfigSEO {
  defaultTitle?: LocalizedString;
  titleTemplate?: string;
  defaultDescription?: LocalizedString;
  twitterHandle?: string;
}

export type HeaderBrandMode = "text" | "logo" | "avatar" | "none";

export interface SiteConfigHeader {
  brandMode?: HeaderBrandMode;
  showRssLink?: boolean;
  showSocials?: boolean;
}

export interface SiteConfigGlobal {
  identity: SiteConfigIdentity;
  brand: SiteConfigBrand;
  author: SiteConfigAuthor;
  footer: SiteConfigFooter;
  seo: SiteConfigSEO;
  header?: SiteConfigHeader;
}

export interface SiteConfigFeatures {
  publicPages: {
    home: boolean;
    blog: boolean;
    contact: boolean;
    about: boolean;
    experts: boolean;
    coreServices: boolean;
    advantages: boolean;
    cases: boolean;
  };
  blog: {
    comments: boolean;
    rss: boolean;
  };
}

// Hard-coded defaults used when published config is missing or partial.
// Kept here (not in a context) so test code can import directly.
export const SITE_CONFIG_GLOBAL_DEFAULT: SiteConfigGlobal = {
  identity: {
    name: { zh: "My Site" },
    localeMode: "mono-zh",
    defaultLocale: "zh",
  },
  brand: {
    logo: { light: "" },
    favicon: "",
    ogImage: "",
    primaryColor: "#1e40af",
  },
  author: { name: "", socials: [] },
  footer: {},
  seo: {},
};

export const SITE_CONFIG_FEATURES_DEFAULT: SiteConfigFeatures = {
  publicPages: {
    home: true,
    blog: true,
    contact: true,
    about: false,
    experts: false,
    coreServices: false,
    advantages: false,
    cases: false,
  },
  blog: { comments: true, rss: true },
};
