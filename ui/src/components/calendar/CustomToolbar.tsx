import React from "react";
import { ToolbarProps } from "react-big-calendar";
import {
  Trello as IconTrello,
  ArrowLeft as IconArrowLeft,
  ArrowRight as IconArrowRight,
  SkipForward as IconSkipForward,
  SkipBack as IconSkipBack,
} from "react-feather";
import Formatting from "@/util/Formatting";
import moment from "moment-timezone";
import { TranslationFunc } from "@/components/withTranslation";
import { CalendarEvent } from "@/components/calendar/CalendarEvent";

interface Props {
  toolbar: ToolbarProps<object, object>;
  t: TranslationFunc;
  events?: CalendarEvent[];
}

const CustomToolbar: React.FC<Props> = ({ toolbar, t, events }) => {
  const weekStart = moment(toolbar.date).clone().startOf("week");
  const weekEnd = moment(toolbar.date).clone().endOf("week");
  const formatter = Formatting.getFormatterDate();
  const isDayView = toolbar.view === "day";

  const eventWeeks = (events ?? []).reduce((map, event) => {
    const weekKey = moment(event.enter).startOf("week").valueOf();
    const existing = map.get(weekKey);
    if (!existing || event.enter.getTime() < existing.getTime()) {
      map.set(weekKey, event.enter);
    }
    return map;
  }, new Map<number, Date>());
  const sortedEventWeeks = Array.from(eventWeeks.entries()).sort(
    (a, b) => a[0] - b[0],
  );
  const currentWeekStartMs = weekStart.valueOf();
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

  return (
    <div
      className="custom-toolbar"
      style={{ marginBottom: "5px", textAlign: "left" }}
    >
      <button
        type="button"
        className="btn btn-sm btn-outline-secondary"
        onClick={() => toolbar.onNavigate("TODAY")}
      >
        <IconTrello className="feather" /> {t("today")}
      </button>{" "}
      <button
        type="button"
        className="btn btn-sm btn-outline-secondary"
        onClick={() => toolbar.onNavigate("PREV")}
      >
        <IconArrowLeft className="feather" />
      </button>{" "}
      <button
        type="button"
        className="btn btn-sm btn-outline-secondary"
        onClick={() => toolbar.onNavigate("NEXT")}
      >
        <IconArrowRight className="feather" />
      </button>{" "}
      {events && events.length > 0 && (
        <>
          <button
            type="button"
            className="btn btn-sm btn-outline-secondary"
            disabled={!prevEventWeek}
            onClick={() => goToPreviousEvent()}
          >
            <IconSkipBack className="feather" /> {t("previousBooking")}
          </button>{" "}
          <button
            type="button"
            className="btn btn-sm btn-outline-secondary"
            disabled={!nextEventWeek}
            onClick={() => goToNextEvent()}
          >
            <IconSkipForward className="feather" /> {t("nextBooking")}
          </button>
        </>
      )}{" "}
      <span
        className="toolbar-label"
        style={{
          display: "flex",
          float: "right",
          height: "100%",
          alignItems: "center",
        }}
      >
        {isDayView
          ? formatter.format(toolbar.date)
          : `${formatter.format(weekStart.toDate())} – ${formatter.format(weekEnd.toDate())}`}
      </span>
    </div>
  );
};

export default CustomToolbar;
