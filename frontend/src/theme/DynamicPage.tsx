import { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { http } from "@/api/http";
import type { Locale } from "@/api/publicContent";
import type { PageConfig } from "./types";
import { SectionRenderer } from "./sections";

interface DynamicPageProps {
  slug?: string;
}

export default function DynamicPage({ slug: slugProp }: DynamicPageProps = {}) {
  const { "*": paramSlug } = useParams();
  const slug = slugProp || paramSlug;
  const { i18n } = useTranslation("common");
  const locale = (i18n.language === "zh" || i18n.language.startsWith("zh") ? "zh" : "en") as Locale;

  const [config, setConfig] = useState<PageConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!slug) {
      setError("No page slug provided");
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    http
      .get(`/public/pages/${slug}`, { params: { locale } })
      .then((res) => {
        const raw = res.data.publishedConfig ?? res.data.config ?? res.data;
        // Normalize sections: backend uses "props", frontend SectionData uses "data"
        if (raw?.sections) {
          raw.sections = raw.sections.map((s: any) => ({
            ...s,
            data: s.data || s.props || {},
          }));
        }
        setConfig(raw);
        setLoading(false);
      })
      .catch((e) => {
        setError(e.response?.data?.error || e.message);
        setLoading(false);
      });
  }, [slug, locale]);

  if (loading) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  if (error || !config) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        <div className="text-red-600">{error || "Page not found"}</div>
      </div>
    );
  }

  return (
    <>
      {config.sections
        .filter((s) => !s.settings?.hidden)
        .map((section) => (
          <SectionRenderer key={section.id} section={section} />
        ))}
    </>
  );
}
