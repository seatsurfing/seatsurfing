import { TranslationFunc } from "@/components/withTranslation";
import Booking from "@/types/Booking";
import RendererUtils from "@/util/RendererUtils";

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
  firstname: string;
  lastname: string;
  email: string;
  mode: CalendarEventMode;
};

export type CalendarEventMode = "user" | "space";

export const bookingToCalendarEvent = (
  booking: Booking,
  mode: CalendarEventMode,
  t: TranslationFunc,
): CalendarEvent => {
  let title: string;
  if (mode === "space") {
    title =
      RendererUtils.fullname(booking.user.firstname, booking.user.lastname) ||
      booking.user.email;
    if (booking.subject) {
      title = title ? `${title} – ${booking.subject}` : booking.subject;
    }
  } else {
    title = `${booking.space.location.name} (${booking.space.name})`;
    if (booking.subject) {
      title += `, ${booking.subject}`;
    }
  }
  const recurring = booking.isRecurring();
  if (recurring) {
    title += (title ? " " : "") + `(${t("recurring")})`;
  }

  return {
    title, // used as HTML tooltip
    enter: booking.enter,
    leave: booking.leave,
    approved: booking.approved,
    recurring,
    subject: booking.subject,
    spaceName: booking.space.name,
    locationName: booking.space.location.name,
    bookingId: booking.id,
    firstname: booking.user.firstname,
    lastname: booking.user.lastname,
    email: booking.user.email,
    mode,
  };
};
