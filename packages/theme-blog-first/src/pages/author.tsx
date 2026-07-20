import { useTranslation } from "react-i18next";
import {
  AuthorIntro,
  BlogPageShell,
  pickLocaleValue,
  SeoHead,
  SITE_CONFIG_GLOBAL_DEFAULT,
  useGlobalConfig,
  useLocaleMode,
  useSEODefaults,
} from "@inkless/theme-host";

/**
 * blog-first author page — full identity, bio, and socials.
 * Home stays article-first; header socials stay quiet on home.
 */
export default function BlogFirstAuthorPage() {
  const { t } = useTranslation("common");
  const { config } = useGlobalConfig();
  const { buildTitle, defaultDescription, defaultOgImage } = useSEODefaults();
  const { localeMode, defaultLocale, currentLocale } = useLocaleMode();

  const siteConfig = config.siteConfig ?? SITE_CONFIG_GLOBAL_DEFAULT;
  const siteName = pickLocaleValue({
    value: siteConfig.identity.name,
    mode: localeMode,
    defaultLocale,
    currentLocale,
  });
  const authorName = siteConfig.author?.name?.trim() || siteName;
  const bio = pickLocaleValue({
    value: siteConfig.author?.bio,
    mode: localeMode,
    defaultLocale,
    currentLocale,
  });
  const tagline = pickLocaleValue({
    value: siteConfig.identity.tagline,
    mode: localeMode,
    defaultLocale,
    currentLocale,
  });
  const intro = bio || tagline || defaultDescription;
  const pageLabel = t("blog.authorPageTitle");

  return (
    <>
      <SeoHead
        title={buildTitle(pageLabel)}
        description={intro}
        ogTitle={authorName}
        ogDescription={intro}
        ogImage={siteConfig.author?.avatar || siteConfig.brand.ogImage || defaultOgImage}
        ogType="profile"
        canonicalUrl="/author"
      />
      <BlogPageShell>
        <AuthorIntro
          avatar={siteConfig.author?.avatar}
          name={authorName}
          tagline={tagline}
          bio={bio}
          intro={intro}
          showSocials
          showBio
          showSubtitle={false}
        />
      </BlogPageShell>
    </>
  );
}
