import { useEffect, useState } from "react";
import { getPublicCategories, getPublicTags, type Category, type Tag } from "@/api/articles";

/** Shared public categories + tags for archive navigation. */
export function useBlogTaxonomy() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    Promise.all([getPublicCategories(), getPublicTags()])
      .then(([cats, tagList]) => {
        if (cancelled) return;
        setCategories(cats.filter((c) => !c.hideFromList));
        setTags(tagList);
      })
      .catch(() => {})
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return { categories, tags, loading };
}
