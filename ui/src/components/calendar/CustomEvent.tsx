import { useEffect, useRef, useState } from "react";
import {
  MapPin as IconLocation,
  RefreshCw as IconRecurring,
} from "react-feather";
import { TranslationFunc } from "@/components/withTranslation";
import Booking from "@/types/Booking";

export const bookingToCalendarEvent = (
  booking: Booking,
  t: TranslationFunc,
): CalendarEvent => {
  let title = `${booking.space.location.name} (${booking.space.name})`;
  if (booking.subject) {
    title += `, ${booking.subject}`;
  }
  if (booking.isRecurring()) {
    title += ` (${t("recurring")})`;
  }

  return {
    title, // used in tooltip
    booking,
  };
};

export const bookingToSpaceCalendarEvent = (booking: Booking): CalendarEvent => ({
  title: booking.user.email + (booking.subject ? ` – ${booking.subject}` : ""),
  booking,
});

export type CalendarEvent = {
  title: string;
  booking: Booking;
};

const WIDTH_THRESHOLD = 100;

const createCustomEvent =
  (t: TranslationFunc) =>
  ({ event }: { event: CalendarEvent }) => {
    const containerRef = useRef<HTMLDivElement>(null);
    const [showDetails, setShowDetails] = useState(true);

    useEffect(() => {
      const el = containerRef.current;
      if (!el) return;
      const observer = new ResizeObserver((entries) => {
        setShowDetails(entries[0].contentRect.width >= WIDTH_THRESHOLD);
      });
      observer.observe(el);
      return () => observer.disconnect();
    }, []);

    // show no information for events < 1 hr
    if (
      event.booking.leave.getTime() - event.booking.enter.getTime() <=
      60 * 60 * 1000
    ) {
      return null;
    }

    let recurringIcon = <></>;
    if (event.booking.isRecurring()) {
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
        <p hidden={!event.booking.subject}>
          <strong>{event.booking.subject}</strong>
        </p>
        {showDetails && (
          <>
            <IconLocation
              className="feather"
              style={{ width: "12px", height: "12px" }}
            />{" "}
            {event.booking.space.location.name}, {event.booking.space.name}
          </>
        )}
      </div>
    );
  };

export default createCustomEvent;
