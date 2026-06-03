import type { LayoutConfig } from "./types";

export const CORPORATE_DEFAULT_LAYOUT: LayoutConfig = {
  type: "default",
  header: { style: "sticky" },
  footer: { style: "full" },
};

export const BLOG_DEFAULT_LAYOUT: LayoutConfig = {
  type: "default",
  header: { style: "sticky" },
  footer: { style: "minimal" },
};
