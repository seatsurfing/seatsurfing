import RuntimeConfig from "@/components/RuntimeConfig";
import { TranslationFunc } from "@/components/withTranslation";

const ResponseCodeBookingSlotConflict: number = 1001;
const ResponseCodeBookingLocationMaxConcurrent: number = 1002;
const ResponseCodeBookingTooManyUpcomingBookings: number = 1003;
const ResponseCodeBookingTooManyDaysInAdvance: number = 1004;
const ResponseCodeBookingInvalidMaxBookingDuration: number = 1005;
const ResponseCodeBookingMaxConcurrentForUser: number = 1006;
const ResponseCodeBookingInvalidMinBookingDuration: number = 1007;
const ResponseCodeBookingMaxHoursBeforeDelete: number = 1008;

const ResponseCodePresenceReportDateRangeTooLong: number = 2001;

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
    } else if (code === ResponseCodePresenceReportDateRangeTooLong) {
      return t("errorDateRangeTooLong");
    } else {
      return t("errorUnknown");
    }
  }
}
