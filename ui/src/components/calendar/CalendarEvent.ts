import { TranslationFunc } from "@/components/withTranslation";
import Booking from "@/types/Booking";

export type CalendarEvent = {
  title: string;
  enter: Date;
  leave: Date;
  approved: boolean;
  recurring: boolean;
  subject: string;
  spaceName: string;
  locationName: string;
  bookingId: string;
};

export type CalendarEventMode = "user" | "space";

export const bookingToCalendarEvent = (
  booking: Booking,
  mode: CalendarEventMode,
  t?: TranslationFunc,
): CalendarEvent => {
  let title: string;
  if (mode === "space") {
    title =
      booking.user.email + (booking.subject ? ` – ${booking.subject}` : "");
  } else {
    title = `${booking.space.location.name} (${booking.space.name})`;
    if (booking.subject) {
      title += `, ${booking.subject}`;
    }
    if (booking.isRecurring() && t) {
      title += ` (${t("recurring")})`;
    }
  }

  return {
    title,
    enter: booking.enter,
    leave: booking.leave,
    approved: booking.approved,
    recurring: booking.isRecurring(),
    subject: booking.subject,
    spaceName: booking.space.name,
    locationName: booking.space.location.name,
    bookingId: booking.id,
  };
};
