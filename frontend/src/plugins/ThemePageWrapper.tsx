import { lazy, Suspense, useMemo } from "react";
import type { ComponentType } from "react";
import type { ThemePageDefinition } from "./types";
import { useGlobalConfig } from "@/contexts/GlobalConfigContext";
import { isBlogSiteMode, isHomePageDef } from "@/hooks/useSiteMode";

const DynamicPage = lazy(() => import("@/theme/DynamicPage"));
const BlogHomePage = lazy(() => import("@/pages/blog-home/page"));

function Loading() {
  return (
    <div className="min-h-[60vh] flex items-center justify-center">
      <div className="text-gray-400 animate-pulse">加载中...</div>
    </div>
  );
}

// Cache lazy components so they are only created once per lazyComponent function
const lazyCache = new WeakMap<() => Promise<{ default: ComponentType }>, ComponentType>();

function getLazyComponent(loader: () => Promise<{ default: ComponentType }>): ComponentType {
  let comp = lazyCache.get(loader);
  if (!comp) {
    comp = lazy(loader);
    lazyCache.set(loader, comp);
  }
  return comp;
}

interface ThemePageWrapperProps {
  pageDef: ThemePageDefinition;
}

export default function ThemePageWrapper({ pageDef }: ThemePageWrapperProps) {
  const { features } = useGlobalConfig();

  const Component = useMemo(() => {
    if (pageDef.renderMode === "hardcoded") {
      if (pageDef.lazyComponent) {
        return getLazyComponent(pageDef.lazyComponent);
      }
      if (pageDef.component) {
        return pageDef.component;
      }
    }
    return DynamicPage;
  }, [pageDef]);

  if (isHomePageDef(pageDef) && isBlogSiteMode(features)) {
    return (
      <Suspense fallback={<Loading />}>
        <BlogHomePage />
      </Suspense>
    );
  }

  return (
    <Suspense fallback={<Loading />}>
      <Component slug={pageDef.slug} />
    </Suspense>
  );
}
