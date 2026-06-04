import type { FontPreset } from "./types";

/** Built-in font stacks — P1 preset library (no upload required). */
export const FONT_PRESETS: FontPreset[] = [
  {
    id: "system-ui",
    role: "sans",
    name: "System UI",
    nameZh: "系统无衬线",
    stack: 'ui-sans-serif, system-ui, -apple-system, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
  },
  {
    id: "humanist-sans",
    role: "sans",
    name: "Humanist Sans",
    nameZh: "人文无衬线",
    stack: '"Helvetica Neue", Helvetica, Arial, "PingFang SC", "Microsoft YaHei", sans-serif',
  },
  {
    id: "editorial-georgia",
    role: "serif",
    name: "Editorial Georgia",
    nameZh: "纸墨 Georgia",
    stack: "Georgia, 'Iowan Old Style', 'Palatino Linotype', 'Book Antiqua', Palatino, serif",
  },
  {
    id: "noto-serif-sc",
    role: "serif",
    name: "Noto Serif SC",
    nameZh: "思源宋体（Noto Serif SC）",
    stack: '"Noto Serif SC", "Source Han Serif SC", Georgia, "Songti SC", serif',
    googleCssUrl: "https://fonts.googleapis.com/css2?family=Noto+Serif+SC:wght@400;600;700&display=swap",
  },
  {
    id: "source-han-serif",
    role: "serif",
    name: "Source Han Serif",
    nameZh: "思源宋体（Source Han）",
    stack: '"Source Han Serif SC", "Noto Serif SC", "Songti SC", Georgia, serif',
  },
  {
    id: "lxgw-wenkai",
    role: "serif",
    name: "LXGW WenKai",
    nameZh: "霞鹜文楷",
    stack: '"LXGW WenKai", "KaiTi", "STKaiti", Georgia, serif',
    googleCssUrl: "https://fonts.googleapis.com/css2?family=LXGW+WenKai&display=swap",
  },
  {
    id: "system-mono",
    role: "mono",
    name: "System Mono",
    nameZh: "系统等宽",
    stack: 'ui-monospace, "SF Mono", Menlo, Monaco, Consolas, monospace',
  },
];

export function getFontPreset(id: string | undefined): FontPreset | undefined {
  if (!id) return undefined;
  return FONT_PRESETS.find((p) => p.id === id);
}

export function presetsForRole(role: FontPreset["role"]): FontPreset[] {
  return FONT_PRESETS.filter((p) => p.role === role);
}

export const DEFAULT_SERIF_PRESET_ID = "editorial-georgia";
export const DEFAULT_SANS_PRESET_ID = "system-ui";
export const DEFAULT_MONO_PRESET_ID = "system-mono";
