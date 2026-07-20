import { Link } from "react-router-dom";
import AuthorSocialLinks from "@/components/blog/AuthorSocialLinks";
import { useIsReadingLayout, useIsThemeHomePath } from "@/plugins/hooks";

interface AuthorIntroProps {
  avatar?: string;
  name: string;
  tagline?: string;
  bio?: string;
  intro: string;
  /** Optional secondary line under the name (e.g. English name). Default: hidden on compact home. */
  subtitle?: string;
  /**
   * Hero social links. Prefer false when Header already shows socials
   * (avoids duplicate GitHub/Twitter on the home first screen).
   */
  showSocials?: boolean;
  /** When false (default on home), bio/intro body is omitted to free first-screen space. */
  showBio?: boolean;
  /** When false (default on home), hide English/alternate subtitle under the name. */
  showSubtitle?: boolean;
  /** Optional quiet link under the home hero (e.g. to the author page). */
  moreHref?: string;
  moreLabel?: string;
}

/**
 * Author identity block: compact on theme home, full on author page.
 * Socials belong on the author page (or header utilities), not both.
 */
export default function AuthorIntro({
  avatar,
  name,
  tagline,
  bio,
  intro,
  subtitle,
  showSocials = false,
  showBio = false,
  showSubtitle = false,
  moreHref,
  moreLabel,
}: AuthorIntroProps) {
  const isReading = useIsReadingLayout();
  const isThemeHome = useIsThemeHomePath();
  const isHomeHero = isReading && isThemeHome;

  const showHeroAvatar = Boolean(avatar) && (!isReading || isThemeHome || showBio);
  const bodyIntro =
    bio?.trim() ||
    (intro.trim() && intro.trim() !== tagline?.trim() ? intro.trim() : "");

  if (isHomeHero) {
    return (
      <header className="mb-6 md:mb-8">
        <div className="flex flex-col items-center text-center px-1 pt-1 pb-4 md:pt-2 md:pb-5">
          {showHeroAvatar && (
            <img
              src={avatar}
              alt=""
              className="w-10 h-10 md:w-11 md:h-11 rounded-full object-contain bg-[#141310] mb-3 ring-1 ring-border/70"
            />
          )}
          <h1 className="text-xl sm:text-[1.35rem] md:text-2xl font-heading font-normal text-on-surface tracking-tight leading-tight">
            {name}
          </h1>
          {showSubtitle && subtitle?.trim() && (
            <p className="mt-1 text-[11px] font-sans tracking-[0.12em] text-on-surface-muted uppercase">
              {subtitle.trim()}
            </p>
          )}
          {tagline?.trim() && (
            <p className="mt-1.5 max-w-sm text-sm font-sans font-normal text-on-surface-muted leading-snug">
              {tagline.trim()}
            </p>
          )}
          {moreHref && moreLabel && (
            <p className="mt-3">
              <Link
                to={moreHref}
                className="text-[11px] font-sans uppercase tracking-[0.14em] text-on-surface-muted hover:text-on-surface transition-colors"
              >
                {moreLabel}
              </Link>
            </p>
          )}
        </div>
        <div className="border-t border-border" aria-hidden="true" />
      </header>
    );
  }

  return (
    <header className={isReading ? "mb-10 pb-8 border-b border-border" : "mb-12"}>
      {showHeroAvatar && (
        <img
          src={avatar}
          alt={name}
          className={
            isReading
              ? "w-20 h-20 rounded-full object-contain bg-[#141310] mb-5 ring-1 ring-border"
              : "w-20 h-20 rounded-full object-cover mb-4 border border-border"
          }
        />
      )}
      <h1
        className={
          isReading
            ? "text-3xl md:text-4xl font-heading font-normal text-on-surface tracking-tight leading-[1.15]"
            : "text-3xl md:text-4xl font-heading font-bold text-on-surface tracking-tight"
        }
      >
        {name}
      </h1>
      {tagline && (
        <p
          className={
            isReading
              ? "mt-2 text-base text-on-surface-muted font-sans font-normal"
              : "mt-2 text-lg text-on-surface-muted"
          }
        >
          {tagline}
        </p>
      )}
      {(showBio || !isReading) && bodyIntro && (
        <p
          className={
            isReading
              ? "mt-4 text-on-surface text-base leading-relaxed whitespace-pre-wrap max-w-prose"
              : "mt-4 text-on-surface leading-relaxed whitespace-pre-wrap"
          }
        >
          {bodyIntro}
        </p>
      )}
      {showSocials && (
        <div className="mt-6">
          <AuthorSocialLinks
            className={
              isReading
                ? "flex flex-wrap items-center gap-x-5 gap-y-2"
                : "flex flex-wrap items-center justify-center gap-x-5 gap-y-2"
            }
            linkClassName="text-sm text-on-surface-muted hover:text-primary transition-colors"
          />
        </div>
      )}
    </header>
  );
}
