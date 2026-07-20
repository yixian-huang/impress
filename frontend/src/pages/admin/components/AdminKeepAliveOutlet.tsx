import { Suspense, useEffect, useRef, useState, type ReactElement } from "react";
import { useLocation, useOutlet } from "react-router-dom";
import { AdminRouteFallback } from "@/components/admin/ui";
import {
  ADMIN_KEEP_ALIVE_MAX,
  onAdminKeepAliveClear,
  resolveKeepAliveKey,
  touchKeepAliveLru,
} from "@/pages/admin/adminKeepAlive";

/**
 * Keeps recently visited admin list/settings screens mounted (hidden when
 * inactive) so filter state, scroll, and in-memory UI survive menu switches.
 * Editors and other non-listed routes render normally without caching.
 */
export default function AdminKeepAliveOutlet() {
  const outlet = useOutlet();
  const { pathname } = useLocation();
  const keepKey = resolveKeepAliveKey(pathname);

  const cacheRef = useRef(new Map<string, ReactElement>());
  const lruRef = useRef<string[]>([]);
  const scrollRef = useRef(new Map<string, number>());
  const prevKeepKeyRef = useRef<string | null>(null);
  const [, setVersion] = useState(0);

  // Persist current outlet the first time we visit a keep-alive path.
  // Do not replace existing entries — identity stability preserves React state.
  if (keepKey && outlet && !cacheRef.current.has(keepKey)) {
    cacheRef.current.set(keepKey, outlet as ReactElement);
    const { order, evicted } = touchKeepAliveLru(lruRef.current, keepKey, ADMIN_KEEP_ALIVE_MAX);
    lruRef.current = order;
    for (const key of evicted) {
      cacheRef.current.delete(key);
      scrollRef.current.delete(key);
    }
  } else if (keepKey && cacheRef.current.has(keepKey)) {
    const { order } = touchKeepAliveLru(lruRef.current, keepKey, ADMIN_KEEP_ALIVE_MAX);
    lruRef.current = order;
  }

  // Save / restore window scroll for keep-alive panes
  useEffect(() => {
    const prev = prevKeepKeyRef.current;
    if (prev && prev !== keepKey) {
      scrollRef.current.set(prev, window.scrollY);
    }
    prevKeepKeyRef.current = keepKey;

    if (keepKey && scrollRef.current.has(keepKey)) {
      const y = scrollRef.current.get(keepKey) ?? 0;
      requestAnimationFrame(() => {
        window.scrollTo(0, y);
      });
    } else if (!keepKey) {
      // Fresh non-cached route (e.g. editor) — start at top
      window.scrollTo(0, 0);
    }
  }, [keepKey, pathname]);

  useEffect(() => {
    return onAdminKeepAliveClear(() => {
      cacheRef.current.clear();
      lruRef.current = [];
      scrollRef.current.clear();
      prevKeepKeyRef.current = null;
      setVersion((v) => v + 1);
    });
  }, []);

  const cachedEntries = Array.from(cacheRef.current.entries());

  return (
    <>
      {cachedEntries.map(([key, element]) => {
        const active = key === keepKey;
        return (
          <div
            key={key}
            data-admin-keep-alive={key}
            hidden={!active}
            className={active ? undefined : "hidden"}
            aria-hidden={!active}
          >
            <Suspense fallback={active ? <AdminRouteFallback /> : null}>{element}</Suspense>
          </div>
        );
      })}

      {!keepKey && (
        <Suspense fallback={<AdminRouteFallback />}>{outlet}</Suspense>
      )}
    </>
  );
}
