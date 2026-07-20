import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { futureDateTimeLocalValue } from "@/api/scheduledPublications";
import { AdminButton, AdminField, AdminInput, AdminModal } from "@/components/admin/ui";

interface SchedulePublishModalProps {
  open: boolean;
  onClose: () => void;
  onSchedule: (date: string) => void | Promise<void>;
  currentSchedule?: string | null;
  submitting?: boolean;
}

export default function SchedulePublishModal({
  open,
  onClose,
  onSchedule,
  currentSchedule,
  submitting = false,
}: SchedulePublishModalProps) {
  const [date, setDate] = useState(currentSchedule ?? "");
  const { t } = useTranslation();

  useEffect(() => {
    if (open) {
      setDate(currentSchedule ?? "");
    }
  }, [currentSchedule, open]);

  const minDate = futureDateTimeLocalValue();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (date) onSchedule(date);
  };

  return (
    <AdminModal
      open={open}
      title={t("schedule.title", "Schedule Publish")}
      onClose={onClose}
      widthClass="max-w-sm"
      footer={
        <>
          <AdminButton type="button" variant="secondary" size="sm" onClick={onClose} disabled={submitting}>
            {t("common.cancel", "Cancel")}
          </AdminButton>
          {currentSchedule ? (
            <AdminButton
              type="button"
              variant="ghost"
              size="sm"
              className="text-red-600 hover:bg-red-50"
              onClick={() => onSchedule("")}
              disabled={submitting}
            >
              {t("schedule.cancel_schedule", "Cancel Schedule")}
            </AdminButton>
          ) : null}
          <AdminButton
            type="submit"
            form="schedule-publish-form"
            size="sm"
            disabled={submitting}
          >
            {submitting ? t("common.saving", "Saving...") : t("schedule.confirm", "Schedule")}
          </AdminButton>
        </>
      }
    >
      <form id="schedule-publish-form" onSubmit={handleSubmit}>
        <AdminField label={t("schedule.datetime", "Publish Date & Time")}>
          <AdminInput
            type="datetime-local"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            min={minDate}
            disabled={submitting}
            required
          />
        </AdminField>
      </form>
    </AdminModal>
  );
}
