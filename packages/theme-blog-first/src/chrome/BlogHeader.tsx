import { useLocation } from "react-router-dom";
import {
  BaseSiteHeader,
  BrandMark,
  HeaderUtilities,
  useBranding,
  useContentMaxWidth,
  useHeaderSettings,
  useIsReadingLayout,
  useIsThemeHomePath,
  type HeaderChromeProps,
} from "@inkless/theme-host";
import { resolveBlogHomeBrandMode } from "./resolveHomeBrand";

export default function BlogHeader({ config }: HeaderChromeProps) {
  const { brandMode } = useHeaderSettings();
  const branding = useBranding();
  const maxWidth = useContentMaxWidth();
  const isReading = useIsReadingLayout();
  const isThemeHome = useIsThemeHomePath();
  const { pathname } = useLocation();
  const compactHome = isReading && isThemeHome;
  const isAuthorPage = pathname === "/author" || pathname.startsWith("/author/");
  const resolvedBrandMode = resolveBlogHomeBrandMode(brandMode, branding, compactHome);

  return (
    <BaseSiteHeader
      config={config}
      variant="blog"
      languagePlacement="inline"
      headerClassName={
        compactHome
          ? "bg-surface/90 backdrop-blur-sm border-b border-border/60 font-sans"
          : "bg-surface border-b border-border font-sans"
      }
      navPaddingClassName={compactHome ? "py-2.5" : "py-3.5"}
      containerClassName="mx-auto px-4 md:px-content w-full"
      containerStyle={{ maxWidth }}
      brand={
        <BrandMark
          brandMode={resolvedBrandMode}
          hideDefaultLogo
          showLabel={!compactHome}
          textClassName="text-sm font-sans font-medium tracking-tight text-on-surface"
          avatarClassName="h-8 w-8 rounded-full object-contain bg-[#141310] ring-1 ring-border"
          logoClassName="h-6 w-auto opacity-90"
        />
      }
      // Home + author: RSS only — socials live on /author body, not the top bar.
      utilities={<HeaderUtilities hideSocials={compactHome || isAuthorPage} />}
    />
  );
}
