/** Canonical shortcut list for the article editor cheatsheet. */

export type ShortcutDef = {
  id: string;
  keys: string[];
  label: string;
  group: "file" | "view" | "edit" | "mode";
};

export const EDITOR_SHORTCUTS: ShortcutDef[] = [
  { id: "save", keys: ["⌘", "S"], label: "保存草稿", group: "file" },
  { id: "publish", keys: ["⌘", "⇧", "S"], label: "发布", group: "file" },
  { id: "preview", keys: ["⌘", "P"], label: "预览", group: "view" },
  { id: "find", keys: ["⌘", "F"], label: "查找 / 替换", group: "edit" },
  { id: "zen", keys: ["⌘", "\\"], label: "专注模式", group: "view" },
  { id: "help", keys: ["⌘", "/"], label: "快捷键帮助", group: "view" },
  { id: "esc", keys: ["Esc"], label: "关闭面板 / 退出专注", group: "view" },
  { id: "bold", keys: ["⌘", "B"], label: "粗体（富文本 / MD）", group: "edit" },
  { id: "italic", keys: ["⌘", "I"], label: "斜体（富文本 / MD）", group: "edit" },
  { id: "link", keys: ["⌘", "K"], label: "链接（Markdown）", group: "edit" },
];

export const SHORTCUT_GROUP_LABELS: Record<ShortcutDef["group"], string> = {
  file: "文件",
  view: "视图",
  edit: "编辑",
  mode: "模式",
};

/** Display-friendly: replace ⌘ with Ctrl on non-Apple platforms when needed. */
export function formatShortcutKeys(keys: string[], isApple: boolean): string[] {
  if (isApple) return keys;
  return keys.map((k) => {
    if (k === "⌘") return "Ctrl";
    if (k === "⇧") return "Shift";
    if (k === "⌥") return "Alt";
    return k;
  });
}

export function detectApplePlatform(): boolean {
  if (typeof navigator === "undefined") return true;
  return /Mac|iPhone|iPad|iPod/i.test(navigator.platform || navigator.userAgent || "");
}
