interface AuthorIntroProps {
  avatar?: string;
  name: string;
  tagline?: string;
  bio?: string;
  intro: string;
}

export default function AuthorIntro({ avatar, name, tagline, bio, intro }: AuthorIntroProps) {
  return (
    <header className="mb-12">
      {avatar && (
        <img
          src={avatar}
          alt={name}
          className="w-20 h-20 rounded-full object-cover mb-4 border border-border"
        />
      )}
      <h1 className="text-3xl md:text-4xl font-heading font-bold text-on-surface tracking-tight">
        {name}
      </h1>
      {tagline && bio && (
        <p className="mt-2 text-lg text-on-surface-muted">{tagline}</p>
      )}
      {intro && (
        <p className="mt-4 text-on-surface leading-relaxed whitespace-pre-wrap">{intro}</p>
      )}
    </header>
  );
}
