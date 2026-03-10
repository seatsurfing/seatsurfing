import RuntimeConfig from "@/components/RuntimeConfig";
import { TranslationFunc } from "@/components/withTranslation";

const enum ResponseCode {
  BookingSlotConflict = 1001,
  BookingLocationMaxConcurrent = 1002,
  BookingTooManyUpcomingBookings = 1003,
  BookingTooManyDaysInAdvance = 1004,
  BookingInvalidMaxBookingDuration = 1005,
  BookingMaxConcurrentForUser = 1006,
  BookingInvalidMinBookingDuration = 1007,
  BookingMaxHoursBeforeDelete = 1008,
  BookingInPast = 1011,
  PresenceReportDateRangeTooLong = 2001,
}

const ResponseCodeUsernameExists: number = 3001;

export default class ErrorText {
  static getTextForAppCode(code: number, t: TranslationFunc): string {
    if (code === ResponseCodeBookingSlotConflict) {
      return t("errorSlotConflict");
    } else if (code === ResponseCodeBookingLocationMaxConcurrent) {
      return t("errorTooManyConcurrent");
    } else if (code === ResponseCodeBookingInvalidMaxBookingDuration) {
      return t("errorMaxBookingDuration", {
        num: RuntimeConfig.INFOS.maxBookingDurationHours,
      });
    } else if (code === ResponseCodeBookingInvalidMinBookingDuration) {
      return t("errorMinBookingDuration", {
        num: RuntimeConfig.INFOS.minBookingDurationHours,
      });
    } else if (code === ResponseCodeBookingTooManyDaysInAdvance) {
      return t("errorDaysAdvance", {
        num: RuntimeConfig.INFOS.maxDaysInAdvance,
      });
    } else if (code === ResponseCodeBookingTooManyUpcomingBookings) {
      return t("errorBookingLimit", {
        num: RuntimeConfig.INFOS.maxBookingsPerUser,
      });
    } else if (code === ResponseCodeBookingMaxConcurrentForUser) {
      return t("errorConcurrentBookingLimit", {
        num: RuntimeConfig.INFOS.maxConcurrentBookingsPerUser,
      });
    } else if (code === ResponseCodeBookingMaxHoursBeforeDelete) {
      return t("errorDeleteBookingBeforeMaxCancel", {
        num: RuntimeConfig.INFOS.maxHoursBeforeDelete,
      });
    } else if (code === ResponseCodeBookingInPast) {
      return t("errorInPast");
    } else if (code === ResponseCodePresenceReportDateRangeTooLong) {
      return t("errorDateRangeTooLong");
    } else if (code === ResponseCodeUsernameExists) {
      return t("errorUsernameExists");
    } else {
      return t("errorUnknown");
    }
  }
}
