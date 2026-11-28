import React from "react";
import { ToolbarProps } from "react-big-calendar";
import {
  ChevronLeft as IconChevronLeft,
  ChevronRight as IconChevronRight,
  SkipForward as IconSkipForward,
  SkipBack as IconSkipBack,
  Globe as IconGlobe,
} from "react-feather";
import CalendarUtil from "@/util/CalendarUtil";
import moment from "moment-timezone";
import { TranslationFunc } from "@/components/withTranslation";
import { CalendarEvent } from "@/components/calendar/CalendarEvent";

interface Props {
  toolbar: ToolbarProps<object, object>;
  t: TranslationFunc;
  events?: CalendarEvent[];
  timezone?: string;
  minDate?: Date;
}

const CustomToolbar: React.FC<Props> = ({
  toolbar,
  t,
  events,
  timezone,
  minDate,
}) => {
  const isMonthView = toolbar.view === "month";
  const periodUnit = isMonthView ? "month" : "week";
  const periodStart = moment(toolbar.date).clone().startOf(periodUnit);
  const periodEnd = moment(toolbar.date).clone().endOf(periodUnit);
  const canNavigateBack = !minDate || periodStart.isAfter(minDate);
  const availableViews = Array.isArray(toolbar.views)
    ? toolbar.views
    : Object.keys(toolbar.views ?? {});
  const canToggleMonth =
    availableViews.includes("week") && availableViews.includes("month");

  const eventWeeks = (events ?? []).reduce((map, event) => {
    const weekKey = moment(event.enter).startOf(periodUnit).valueOf();
    const existing = map.get(weekKey);
    if (!existing || event.enter.getTime() < existing.getTime()) {
      map.set(weekKey, event.enter);
    }
    return map;
  }, new Map<number, Date>());
  const sortedEventWeeks = Array.from(eventWeeks.entries()).sort(
    (a, b) => a[0] - b[0],
  );
  const currentWeekStartMs = periodStart.valueOf();
  const nextEventWeek = sortedEventWeeks.find(
    ([weekMs]) => weekMs > currentWeekStartMs,
  );
  const prevEventWeek = [...sortedEventWeeks]
    .reverse()
    .find(([weekMs]) => weekMs < currentWeekStartMs);
  const goToNextEvent = () => {
    if (!nextEventWeek) {
      return;
    }
    toolbar.onNavigate("DATE", nextEventWeek[1]);
  };

  const goToPreviousEvent = () => {
    if (!prevEventWeek) {
      return;
    }
    toolbar.onNavigate("DATE", prevEventWeek[1]);
  };

  const label = CalendarUtil.getPeriodLabel(toolbar.view, toolbar.date);

  return (
    <div className="custom-toolbar">
      <div className="toolbar-group">
        <button
          type="button"
          className="btn btn-sm btn-outline-secondary toolbar-today"
          onClick={() => toolbar.onNavigate("TODAY")}
        >
          {t("today")}
        </button>
        <div className="btn-group btn-group-sm toolbar-nav" role="group">
          <button
            type="button"
            className="btn btn-outline-secondary"
            disabled={!canNavigateBack}
            onClick={() => toolbar.onNavigate("PREV")}
            aria-label={t("previous")}
            title={t("previous")}
          >
            <IconChevronLeft className="feather" />
          </button>
          <button
            type="button"
            className="btn btn-outline-secondary"
            onClick={() => toolbar.onNavigate("NEXT")}
            aria-label={t("next")}
            title={t("next")}
          >
            <IconChevronRight className="feather" />
          </button>
        </div>
        <h2 className="toolbar-label" aria-live="polite">
          {label}
        </h2>
      </div>
      <div className="toolbar-group">
        {timezone && (
          <span
            className="toolbar-timezone d-none d-md-inline-flex"
            title={t("timezone")}
            data-testid="calendar-timezone"
          >
            <IconGlobe className="feather" />
            {timezone}
          </span>
        )}
        {events && events.length > 0 && (
          <div className="btn-group btn-group-sm" role="group">
            <button
              type="button"
              className="btn btn-outline-secondary"
              disabled={!prevEventWeek}
              onClick={() => goToPreviousEvent()}
            >
              <IconSkipBack className="feather" /> {t("previousBooking")}
            </button>
            <button
              type="button"
              className="btn btn-outline-secondary"
              disabled={!nextEventWeek}
              onClick={() => goToNextEvent()}
            >
              {t("nextBooking")} <IconSkipForward className="feather" />
            </button>
          </div>
        )}
        {canToggleMonth && (
          <div className="toolbar-view-switch" role="group">
            <button
              type="button"
              className={isMonthView ? "" : "active"}
              aria-pressed={!isMonthView}
              onClick={() => toolbar.onView("week")}
            >
              {t("week")}
            </button>
            <button
              type="button"
              className={isMonthView ? "active" : ""}
              aria-pressed={isMonthView}
              onClick={() => toolbar.onView("month")}
            >
              {t("month")}
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default CustomToolbar;
