import { useEffect, useState } from "react";
import { fetchSetupStatus, type SetupStatus } from "@/api/setup";

let cachedStatus: SetupStatus | null = null;
let inflight: Promise<SetupStatus> | null = null;

async function loadSetupStatus(): Promise<SetupStatus> {
  if (cachedStatus) {
    return cachedStatus;
  }
  if (!inflight) {
    inflight = fetchSetupStatus()
      .then((status) => {
        cachedStatus = status;
        return status;
      })
      .finally(() => {
        inflight = null;
      });
  }
  return inflight;
}

export function clearSetupStatusCache() {
  cachedStatus = null;
}

export function useSetupStatus() {
  const [status, setStatus] = useState<SetupStatus | null>(cachedStatus);
  const [loading, setLoading] = useState(!cachedStatus);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    loadSetupStatus()
      .then((next) => {
        if (!cancelled) {
          setStatus(next);
          setLoading(false);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load setup status");
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return { status, loading, error };
}
