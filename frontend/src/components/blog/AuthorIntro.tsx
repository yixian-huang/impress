import { useIsReadingLayout, useIsThemeHomePath } from "@/plugins/hooks";

interface AuthorIntroProps {
  avatar?: string;
  name: string;
  tagline?: string;
  bio?: string;
  intro: string;
}

export default function AuthorIntro({ avatar, name, tagline, bio, intro }: AuthorIntroProps) {
  const isReading = useIsReadingLayout();
  const isThemeHome = useIsThemeHomePath();
  const showHeroAvatar = Boolean(avatar) && (!isReading || isThemeHome);
  const bodyIntro =
    bio?.trim() ||
    (intro.trim() && intro.trim() !== tagline?.trim() ? intro.trim() : "");

  return (
    <header className={isReading ? "mb-10 pb-8 border-b border-border" : "mb-12"}>
      {showHeroAvatar && (
        <img
          src={avatar}
          alt={name}
          className={
            isReading
              ? "w-24 h-24 rounded-full object-cover mb-6 ring-1 ring-border"
              : "w-20 h-20 rounded-full object-cover mb-4 border border-border"
          }
        />
      )}
      <h1
        className={
          isReading
            ? "text-4xl md:text-[2.75rem] font-heading font-normal text-on-surface tracking-tight leading-[1.15]"
            : "text-3xl md:text-4xl font-heading font-bold text-on-surface tracking-tight"
        }
      >
        {name}
      </h1>
      {tagline && (
        <p
          className={
            isReading
              ? "mt-3 text-lg text-on-surface-muted font-sans font-normal"
              : "mt-2 text-lg text-on-surface-muted"
          }
        >
          {tagline}
        </p>
      )}
      {bodyIntro && (
        <p
          className={
            isReading
              ? "mt-6 text-on-surface text-[1.125rem] leading-relaxed whitespace-pre-wrap"
              : "mt-4 text-on-surface leading-relaxed whitespace-pre-wrap"
          }
        >
          {bodyIntro}
        </p>
      )}
    </header>
  );
}
