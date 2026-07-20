// Types (light)
export type {
  EditorPreset,
  EditorFeatures,
  ToolbarConfig,
  BubbleMenuConfig,
  ToolbarRow,
  ToolbarItemDef,
} from "./types";
export type { ModalControls, ModalState } from "./types-internal";
export { useModalState } from "./useModalState";

// Extension building — heavy; prefer dynamic import of RichTextEditor / LangEditorMountInner
export { buildExtensions } from "./extension-groups";

// Presets
export { getPreset, fullPreset, standardPreset, minimalPreset } from "./presets";

// Toolbar
export { default as EditorToolbar, ToolbarButton, ToolbarDivider } from "./EditorToolbar";
export { getToolbarItem, TOOLBAR_ITEMS } from "./toolbar-registry";

// Menus
export { default as EditorBubbleMenu } from "./EditorBubbleMenu";
export { default as TableBubbleMenu } from "./TableBubbleMenu";
export { default as EditorFloatingMenu } from "./EditorFloatingMenu";
export { default as LinkEditPopover } from "./LinkEditPopover";
