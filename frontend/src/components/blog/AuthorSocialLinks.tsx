import { useTranslation } from "react-i18next";
import { useBranding } from "@/hooks/useBranding";
import type { SiteConfigSocial } from "@/types/siteConfig";

function socialLabel(s: SiteConfigSocial): string {
  if (s.label?.trim()) return s.label.trim();
  if (s.kind === "custom") return s.url;
  return s.kind.charAt(0).toUpperCase() + s.kind.slice(1);
}

interface AuthorSocialLinksProps {
  className?: string;
  linkClassName?: string;
}

export default function AuthorSocialLinks({
  className = "flex flex-wrap items-center justify-center gap-x-5 gap-y-2",
  linkClassName = "text-sm text-on-surface-muted hover:text-primary transition-colors",
}: AuthorSocialLinksProps) {
  const { t } = useTranslation("common");
  const { author } = useBranding();
  const links = author.socials.filter((s) => s.url?.trim());

  if (links.length === 0) return null;

  return (
    <nav className={className} aria-label={t("blog.socialLinks")}>
      {links.map((s) => (
        <a
          key={`${s.kind}-${s.url}`}
          href={s.url.trim()}
          className={linkClassName}
          target={s.kind === "email" ? undefined : "_blank"}
          rel={s.kind === "email" ? undefined : "noopener noreferrer"}
        >
          {socialLabel(s)}
        </a>
      ))}
    </nav>
  );
}
