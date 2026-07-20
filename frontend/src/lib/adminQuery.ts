import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from "react";

export type AdminQueryKey = string | readonly unknown[];

type CacheEntry = {
  data: unknown;
  updatedAt: number;
  error: Error | null;
};

const cache = new Map<string, CacheEntry>();
const listeners = new Map<string, Set<() => void>>();
const inflight = new Map<string, Promise<unknown>>();

const DEFAULT_STALE_MS = 30_000;

export function serializeAdminQueryKey(key: AdminQueryKey): string {
  if (typeof key === "string") return key;
  return JSON.stringify(key);
}

function notify(serialized: string) {
  const set = listeners.get(serialized);
  if (!set) return;
  for (const listener of set) listener();
}

function subscribe(serialized: string, listener: () => void) {
  let set = listeners.get(serialized);
  if (!set) {
    set = new Set();
    listeners.set(serialized, set);
  }
  set.add(listener);
  return () => {
    set!.delete(listener);
    if (set!.size === 0) listeners.delete(serialized);
  };
}

function getSnapshot(serialized: string): CacheEntry | undefined {
  return cache.get(serialized);
}

export function getAdminQueryData<T>(key: AdminQueryKey): T | undefined {
  return cache.get(serializeAdminQueryKey(key))?.data as T | undefined;
}

export function setAdminQueryData<T>(key: AdminQueryKey, data: T): void {
  const serialized = serializeAdminQueryKey(key);
  cache.set(serialized, { data, updatedAt: Date.now(), error: null });
  notify(serialized);
}

/**
 * Mark matching array-prefix keys as stale (updatedAt=0).
 * Example: invalidateAdminQueryPrefix(["admin", "articles"]) refreshes all article list pages.
 */
export function invalidateAdminQueryPrefix(prefix: readonly unknown[]): void {
  for (const [serialized, entry] of cache.entries()) {
    try {
      const key = JSON.parse(serialized) as unknown;
      if (!Array.isArray(key)) continue;
      if (prefix.length > key.length) continue;
      const matches = prefix.every((part, index) => Object.is(key[index], part));
      if (!matches) continue;
      cache.set(serialized, { ...entry, updatedAt: 0 });
      notify(serialized);
    } catch {
      // non-JSON keys ignored for prefix invalidation
    }
  }
}

export function clearAdminQueryCache(): void {
  cache.clear();
  inflight.clear();
  for (const set of listeners.values()) {
    for (const listener of set) listener();
  }
}

async function runFetch<T>(serialized: string, fetcher: () => Promise<T>): Promise<T> {
  const existing = inflight.get(serialized);
  if (existing) return existing as Promise<T>;

  const promise = (async () => {
    try {
      const data = await fetcher();
      cache.set(serialized, { data, updatedAt: Date.now(), error: null });
      notify(serialized);
      return data;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      const prev = cache.get(serialized);
      cache.set(serialized, {
        data: prev?.data,
        updatedAt: prev?.updatedAt ?? 0,
        error,
      });
      notify(serialized);
      throw error;
    } finally {
      inflight.delete(serialized);
    }
  })();

  inflight.set(serialized, promise);
  return promise;
}

export type UseAdminQueryResult<T> = {
  data: T | undefined;
  error: Error | null;
  /** True only when there is no cached data yet. */
  loading: boolean;
  /** True whenever a fetch is in flight (including background revalidation). */
  isFetching: boolean;
  isStale: boolean;
  refetch: (opts?: { force?: boolean }) => Promise<T | undefined>;
};

export function useAdminQuery<T>(
  key: AdminQueryKey,
  fetcher: () => Promise<T>,
  options?: {
    staleTime?: number;
    enabled?: boolean;
  },
): UseAdminQueryResult<T> {
  const serialized = serializeAdminQueryKey(key);
  const staleTime = options?.staleTime ?? DEFAULT_STALE_MS;
  const enabled = options?.enabled !== false;
  const fetcherRef = useRef(fetcher);
  fetcherRef.current = fetcher;

  const entry = useSyncExternalStore(
    (onStoreChange) => subscribe(serialized, onStoreChange),
    () => getSnapshot(serialized),
    () => getSnapshot(serialized),
  );

  const [isFetching, setIsFetching] = useState(false);
  const [localError, setLocalError] = useState<Error | null>(null);

  const data = entry?.data as T | undefined;
  const isStale =
    !entry || entry.updatedAt === 0 || Date.now() - entry.updatedAt > staleTime;
  const error = localError ?? entry?.error ?? null;

  const refetch = useCallback(
    async (opts?: { force?: boolean }) => {
      if (!enabled) return undefined;
      const current = cache.get(serialized);
      const stale =
        !current ||
        current.updatedAt === 0 ||
        Date.now() - current.updatedAt > staleTime;
      if (!opts?.force && !stale && current?.data !== undefined) {
        return current.data as T;
      }
      setIsFetching(true);
      setLocalError(null);
      try {
        return await runFetch(serialized, () => fetcherRef.current());
      } catch (err) {
        const next = err instanceof Error ? err : new Error(String(err));
        setLocalError(next);
        return undefined;
      } finally {
        setIsFetching(false);
      }
    },
    [enabled, serialized, staleTime],
  );

  // Track generation so we re-run when cache is marked stale (updatedAt flips to 0).
  const generation = entry?.updatedAt ?? -1;

  useEffect(() => {
    if (!enabled) return;

    const current = cache.get(serialized);
    const stale =
      !current ||
      current.updatedAt === 0 ||
      Date.now() - current.updatedAt > staleTime;
    if (!stale) return;

    let cancelled = false;
    setIsFetching(true);
    setLocalError(null);
    runFetch(serialized, () => fetcherRef.current())
      .catch((err) => {
        if (!cancelled) {
          setLocalError(err instanceof Error ? err : new Error(String(err)));
        }
      })
      .finally(() => {
        if (!cancelled) setIsFetching(false);
      });

    return () => {
      cancelled = true;
    };
  }, [enabled, serialized, staleTime, generation]);

  return {
    data,
    error,
    loading: enabled && data === undefined,
    isFetching,
    isStale,
    refetch,
  };
}

/** Test helpers */
export function __adminQueryCacheSizeForTests(): number {
  return cache.size;
}
