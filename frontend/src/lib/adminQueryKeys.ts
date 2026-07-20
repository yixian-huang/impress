/** Canonical query key roots for admin list/detail caches. */
export const adminQueryKeys = {
  dashboardStats: ["admin", "dashboard", "stats"] as const,
  articles: ["admin", "articles"] as const,
  pages: ["admin", "pages"] as const,
  media: ["admin", "media"] as const,
  users: ["admin", "users"] as const,
  roles: ["admin", "roles"] as const,
  formSubmissions: ["admin", "form-submissions"] as const,
  scheduled: ["admin", "scheduled-publications"] as const,
  analytics: ["admin", "analytics"] as const,
  auditLogs: ["admin", "audit-logs"] as const,
  comments: ["admin", "comments"] as const,
};
