import { http } from "@/api/http";

export interface WizardPlanRequest {
  industry: string;
  style_preference: string;
  features: string[];
  content_types: string[];
  brand_name: string;
  description: string;
  locale: string;
}

export interface WizardPlan {
  recommended_theme: string;
  pages: Array<{
    name: string;
    slug: string;
    description: string;
    title: Record<string, string>;
    layout: string;
    sections: string[];
    sortOrder: number;
  }>;
  color_scheme: {
    primary: string;
    secondary: string;
    background: string;
    text: string;
    rationale?: string;
  };
  suggested_content?: Array<{
    pageSlug: string;
    heading: string;
    subheading: string;
    body: string;
    ctaText: string;
  }>;
  rationale?: string;
  [key: string]: unknown;
}

export interface ColorPalette {
  primary: string;
  secondary: string;
  accent: string;
  background: string;
  [key: string]: string;
}

export interface GeneratedContent {
  title?: string;
  body?: string;
  sections?: unknown[];
  [key: string]: unknown;
}

export async function generateWizardPlan(req: WizardPlanRequest): Promise<WizardPlan> {
  const response = await http.post<{
    recommendedTheme: string;
    pages: Array<{
      slug: string;
      title: Record<string, string>;
      layout: string;
      sections: string[];
      sortOrder: number;
    }>;
    colorScheme: WizardPlan["color_scheme"];
    suggestedContent?: WizardPlan["suggested_content"];
    rationale?: string;
  }>("/admin/wizard/generate-plan", {
    industry: req.industry,
    stylePreference: req.style_preference,
    features: req.features,
    contentTypes: req.content_types,
    brandName: req.brand_name,
    description: req.description,
    locale: req.locale,
  });
  const plan = response.data;
  return {
    recommended_theme: plan.recommendedTheme,
    pages: plan.pages.map((page) => ({
      ...page,
      name: page.title.zh || page.title.en || page.slug,
      description: [page.layout, ...page.sections].filter(Boolean).join(" · "),
    })),
    color_scheme: plan.colorScheme,
    suggested_content: plan.suggestedContent,
    rationale: plan.rationale,
  };
}

export async function applyWizardPlan(plan: WizardPlan): Promise<{ success: boolean; pages_created: number }> {
  const response = await http.post<{ createdPages: string[] }>("/admin/wizard/apply-plan", {
    recommendedTheme: plan.recommended_theme,
    pages: plan.pages.map(({ slug, title, layout, sections, sortOrder }) => ({
      slug,
      title,
      layout,
      sections,
      sortOrder,
    })),
    colorScheme: plan.color_scheme,
    suggestedContent: plan.suggested_content || [],
    rationale: plan.rationale || "",
  });
  return {
    success: true,
    pages_created: response.data.createdPages.length,
  };
}

export async function suggestColors(industry: string, brand_name: string): Promise<ColorPalette> {
  const response = await http.post<ColorPalette>("/admin/wizard/suggest-colors", {
    industry,
    brandName: brand_name,
  });
  return response.data;
}

export async function generateContent(page_type: string, industry: string, locale: string): Promise<GeneratedContent> {
  const response = await http.post<GeneratedContent>("/admin/wizard/generate-content", {
    pageType: page_type,
    industry,
    locale,
  });
  return response.data;
}
