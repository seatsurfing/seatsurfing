import { useEffect, useRef, useState } from "react";
import {
  MapPin as IconLocation,
  RefreshCw as IconRecurring,
} from "react-feather";
export { bookingToCalendarEvent } from "./CalendarEvent";
import type { CalendarEvent } from "./CalendarEvent";
export type { CalendarEvent };

const WIDTH_THRESHOLD = 100;

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

    let recurringIcon = <></>;
    if (event.recurring) {
      recurringIcon = (
        <IconRecurring
          className="feather recurring-booking-icon"
          style={{ width: "12px", height: "12px", top: "4px", right: "4px" }}
        />
      );
    }

    return (
      <div ref={containerRef} style={{ fontSize: "12px" }}>
        {recurringIcon}
        <p hidden={!event.subject}>
          <strong>{event.subject}</strong>
        </p>
        {showDetails && event.mode == "user" && (
          <>
            <IconLocation
              className="feather"
              style={{ width: "12px", height: "12px" }}
            />{" "}
            {event.locationName}, {event.spaceName}
          </>
        )}
        {showDetails && event.mode == "space" && event.email && (
          <>{event.email}</>
        )}
      </div>
    );
  };

export default createCustomEvent;
