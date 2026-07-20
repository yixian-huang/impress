import {
  AuthorSocialLinks,
  useBranding,
  useContentMaxWidth,
  useHeaderSettings,
  useIsReadingLayout,
  type FooterChromeProps,
} from "@inkless/theme-host";

export default function BlogFooter({ config }: FooterChromeProps) {
  const branding = useBranding();
  const maxWidth = useContentMaxWidth();
  const isReading = useIsReadingLayout();
  const { showSocials } = useHeaderSettings();
  const copyright = config?.copyright ?? branding.footer.copyright;
  const style = config?.style ?? "minimal";

  // Never duplicate header socials. Only show in footer when header is not
  // already displaying them (and the author has links).
  const showFooterSocials =
    !showSocials && branding.author.socials.some((s) => s.url?.trim());

  if (style === "none") {
    return null;
  }

  return (
    <footer className="mt-auto border-t border-border bg-surface font-sans">
      <div className="mx-auto px-4 md:px-content py-8 w-full" style={{ maxWidth }}>
        <div
          className={
            isReading
              ? "text-center space-y-3"
              : "flex flex-col items-center gap-3 text-center"
          }
        >
          {showFooterSocials && <AuthorSocialLinks />}
          <p className="text-sm text-on-surface-muted">{copyright}</p>
        </div>
        {branding.footer.icp && (
          <p className="text-xs text-on-surface-muted mt-3 text-center">{branding.footer.icp}</p>
        )}
      </div>
    </footer>
  );
}
