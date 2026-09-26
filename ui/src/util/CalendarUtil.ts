import { momentLocalizer, View } from "react-big-calendar";
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

  static eventPropGetter = (event: CalendarEvent) => {
    return event.approved === false ? { style: { opacity: 0.5 } } : {};
  };

  static getDayPropGetter = (allowedDays: number[]) => {
    return (date: Date) => {
      if (allowedDays.length > 0 && !allowedDays.includes(date.getUTCDay())) {
        return {
          style: { backgroundColor: "rgba(0, 0, 0, 0.05)" },
        };
      }
      return {};
    };
  };
}
