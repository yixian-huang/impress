import { useCallback, useEffect, useState } from "react";
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
  retry?: () => void;
};

const SUCCESS_DISMISS_MS = 2200;

/**
 * Subscribe to media upload bus and keep a small tray of active/failed jobs.
 */
export function useMediaUploadTray() {
  const [items, setItems] = useState<UploadTrayItem[]>([]);

  useEffect(() => {
    const timers = new Map<string, number>();

    const onEvent = (event: MediaUploadEvent) => {
      if (event.type === "start") {
        setItems((prev) => {
          const rest = prev.filter((i) => i.id !== event.id);
          return [
            {
              id: event.id,
              name: event.name,
              size: event.size,
              percent: 0,
              status: "uploading" as const,
            },
            ...rest,
          ].slice(0, 8);
        });
        return;
      }
      if (event.type === "progress") {
        setItems((prev) =>
          prev.map((i) =>
            i.id === event.id ? { ...i, percent: event.percent, status: "uploading" } : i,
          ),
        );
        return;
      }
      if (event.type === "success") {
        setItems((prev) =>
          prev.map((i) =>
            i.id === event.id
              ? { ...i, percent: 100, status: "success", message: undefined, retry: undefined }
              : i,
          ),
        );
        const t = window.setTimeout(() => {
          setItems((prev) => prev.filter((i) => i.id !== event.id));
          timers.delete(event.id);
        }, SUCCESS_DISMISS_MS);
        timers.set(event.id, t);
        return;
      }
      if (event.type === "error") {
        setItems((prev) => {
          const existing = prev.find((i) => i.id === event.id);
          if (existing) {
            return prev.map((i) =>
              i.id === event.id
                ? {
                    ...i,
                    status: "error" as const,
                    message: event.message,
                    retry: event.retry,
                  }
                : i,
            );
          }
          return [
            {
              id: event.id,
              name: "图片",
              size: 0,
              percent: 0,
              status: "error" as const,
              message: event.message,
              retry: event.retry,
            },
            ...prev,
          ].slice(0, 8);
        });
        return;
      }
      if (event.type === "dismiss") {
        setItems((prev) => prev.filter((i) => i.id !== event.id));
      }
    };

    const unsub = subscribeMediaUpload(onEvent);
    return () => {
      unsub();
      timers.forEach((t) => window.clearTimeout(t));
      timers.clear();
    };
  }, []);

  const dismiss = useCallback((id: string) => {
    emitMediaUpload({ type: "dismiss", id });
    setItems((prev) => prev.filter((i) => i.id !== id));
  }, []);

  const retry = useCallback((id: string) => {
    setItems((prev) => {
      const item = prev.find((i) => i.id === id);
      if (item?.retry) {
        item.retry();
      }
      return prev;
    });
  }, []);

  return { items, dismiss, retry };
}
