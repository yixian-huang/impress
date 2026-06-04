import { useAuth } from "@/contexts/AuthContext";

/** True when an admin session can post author replies via /admin/comments/reply. */
export function useCanAuthorReply(): boolean {
  const { isAuthenticated, hasPermission, user } = useAuth();
  if (!isAuthenticated) return false;
  if (user?.isSuperAdmin) return true;
  return (
    hasPermission("comments:create") ||
    hasPermission("comments:update") ||
    hasPermission("comments:delete")
  );
}
