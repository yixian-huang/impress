import { useCallback, useEffect, useState, type MutableRefObject } from "react";
import type { NavigateFunction } from "react-router-dom";
import type { Article } from "@/api/articles";
import { createArticle } from "@/api/articles";
import {
  cancelScheduledPublication,
  createScheduledPublication,
  getResourceScheduledPublication,
  retryScheduledPublication,
  updateScheduledPublication,
  type ScheduledPublication,
} from "@/api/scheduledPublications";
import { slugifyTitle } from "../utils/slugify";
import type { ArticleSaveSource } from "./useArticlePersistence";

type BuildPayload = (status: "draft" | "published", publishedAt?: string) => Record<string, unknown>;

export function useArticleSchedule(opts: {
  id: string | undefined;
  isEditing: boolean;
  canPublish: boolean;
  sourceRef: MutableRefObject<ArticleSaveSource>;
  buildPayload: BuildPayload;
  navigate: NavigateFunction;
  setError: (e: string | null) => void;
}) {
  const { id, isEditing, canPublish, sourceRef, buildPayload, navigate, setError } = opts;

  const [scheduledPublication, setScheduledPublication] = useState<ScheduledPublication | null>(null);
  const [scheduleLoading, setScheduleLoading] = useState(!!id);
  const [scheduleBusy, setScheduleBusy] = useState(false);
  const [scheduleMessage, setScheduleMessage] = useState("");

  const loadArticleSchedule = useCallback(async () => {
    if (!id) {
      setScheduleLoading(false);
      return;
    }
    setScheduleLoading(true);
    try {
      const schedule = await getResourceScheduledPublication("article", Number(id));
      setScheduledPublication(schedule);
    } catch {
      setScheduledPublication(null);
    } finally {
      setScheduleLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void loadArticleSchedule();
  }, [loadArticleSchedule]);

  const handleSchedulePublish = useCallback(
    async (scheduledAt: string) => {
      if (!canPublish) return;
      const src = sourceRef.current;
      const fields = src.getFields();
      if (!fields.zhTitle.trim()) {
        setError("请填写中文标题");
        return;
      }
      const finalSlug = fields.slug.trim() || slugifyTitle(fields.zhTitle);
      if (!fields.slug.trim()) src.setSlug(finalSlug);
      setScheduleBusy(true);
      setError(null);
      setScheduleMessage("");
      try {
        const publishPayload = buildPayload("published", scheduledAt);
        if (scheduledPublication?.status === "pending") {
          const updated = await updateScheduledPublication(scheduledPublication.id, {
            scheduledAt,
            publishPayload,
          });
          setScheduledPublication(updated);
          setScheduleMessage("定时发布已更新");
        } else {
          let resourceId = Number(id);
          if (!isEditing) {
            const created = await createArticle(buildPayload("draft") as Partial<Article>);
            resourceId = created.id;
          }
          const createdSchedule = await createScheduledPublication({
            resourceType: "article",
            resourceId,
            scheduledAt,
            publishPayload,
          });
          setScheduledPublication(createdSchedule);
          const status = src.getArticleStatus();
          src.setArticleStatus(status === "published" ? "published" : "scheduled");
          setScheduleMessage("定时发布已安排");
          if (!isEditing) {
            navigate(`/admin/articles/edit/${resourceId}`, { replace: true });
          }
        }
      } catch (err: unknown) {
        const ax = err as { response?: { data?: { error?: { message?: string } | string } } };
        const msg = (ax?.response?.data?.error as { message?: string })?.message
          ?? (typeof ax?.response?.data?.error === "string" ? ax.response.data.error : undefined);
        setError(msg || (err instanceof Error ? err.message : "定时发布失败"));
      } finally {
        setScheduleBusy(false);
      }
    },
    [canPublish, sourceRef, buildPayload, scheduledPublication, id, isEditing, navigate, setError],
  );

  const handleCancelSchedule = useCallback(async () => {
    if (!scheduledPublication || !canPublish) return;
    setScheduleBusy(true);
    setError(null);
    setScheduleMessage("");
    try {
      await cancelScheduledPublication(scheduledPublication.id);
      setScheduledPublication(null);
      setScheduleMessage("定时发布已取消");
    } catch (err) {
      setError(err instanceof Error ? err.message : "取消定时发布失败");
    } finally {
      setScheduleBusy(false);
    }
  }, [scheduledPublication, canPublish, setError]);

  const handleRetrySchedule = useCallback(async () => {
    if (!scheduledPublication || !canPublish) return;
    setScheduleBusy(true);
    setError(null);
    setScheduleMessage("");
    try {
      const retried = await retryScheduledPublication(scheduledPublication.id);
      setScheduledPublication(retried);
      setScheduleMessage("定时发布已重新入队");
    } catch (err) {
      setError(err instanceof Error ? err.message : "重试定时发布失败");
    } finally {
      setScheduleBusy(false);
    }
  }, [scheduledPublication, canPublish, setError]);

  return {
    scheduledPublication,
    scheduleLoading,
    scheduleBusy,
    scheduleMessage,
    setScheduleMessage,
    loadArticleSchedule,
    handleSchedulePublish,
    handleCancelSchedule,
    handleRetrySchedule,
  };
}
