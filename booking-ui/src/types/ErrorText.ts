import RuntimeConfig from "@/components/RuntimeConfig";
import { TranslationFunc } from "@/components/withTranslation";

//var ResponseCodeBookingSlotConflict: number             = 1001;
var ResponseCodeBookingLocationMaxConcurrent: number = 1002;
var ResponseCodeBookingTooManyUpcomingBookings: number = 1003;
var ResponseCodeBookingTooManyDaysInAdvance: number = 1004;
var ResponseCodeBookingInvalidMaxBookingDuration: number = 1005;
var ResponseCodeBookingMaxConcurrentForUser: number = 1006;
var ResponseCodeBookingInvalidMinBookingDuration: number = 1007;
var ResponseCodeBookingMaxHoursBeforeDelete: number = 1008;

export default class ErrorText {
    static getTextForAppCode(code: number, t: TranslationFunc): string {
        if (code === ResponseCodeBookingLocationMaxConcurrent) {
            return t("errorTooManyConcurrent");
        } else if (code === ResponseCodeBookingInvalidMaxBookingDuration) {
            return t("errorMaxBookingDuration", { "num": RuntimeConfig.INFOS.maxBookingDurationHours });
        } else if (code === ResponseCodeBookingInvalidMinBookingDuration) {
            return t("errorMinBookingDuration", { "num": RuntimeConfig.INFOS.minBookingDurationHours });
        } else if (code === ResponseCodeBookingTooManyDaysInAdvance) {
            return t("errorDaysAdvance", { "num": RuntimeConfig.INFOS.maxDaysInAdvance });
        } else if (code === ResponseCodeBookingTooManyUpcomingBookings) {
            return t("errorBookingLimit", { "num": RuntimeConfig.INFOS.maxBookingsPerUser });
        } else if (code === ResponseCodeBookingMaxConcurrentForUser) {
            return t("errorConcurrentBookingLimit", { "num": RuntimeConfig.INFOS.maxConcurrentBookingsPerUser });
        } else if (code === ResponseCodeBookingMaxHoursBeforeDelete) {
            return t("errorDeleteBookingBeforeMaxCancel", { "num": RuntimeConfig.INFOS.maxHoursBeforeDelete });
        } else {
            return t("errorUnknown");
        }
    }
}
