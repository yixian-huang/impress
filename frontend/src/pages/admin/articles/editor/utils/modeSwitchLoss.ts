/**
 * Detect TipTap HTML features that do not round-trip cleanly through
 * htmlToMarkdown / markdownToHtml (scheme 3: confirm only when lossy).
 */

export type ModeSwitchLossKey =
  | "color"
  | "highlight"
  | "fontSize"
  | "lineHeight"
  | "textAlign"
  | "columns"
  | "details"
  | "gallery"
  | "subSup"
  | "mediaWidth";

export const MODE_SWITCH_LOSS_LABELS: Record<ModeSwitchLossKey, string> = {
  color: "文字颜色",
  highlight: "高亮背景",
  fontSize: "字号",
  lineHeight: "行高",
  textAlign: "对齐方式",
  columns: "多栏布局",
  details: "折叠块",
  gallery: "图片画廊",
  subSup: "上/下标",
  mediaWidth: "媒体尺寸",
};

type LossRule = {
  key: ModeSwitchLossKey;
  /** True when HTML contains this feature */
  test: (html: string) => boolean;
};

const LOSS_RULES: LossRule[] = [
  // TipTap Color on textStyle: style="color: #f00" (avoid matching background-color)
  { key: "color", test: (h) => /(?<![\w-])color\s*:/i.test(h) },
  // Highlight extension: <mark> / data-color
  { key: "highlight", test: (h) => /<mark\b/i.test(h) || /\bdata-color=/i.test(h) },
  { key: "fontSize", test: (h) => /font-size\s*:/i.test(h) },
  { key: "lineHeight", test: (h) => /line-height\s*:/i.test(h) },
  { key: "textAlign", test: (h) => /text-align\s*:/i.test(h) },
  {
    key: "columns",
    test: (h) => /data-type=["']columns["']/i.test(h) || /data-type=["']column["']/i.test(h),
  },
  {
    key: "details",
    test: (h) => /<details\b/i.test(h) || /data-type=["']details["']/i.test(h),
  },
  { key: "gallery", test: (h) => /data-type=["']image-gallery["']/i.test(h) },
  { key: "subSup", test: (h) => /<(?:sub|sup)\b/i.test(h) },
  // Resized images/videos: width attr or style width (MD image rule keeps only alt/src/title)
  {
    key: "mediaWidth",
    test: (h) =>
      /<(?:img|video)\b[^>]*(?:\bwidth\s*=|style=["'][^"']*width\s*:)/i.test(h),
  },
];

/** Scan a single HTML body for features lost when converting to Markdown. */
export function detectModeSwitchLoss(html: string | undefined | null): ModeSwitchLossKey[] {
  if (!html || html === "<p></p>") return [];
  const found: ModeSwitchLossKey[] = [];
  for (const rule of LOSS_RULES) {
    if (rule.test(html)) found.push(rule.key);
  }
  return found;
}

/** Union of losses across language bodies (order follows LOSS_RULES). */
export function detectModeSwitchLossFromBodies(
  ...htmls: Array<string | undefined | null>
): ModeSwitchLossKey[] {
  const set = new Set<ModeSwitchLossKey>();
  for (const html of htmls) {
    for (const key of detectModeSwitchLoss(html)) set.add(key);
  }
  return LOSS_RULES.map((r) => r.key).filter((k) => set.has(k));
}

/** Whether richtext → markdown needs a confirmation dialog. */
export function shouldConfirmModeSwitch(
  ...htmls: Array<string | undefined | null>
): boolean {
  return detectModeSwitchLossFromBodies(...htmls).length > 0;
}

/** Confirm copy listing lost feature labels. */
export function buildModeSwitchConfirmMessage(keys: ModeSwitchLossKey[]): string {
  if (keys.length === 0) {
    return "切换编辑模式可能丢失部分格式。是否继续？";
  }
  const labels = keys.map((k) => MODE_SWITCH_LOSS_LABELS[k]);
  return `切换到 Markdown 可能丢失以下格式：${labels.join("、")}。是否继续？`;
}
