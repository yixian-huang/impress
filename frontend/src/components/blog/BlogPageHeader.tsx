import { Link } from "react-router-dom";
import { useIsReadingLayout } from "@/plugins/hooks";

interface BlogPageHeaderProps {
  title: string;
  description?: string;
  backTo?: { href: string; label: string };
}

export default function BlogPageHeader({ title, description, backTo }: BlogPageHeaderProps) {
  const isReading = useIsReadingLayout();

  return (
    <header className={isReading ? "mb-8 pb-6 border-b border-border/80" : "mb-10"}>
      {backTo && (
        <p className={`article-page-ui font-sans ${isReading ? "mb-4 text-sm" : "mb-3 text-sm"}`}>
          <Link
            to={backTo.href}
            className="text-on-surface-muted hover:text-primary transition-colors"
          >
            ← {backTo.label}
          </Link>
        </p>
      )}
      <h1
        className={
          isReading
            ? "article-page-title text-3xl md:text-4xl font-heading font-normal text-on-surface tracking-tight"
            : "font-heading text-3xl font-bold text-on-surface"
        }
      >
        {title}
      </h1>
      {description && (
        <p
          className={
            isReading
              ? "mt-3 text-base text-on-surface-muted leading-relaxed font-sans"
              : "mt-2 text-on-surface-muted"
          }
        >
          {description}
        </p>
      )}
    </header>
  );
}
