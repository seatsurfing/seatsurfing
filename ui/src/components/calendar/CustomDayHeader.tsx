import { HeaderProps } from "react-big-calendar";
import CalendarUtil from "@/util/CalendarUtil";

const CustomDayHeader = ({ date }: HeaderProps) => (
  <span className="calendar-day-header">
    <span className="calendar-day-header-weekday">
      {CalendarUtil.formatWeekdayShort(date)}
    </span>
    <span className="calendar-day-header-day">
      {CalendarUtil.formatDayOfMonth(date)}
    </span>
  </span>
);

export default CustomDayHeader;
