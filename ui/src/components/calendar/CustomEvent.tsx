import { useEffect, useRef, useState } from "react";
import {
  Clock as IconPending,
  MapPin as IconLocation,
  RefreshCw as IconRecurring,
} from "react-feather";
export { bookingToCalendarEvent } from "./CalendarEvent";
import type { CalendarEvent } from "./CalendarEvent";
import RendererUtils from "@/util/RendererUtils";
import CalendarUtil from "@/util/CalendarUtil";
export type { CalendarEvent };

const WIDTH_THRESHOLD = 100;

const EventIcons = ({ event }: { event: CalendarEvent }) => {
  if (!event.recurring && event.approved !== false) {
    return null;
  }
  return (
    <span className="calendar-event-icons">
      {event.approved === false && <IconPending className="feather" />}
      {event.recurring && <IconRecurring className="feather" />}
    </span>
  );
};

const createCustomEvent =
  () =>
  ({ event }: { event: CalendarEvent }) => {
    // show no information for events < 1 hr
    if (event.leave.getTime() - event.enter.getTime() <= 60 * 60 * 1000) {
      return null;
    }

    const containerRef = useRef<HTMLDivElement>(null);
    const [showDetails, setShowDetails] = useState(true);

    // dynamically show/hide information based on calender entry width
    useEffect(() => {
      const el = containerRef.current;
      if (!el) return;
      const observer = new ResizeObserver((entries) => {
        setShowDetails(entries[0].contentRect.width >= WIDTH_THRESHOLD);
      });
      observer.observe(el);
      return () => observer.disconnect();
    }, []);

    const person =
      event.mode == "space" &&
      (event.firstname || event.lastname || event.email)
        ? RendererUtils.fullname(event.firstname, event.lastname, event.email)
        : "";

    return (
      <div ref={containerRef} className="calendar-event-body">
        <EventIcons event={event} />
        {event.subject && (
          <div className="calendar-event-subject">{event.subject}</div>
        )}
        {showDetails && event.mode == "user" && (
          <div className="calendar-event-detail">
            <IconLocation className="feather" />
            <span>
              {event.locationName}, {event.spaceName}
            </span>
          </div>
        )}
        {showDetails && person && (
          <div className="calendar-event-detail">{person}</div>
        )}
      </div>
    );
  };

export const createCustomMonthEvent =
  () =>
  ({ event }: { event: CalendarEvent }) => (
    <div className="calendar-month-event" title={event.title}>
      <span className="calendar-month-event-time">
        {CalendarUtil.formatTime(event.enter)}
      </span>
      <span className="calendar-month-event-title">{event.title}</span>
      <EventIcons event={event} />
    </div>
  );

export default createCustomEvent;
