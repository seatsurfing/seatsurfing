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

  UsernameExists = 3001,

  GroupNameAlreadyExists = 4001,
}

export default class ErrorText {
  static getTextForAppCode(code: number, t: TranslationFunc): string {
    const { INFOS } = RuntimeConfig;

    const errorMap: Partial<Record<ResponseCode, () => string>> = {
      [ResponseCode.BookingSlotConflict]: () => t("errorSlotConflict"),
      [ResponseCode.BookingLocationMaxConcurrent]: () =>
        t("errorTooManyConcurrent"),
      [ResponseCode.BookingInvalidMaxBookingDuration]: () =>
        t("errorMaxBookingDuration", { num: INFOS.maxBookingDurationHours }),
      [ResponseCode.BookingInvalidMinBookingDuration]: () =>
        t("errorMinBookingDuration", { num: INFOS.minBookingDurationHours }),
      [ResponseCode.BookingTooManyDaysInAdvance]: () =>
        t("errorDaysAdvance", { num: INFOS.maxDaysInAdvance }),
      [ResponseCode.BookingTooManyUpcomingBookings]: () =>
        t("errorBookingLimit", { num: INFOS.maxBookingsPerUser }),
      [ResponseCode.BookingMaxConcurrentForUser]: () =>
        t("errorConcurrentBookingLimit", {
          num: INFOS.maxConcurrentBookingsPerUser,
        }),
      [ResponseCode.BookingMaxHoursBeforeDelete]: () =>
        t("errorDeleteBookingBeforeMaxCancel", {
          num: INFOS.maxHoursBeforeDelete,
        }),
      [ResponseCode.BookingInPast]: () => t("errorInPast"),
      [ResponseCode.PresenceReportDateRangeTooLong]: () =>
        t("errorDateRangeTooLong"),
      [ResponseCode.UsernameExists]: () => t("errorUsernameExists"),
      [ResponseCode.GroupNameAlreadyExists]: () =>
        t("errorGroupNameAlreadyExists"),
    };

    return errorMap[code as ResponseCode]?.() ?? t("errorUnknown");
  }
}
