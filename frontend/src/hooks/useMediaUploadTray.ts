import { useCallback, useEffect, useReducer, useRef } from "react";
import {
  emitMediaUpload,
  subscribeMediaUpload,
  type MediaUploadEvent,
} from "@/lib/mediaUploadTracked";

export type UploadTrayItem = {
  id: string;
  name: string;
  size: number;
  percent: number;
  status: "uploading" | "success" | "error";
  message?: string;
};

const SUCCESS_DISMISS_MS = 2200;
const MAX_ITEMS = 8;

type TrayState = UploadTrayItem[];

type TrayAction =
  | { type: "upsert"; item: UploadTrayItem }
  | { type: "patch"; id: string; patch: Partial<UploadTrayItem> }
  | { type: "remove"; id: string };

function trayReducer(state: TrayState, action: TrayAction): TrayState {
  switch (action.type) {
    case "upsert": {
      const rest = state.filter((i) => i.id !== action.item.id);
      return [action.item, ...rest].slice(0, MAX_ITEMS);
    }
    case "patch":
      return state.map((i) => (i.id === action.id ? { ...i, ...action.patch } : i));
    case "remove":
      return state.filter((i) => i.id !== action.id);
    default:
      return state;
  }
}

/**
 * Subscribe to media upload bus and keep a small tray of active/failed jobs.
 * Retry callbacks live in a ref map so React state stays serializable/plain.
 */
export function useMediaUploadTray() {
  const [items, dispatch] = useReducer(trayReducer, []);
  const retryFns = useRef(new Map<string, () => void>());
  const dismissTimers = useRef(new Map<string, number>());

  const clearTimer = useCallback((id: string) => {
    const t = dismissTimers.current.get(id);
    if (t != null) {
      window.clearTimeout(t);
      dismissTimers.current.delete(id);
    }
  }, []);

  useEffect(() => {
    const onEvent = (event: MediaUploadEvent) => {
      if (event.type === "start") {
        clearTimer(event.id);
        retryFns.current.delete(event.id);
        dispatch({
          type: "upsert",
          item: {
            id: event.id,
            name: event.name,
            size: event.size,
            percent: 0,
            status: "uploading",
          },
        });
        return;
      }

      if (event.type === "progress") {
        dispatch({
          type: "patch",
          id: event.id,
          patch: { percent: event.percent, status: "uploading" },
        });
        return;
      }

      if (event.type === "success") {
        retryFns.current.delete(event.id);
        dispatch({
          type: "patch",
          id: event.id,
          patch: { percent: 100, status: "success", message: undefined },
        });
        clearTimer(event.id);
        const t = window.setTimeout(() => {
          dispatch({ type: "remove", id: event.id });
          dismissTimers.current.delete(event.id);
        }, SUCCESS_DISMISS_MS);
        dismissTimers.current.set(event.id, t);
        return;
      }

      if (event.type === "error") {
        clearTimer(event.id);
        if (event.retry) retryFns.current.set(event.id, event.retry);
        else retryFns.current.delete(event.id);
        dispatch({
          type: "upsert",
          item: {
            id: event.id,
            name: event.name || "图片",
            size: 0,
            percent: 0,
            status: "error",
            message: event.message,
          },
        });
        return;
      }

      if (event.type === "dismiss") {
        clearTimer(event.id);
        retryFns.current.delete(event.id);
        dispatch({ type: "remove", id: event.id });
      }
    };

    return subscribeMediaUpload(onEvent);
  }, [clearTimer]);

  useEffect(() => {
    const timers = dismissTimers.current;
    const retries = retryFns.current;
    return () => {
      timers.forEach((t) => window.clearTimeout(t));
      timers.clear();
      retries.clear();
    };
  }, []);

  const dismiss = useCallback((id: string) => {
    emitMediaUpload({ type: "dismiss", id });
  }, []);

  const retry = useCallback((id: string) => {
    const fn = retryFns.current.get(id);
    fn?.();
  }, []);

  const canRetry = useCallback((id: string) => retryFns.current.has(id), []);

  return { items, dismiss, retry, canRetry };
}
