import { useEffect, useRef, useState } from "react";
import {
  MapPin as IconLocation,
  Clock as IconPending,
  RefreshCw as IconRecurring,
} from "react-feather";
import { TranslationFunc } from "@/components/withTranslation";
import Booking from "@/types/Booking";

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

    let pending = <></>;
    if (showDetails && event.booking.approved === false) {
      pending = (
        <>
          <IconPending
            className="feather"
            style={{ width: "12px", height: "12px" }}
          />
          &nbsp;{t("approval")}: {t("pending")}
          <br />
        </>
      );
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
        {pending}
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
