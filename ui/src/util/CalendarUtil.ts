import { momentLocalizer, View } from "react-big-calendar";
import moment from "moment-timezone";
import RuntimeConfig from "@/components/RuntimeConfig";
import Formatting from "@/util/Formatting";

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

  // Returns the visible range of the given view in fake UTC, including the
  // leading and trailing days of adjacent months shown in the month view.
  static getRange(date: Date, view: View): { start: Date; end: Date } {
    CalendarUtil.setupMoment();
    const unit = view === "month" ? "month" : "week";
    return {
      start: moment(date).startOf(unit).startOf("week").toDate(),
      end: moment(date).endOf(unit).endOf("week").toDate(),
    };
  }
}
