import React from "react";
import { Calendar, momentLocalizer, View } from "react-big-calendar";
import moment from "moment-timezone";
import "react-big-calendar/lib/css/react-big-calendar.css";
import CustomToolbar from "@/components/calendar/CustomToolbar";
import createCustomEvent, {
  CalendarEvent,
} from "@/components/calendar/CustomEvent";
import { TranslationFunc } from "@/components/withTranslation";
import RuntimeConfig from "@/components/RuntimeConfig";
import DateUtil from "@/util/DateUtil";
import Formatting from "@/util/Formatting";

interface Props {
  t: TranslationFunc;
  events: CalendarEvent[];
  date: Date;
  onNavigate: (date: Date, view: View) => void;
  view?: View;
  views?: View[];
  onView?: (view: View) => void;
  onSelectEvent?: (event: CalendarEvent) => void;
  scrollToTime?: Date;
  workdays?: number[];
  height?: string;
  showBookingNavigation?: boolean;
}

const setupMoment = () => {
  moment.tz.setDefault("UTC");
  moment.locale(Formatting.Language);
  const dow = RuntimeConfig.INFOS.weekStartDay;
  if (moment.localeData().firstDayOfWeek() !== dow) {
    moment.updateLocale(moment.locale(), {
      week: { dow },
    });
  }
};

export const getCalendarLocalizer = () => {
  setupMoment();
  return momentLocalizer(moment);
};

// Returns the visible range of the given view in fake UTC, including the
// leading and trailing days of adjacent months shown in the month view.
export const getCalendarRange = (
  date: Date,
  view: View,
): { start: Date; end: Date } => {
  setupMoment();
  const unit = view === "month" ? "month" : "week";
  return {
    start: moment(date).startOf(unit).startOf("week").toDate(),
    end: moment(date).endOf(unit).endOf("week").toDate(),
  };
};

const BookingCalendar: React.FC<Props> = (props) => {
  const views = props.views ?? ["week"];
  const toolbar = (toolbarProps: object) => (
    <CustomToolbar
      toolbar={toolbarProps as any}
      t={props.t}
      events={props.showBookingNavigation === false ? undefined : props.events}
      views={views}
    />
  );
  const workdays = props.workdays ?? [];

  return (
    <Calendar
      showMultiDayTimes={true}
      getNow={() => DateUtil.getNowFakeUTC()}
      localizer={getCalendarLocalizer()}
      events={props.events}
      startAccessor={(event: CalendarEvent) => event.enter}
      endAccessor={(event: CalendarEvent) => event.leave}
      style={{
        height: props.height ?? "calc(100vh - 160px)",
        width: "100%",
        padding: "10px",
        margin: "auto",
      }}
      defaultView="week"
      view={props.view}
      onView={props.onView ?? (() => {})}
      date={props.date}
      onNavigate={props.onNavigate}
      onSelectEvent={props.onSelectEvent}
      culture={Formatting.Language}
      length={7}
      views={views}
      eventPropGetter={(event: CalendarEvent) => {
        if (event.approved === false) {
          return { style: { opacity: 0.5 } };
        }
        return {};
      }}
      components={{
        toolbar,
        event: createCustomEvent(),
      }}
      scrollToTime={props.scrollToTime}
      dayPropGetter={(date: Date) => {
        if (workdays.length > 0 && !workdays.includes(date.getUTCDay())) {
          return {
            style: { backgroundColor: "rgba(0, 0, 0, 0.05)" },
          };
        }
        return {};
      }}
    />
  );
};

export default BookingCalendar;
