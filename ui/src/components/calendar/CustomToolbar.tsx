import React from "react";
import Link from "next/link";
import { ToolbarProps } from "react-big-calendar";
import {
  Trello as IconTrello,
  ArrowLeft as IconArrowLeft,
  ArrowRight as IconArrowRight,
  SkipForward as IconSkipForward,
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

  const goToNextEvent = () => {
    if (!events || events.length === 0) {
      return;
    }
    const sorted = [...events].sort(
      (a, b) => a.enter.getTime() - b.enter.getTime(),
    );
    const currentDate = toolbar.date;
    const nextEvent = sorted.find(
      (event) => event.enter.getTime() > currentDate.getTime(),
    );
    const target = nextEvent ?? sorted[0];
    toolbar.onNavigate("DATE", target.enter);
  };

  return (
    <div
      className="custom-toolbar"
      style={{ marginBottom: "5px", textAlign: "left" }}
    >
      <Link
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={() => toolbar.onNavigate("TODAY")}
      >
        <IconTrello className="feather" /> {t("today")}
      </Link>{" "}
      <Link
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={() => toolbar.onNavigate("PREV")}
      >
        <IconArrowLeft className="feather" />
      </Link>{" "}
      <Link
        href="#"
        className="btn btn-sm btn-outline-secondary"
        onClick={() => toolbar.onNavigate("NEXT")}
      >
        <IconArrowRight className="feather" />
      </Link>{" "}
      {events && events.length > 0 && (
        <Link
          href="#"
          className="btn btn-sm btn-outline-secondary"
          onClick={() => goToNextEvent()}
        >
          <IconSkipForward className="feather" /> {t("nextEvent")}
        </Link>
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
