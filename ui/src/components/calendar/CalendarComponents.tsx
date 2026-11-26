import { Components } from "react-big-calendar";
import createCustomEvent, {
  createCustomMonthEvent,
  CalendarEvent,
} from "@/components/calendar/CustomEvent";
import CustomDayHeader from "@/components/calendar/CustomDayHeader";

export const getCalendarComponents = (
  toolbar: (props: object) => React.ReactNode,
): Components<CalendarEvent, object> =>
  ({
    toolbar,
    event: createCustomEvent(),
    week: { header: CustomDayHeader },
    day: { header: CustomDayHeader },
    month: { event: createCustomMonthEvent() },
  }) as Components<CalendarEvent, object>;
