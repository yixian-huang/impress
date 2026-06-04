export { default as CommentSlot } from "./public/CommentSlot";
export { default as CommentSection } from "./public/CommentSection";
export { CommentFeatureGate } from "./CommentFeatureGate";
export { useCanAuthorReply } from "./useCanAuthorReply";
export { useCommentsEnabled } from "./useCommentsEnabled";
export * from "./api";

export const commentModuleConfig = {
  name: "comment",
  adminRoute: {
    path: "comments",
  },
  sidebar: {
    label: "评论管理",
    path: "/admin/comments",
    permissionKey: "comments",
  },
} as const;
