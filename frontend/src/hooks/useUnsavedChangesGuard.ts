import { useCallback, useEffect } from "react";

const DEFAULT_MESSAGE = "有未保存的更改，确定离开？";

/**
 * Guard against losing unsaved work:
 * - browser refresh / tab close (beforeunload)
 * - in-app link clicks (capture-phase intercept)
 *
 * Returns `confirmLeave()` for programmatic navigation (back button, etc.).
 * Note: requires BrowserRouter-friendly interception; data-router useBlocker
 * is not used so this works with the existing App shell.
 */
export function useUnsavedChangesGuard(
  isDirty: boolean,
  message: string = DEFAULT_MESSAGE,
) {
  useEffect(() => {
    if (!isDirty) return;
    const onBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = message;
    };
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [isDirty, message]);

  useEffect(() => {
    if (!isDirty) return;

    const onClick = (e: MouseEvent) => {
      if (e.defaultPrevented) return;
      if (e.button !== 0) return;
      if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;

      const el = (e.target as HTMLElement | null)?.closest?.(
        "a[href]",
      ) as HTMLAnchorElement | null;
      if (!el) return;
      if (el.target === "_blank" || el.hasAttribute("download")) return;

      const raw = el.getAttribute("href");
      if (!raw || raw.startsWith("#") || raw.startsWith("javascript:")) return;

      let next: URL;
      try {
        next = new URL(el.href, window.location.href);
      } catch {
        return;
      }
      if (next.origin !== window.location.origin) return;
      if (
        next.pathname === window.location.pathname &&
        next.search === window.location.search
      ) {
        return;
      }

      if (!window.confirm(message)) {
        e.preventDefault();
        e.stopPropagation();
      }
    };

    document.addEventListener("click", onClick, true);
    return () => document.removeEventListener("click", onClick, true);
  }, [isDirty, message]);

  const confirmLeave = useCallback(() => {
    if (!isDirty) return true;
    return window.confirm(message);
  }, [isDirty, message]);

  return { confirmLeave };
}
