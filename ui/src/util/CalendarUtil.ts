import { DateRange, Formats, momentLocalizer, View } from "react-big-calendar";
import moment from "moment-timezone";
import RuntimeConfig from "@/components/RuntimeConfig";
import Formatting from "@/util/Formatting";
import DateUtil from "@/util/DateUtil";
import type { CalendarEvent } from "@/components/calendar/CalendarEvent";

export default class CalendarUtil {
  static setupMoment() {
    moment.tz.setDefault("UTC");
    moment.locale(Formatting.Language);
    const dow = RuntimeConfig.INFOS.weekStartDay;
    if (moment.localeData().firstDayOfWeek() !== dow) {
      moment.updateLocale(moment.locale(), {
        week: { dow },
      });
    }
  }

  static getLocalizer() {
    CalendarUtil.setupMoment();
    return momentLocalizer(moment);
  }

  static intl(options: Intl.DateTimeFormatOptions): Intl.DateTimeFormat {
    return new Intl.DateTimeFormat(Formatting.Language, {
      timeZone: "UTC",
      ...options,
    });
  }

  static getTimeFormat(): string {
    return RuntimeConfig.INFOS.use24HourTime ? "HH:mm" : "h:mm A";
  }

  static formatTime(date: Date): string {
    CalendarUtil.setupMoment();
    return moment(date).format(CalendarUtil.getTimeFormat());
  }

  static formatWeekdayShort(date: Date): string {
    return CalendarUtil.intl({ weekday: "short" }).format(date);
  }

  static formatDayOfMonth(date: Date): string {
    return CalendarUtil.intl({ day: "numeric" }).format(date);
  }

  static getPeriodLabel(view: View, date: Date): string {
    if (view === "day") {
      return CalendarUtil.intl({
        weekday: "long",
        day: "numeric",
        month: "long",
        year: "numeric",
      }).format(date);
    }
    if (view === "month") {
      return CalendarUtil.intl({ month: "long", year: "numeric" }).format(date);
    }
    const range = CalendarUtil.getWeekRange(date);
    return CalendarUtil.intl({
      day: "numeric",
      month: "short",
      year: "numeric",
    }).formatRange(range.start, range.end);
  }

  static getFormats(): Formats {
    const time = CalendarUtil.getTimeFormat();
    const range = ({ start, end }: DateRange, culture?: string, local?: any) =>
      local.format(start, time, culture) +
      " – " +
      local.format(end, time, culture);
    const intl = (options: Intl.DateTimeFormatOptions) => {
      const formatter = CalendarUtil.intl(options);
      return (date: Date) => formatter.format(date);
    };
    const weekday = intl({ weekday: "short" });
    const day = intl({ day: "2-digit" });
    return {
      dateFormat: intl({ day: "numeric" }),
      weekdayFormat: weekday,
      dayFormat: (date: Date) => day(date) + " " + weekday(date),
      dayHeaderFormat: intl({
        weekday: "long",
        month: "short",
        day: "2-digit",
      }),
      monthHeaderFormat: intl({ month: "long", year: "numeric" }),
      timeGutterFormat: time,
      agendaTimeFormat: time,
      selectRangeFormat: range,
      eventTimeRangeFormat: range,
      agendaTimeRangeFormat: range,
      eventTimeRangeStartFormat: ({ start }, culture, local: any) =>
        local.format(start, time, culture) + " – ",
      eventTimeRangeEndFormat: ({ end }, culture, local: any) =>
        " – " + local.format(end, time, culture),
    };
  }

  static getWeekRange(date: Date): { start: Date; end: Date } {
    CalendarUtil.setupMoment();
    return {
      start: moment(date).startOf("week").toDate(),
      end: moment(date).endOf("week").toDate(),
    };
  }

  static getMonthRange(date: Date): { start: Date; end: Date } {
    CalendarUtil.setupMoment();
    return {
      start: moment(date).startOf("month").startOf("week").toDate(),
      end: moment(date).endOf("month").endOf("week").toDate(),
    };
  }

  static getRange(view: View, date: Date): { start: Date; end: Date } {
    return view === "month"
      ? CalendarUtil.getMonthRange(date)
      : CalendarUtil.getWeekRange(date);
  }

  static getNow = (): Date => DateUtil.getNowFakeUTC();

  static startAccessor = (event: CalendarEvent): Date => event.enter;

  static endAccessor = (event: CalendarEvent): Date => event.leave;

  static eventPropGetter = (event: CalendarEvent) => ({
    className:
      "calendar-event" +
      (event.approved === false ? " calendar-event-pending" : ""),
  });

  static getDayPropGetter = (allowedDays: number[]) => {
    return (date: Date) => {
      if (allowedDays.length > 0 && !allowedDays.includes(date.getUTCDay())) {
        return { className: "calendar-day-off" };
      }
      return {};
    };
  };
}
