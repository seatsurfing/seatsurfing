import RuntimeConfig from "@/components/RuntimeConfig";
import UserPreference, {
  PreferenceEnterTimeType,
} from "@/types/UserPreference";

export default class DateUtil {
  static MS_PER_MINUTE = 1000 * 60;
  static MS_PER_HOUR = DateUtil.MS_PER_MINUTE * 60;
  static MS_PER_DAY = DateUtil.MS_PER_HOUR * 24;

  /**
   * @param date Date object to format
   * @returns formatted date string in "YYYY-MM-DD" format
   */
  private static formatToDateString(date: Date): string {
    return date.toISOString().split("T")[0];
  }

  /**
   * This methods formats a given date in the format "YYYY-MM-DDTHH:MM"
   * and ignores the date's timezone.
   *
   * @param date Date object to format
   * @returns formatted date string in "YYYY-MM-DDTHH:MM" (ISO 8601) format
   */
  static formatToDateTimeString(date: Date): string {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const hours = String(date.getHours()).padStart(2, "0");
    const minutes = String(date.getMinutes()).padStart(2, "0");

    return `${year}-${month}-${day}T${hours}:${minutes}`;
  }

  /**
   * @param s string to test
   * @returns true if string is in data format "YYYY-MM-DD"
   */
  static isValidDate(s: string): boolean {
    const regex = /^\d{4}-\d{2}-\d{2}$/;
    if (!regex.test(s)) return false;

    const date = new Date(s);
    const timestamp = date.getTime();

    if (typeof timestamp !== "number" || Number.isNaN(timestamp)) {
      return false;
    }

    return s === this.formatToDateString(date);
  }

  /**
   * @param s string to test
   * @returns true if string is in data format "YYYY-MM-DDTHH:MM"
   */
  static isValidDateTime(s: string): boolean {
    const regex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/;
    if (!regex.test(s)) {
      return false;
    }

    const date = new Date(s);
    return !isNaN(date.getTime());
  }

  static getTodayDateString(): string {
    return this.getDateString(0);
  }

  static convertToUTC = (date: Date): Date => {
    return new Date(
      date.getUTCFullYear(),
      date.getUTCMonth(),
      date.getUTCDate(),
      date.getUTCHours(),
      date.getUTCMinutes(),
      date.getUTCSeconds(),
      0,
    );
  };

  static convertToFakeUTCDate(d: Date): Date {
    return new Date(
      Date.UTC(
        d.getFullYear(),
        d.getMonth(),
        d.getDate(),
        d.getHours(),
        d.getMinutes(),
        d.getSeconds(),
        0,
      ),
    );
  }

  static isInFuture(date: Date): boolean {
    return this.convertToUTC(date) > new Date();
  }

  static isAfterToday(date: Date): boolean {
    return this.isInFuture(date) && !this.isToday(date);
  }

  static isInPast(date: Date): boolean {
    return this.convertToUTC(date) < new Date();
  }

  static isToday(date: Date): boolean {
    return this.isSameDay(date, new Date());
  }

  /**
   * @param offset Offset in days
   * @returns return the date in format "YYYY-MM-DD"
   */
  static getDateString(offset: number): string {
    const date = new Date();
    date.setDate(date.getDate() + offset);
    return this.formatToDateString(date);
  }

  static getLastWeekMondayDateString(): string {
    const d = new Date();
    d.setDate(d.getDate() - (d.getDay() === 0 ? 6 : d.getDay() - 1) - 7);
    return this.formatToDateString(d);
  }

  static getLastWeekSundayDateString(): string {
    const d = new Date();
    d.setDate(d.getDate() - (d.getDay() === 0 ? 6 : d.getDay() - 1) - 1);
    return this.formatToDateString(d);
  }

  static getThisWeekMondayDateString(): string {
    const d = new Date();
    d.setDate(d.getDate() - (d.getDay() === 0 ? 6 : d.getDay() - 1));
    return this.formatToDateString(d);
  }

  static getThisWeekSundayDateString(): string {
    const d = new Date();
    d.setDate(d.getDate() - (d.getDay() === 0 ? 6 : d.getDay() - 1) + 6);
    return this.formatToDateString(d);
  }

  /**
   * @param date Date to modify
   * @returns new Date with the seconds set to a maximum (59.999)
   */
  static setSecondsToMax(date: Date): Date {
    const dateMaxSeconds = new Date(date);
    dateMaxSeconds.setSeconds(59, 999);
    return dateMaxSeconds;
  }

  /**
   * @param date Date to modify
   * @returns new Date with the hours set to a maximum (23:59:59.999)
   */
  static setHoursToMax(date: Date): Date {
    const dateMaxHours = new Date(date);
    dateMaxHours.setHours(23, 59, 59, 999);
    return dateMaxHours;
  }

  /**
   * @param date Date to modify
   * @returns new Date with the hours set to a minimum (00:00:00.000)
   */
  static setHoursToMin(date: Date): Date {
    const dateMaxHours = new Date(date);
    dateMaxHours.setHours(0, 0, 0, 0);
    return dateMaxHours;
  }

  /**
   * @returns Today's date with time 00:00:00.000
   */
  static getTodayStart(): Date {
    return this.setHoursToMin(new Date());
  }

  /**
   * @returns Today's date with time 23:59:59.999
   */
  static getTodayEnd(): Date {
    return this.setHoursToMax(new Date());
  }

  /**
   * @returns Today's date with a specified time
   */
  static getTodayTime(hour: number, minute: number, second: number): Date {
    const todayTime = new Date();
    todayTime.setHours(hour, minute, second, 0);
    return todayTime;
  }

  /**
   * @param totalMinutes minutes since midnight (0-1440)
   * @returns today's date with the time of day set accordingly
   */
  static getTodayTimeFromMinutes(totalMinutes: number): Date {
    return this.getTodayTime(
      Math.floor(totalMinutes / 60),
      totalMinutes % 60,
      0,
    );
  }

  /**
   * @param date Date to modify (mutated in place)
   * @param totalMinutes minutes since midnight (0-1440)
   * @returns the same date, with hours/minutes set accordingly
   */
  static setTimeFromMinutes(date: Date, totalMinutes: number): Date {
    date.setHours(Math.floor(totalMinutes / 60), totalMinutes % 60, 0, 0);
    return date;
  }

  /**
   * @param s time string in the format "HH:MM" (24h)
   * @returns minutes since midnight
   */
  static timeStringToMinutes(s: string): number {
    const [hours, minutes] = s.split(":").map((v) => parseInt(v, 10));
    return hours * 60 + minutes;
  }

  /**
   * @param date Date to format
   * @returns time string in the format "HH:MM" (24h)
   */
  static formatTimeString(date: Date): string {
    const hours = String(date.getHours()).padStart(2, "0");
    const minutes = String(date.getMinutes()).padStart(2, "0");
    return `${hours}:${minutes}`;
  }

  /**
   * @param s time string in the format "HH:MM" (24h)
   * @returns today's date with the time of day set accordingly
   */
  static getTodayTimeFromTimeString(s: string): Date {
    return this.getTodayTimeFromMinutes(this.timeStringToMinutes(s));
  }

  /**
   * @param date Date to modify (mutated in place)
   * @param s time string in the format "HH:MM" (24h)
   * @returns the same date, with hours/minutes set accordingly
   */
  static setTimeFromTimeString(date: Date, s: string): Date {
    return this.setTimeFromMinutes(date, this.timeStringToMinutes(s));
  }

  static copyDate(source: Date, target: Date): Date {
    const result = new Date(target);
    result.setFullYear(
      source.getFullYear(),
      source.getMonth(),
      source.getDate(),
    );
    return result;
  }

  static copyTime(source: Date, target: Date): Date {
    const result = new Date(target);
    result.setHours(
      source.getHours(),
      source.getMinutes(),
      source.getSeconds(),
      source.getMilliseconds(),
    );
    return result;
  }

  /**
   * @param date1 Date1 to compare
   * @param date2 Date2 to compare
   * @returns true, if both dates are on the same day
   */
  static isSameDay(date1: Date, date2: Date): boolean {
    return (
      date1.getFullYear() === date2.getFullYear() &&
      date1.getMonth() === date2.getMonth() &&
      date1.getDate() === date2.getDate()
    );
  }

  /**
   * @param date1 Date1 to compare
   * @param date2 Date2 to compare
   * @returns true, if both dates have the same time (hours and minutes)
   */
  static isSameTime(date1: Date, date2: Date): boolean {
    return (
      date1.getHours() === date2.getHours() &&
      date1.getMinutes() === date2.getMinutes()
    );
  }

  static equal(date1: Date, date2: Date): boolean {
    return date1.getTime() === date2.getTime();
  }

  static prevDay(date: Date): Date {
    const nextDay = new Date(date);
    nextDay.setDate(nextDay.getDate() - 1);
    return nextDay;
  }

  static nextDay(date: Date): Date {
    const nextDay = new Date(date);
    nextDay.setDate(nextDay.getDate() + 1);
    return nextDay;
  }

  static getWeekStart(date: Date): Date {
    const d = new Date(date);
    d.setDate(d.getDate() - d.getDay());
    d.setHours(0, 0, 0, 0);
    return d;
  }

  static getWeekEnd(date: Date): Date {
    const d = new Date(date);
    d.setDate(d.getDate() + (6 - d.getDay()));
    d.setHours(23, 59, 59, 999);
    return d;
  }

  static getNowFakeUTC(): Date {
    return this.convertToFakeUTCDate(new Date());
  }

  static hoursToDay(hours: number) {
    return Math.floor(hours / 24);
  }

  /**
   * calculates the "next free enter time" for a booking based on a leave date which is
   *  - next day if "dailyBasisBooking" is active, or
   *  - +1 minute otherwise
   *
   * @param leave Leave date of the last booking
   * @returns new date for "next free enter time"
   */
  static getNextFreeEnterTime(leave: Date): Date {
    if (RuntimeConfig.INFOS.dailyBasisBooking) {
      const nextDay = new Date(leave);
      nextDay.setUTCDate(nextDay.getUTCDate() + 1);
      nextDay.setUTCHours(0, 0, 0, 0);
      return nextDay;
    }
    return new Date(leave.getTime() + DateUtil.MS_PER_MINUTE);
  }

  /**
   * Parses a string into a time string in format "HH:MM" (24h).
   * Accepts "HH:MM" or just "HH" (in which case minutes default to "00").
   * Hours may be 0-24, minutes 0-59.
   *
   * @param s string to parse into time string
   * @returns time string in format "HH:MM" (24h), or null if not parsable
   */
  static parseTimeString(s: string): string | null {
    const match = s.match(/^(\d{1,2})(?::(\d{1,2}))?$/);
    if (!match) {
      return null;
    }

    const hours = parseInt(match[1], 10);
    const minutes = match[2] !== undefined ? parseInt(match[2], 10) : 0;

    if (hours < 0 || hours > 24 || minutes < 0 || minutes > 59) {
      return null;
    }

    return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
  }

  /**
   * Calculates the next enter and leave time based on the user's
   * preferred (workday) times and the org's global (booking) settings
   *
   * @returns default enter and leave time for a new booking
   */
  static getNextPreferredEnterAndLeaveTime(
    prefEnterTime: PreferenceEnterTimeType,
    prefWorkdayStart: string,
    prefWorkdayEnd: string,
    prefWorkdays: number[],
    dailyBasisBooking: boolean,
  ): { enter: Date; leave: Date } {
    const prefWorkdayStartMinutes =
      DateUtil.timeStringToMinutes(prefWorkdayStart);
    const prefWorkdayEndMinutes = DateUtil.timeStringToMinutes(prefWorkdayEnd);

    let enter = new Date();
    if (prefEnterTime === UserPreference.PreferenceEnterTime.Now) {
      enter.setHours(enter.getHours() + 1, 0, 0);
      const enterMinutes = enter.getHours() * 60 + enter.getMinutes();
      if (enterMinutes < prefWorkdayStartMinutes) {
        // preferred start time works for today
        DateUtil.setTimeFromTimeString(enter, prefWorkdayStart);
      }
      if (enterMinutes >= prefWorkdayEndMinutes) {
        // todays next start time is after preferred end date -> switch to next day
        enter.setDate(enter.getDate() + 1);
        DateUtil.setTimeFromTimeString(enter, prefWorkdayStart);
      }
    } else if (prefEnterTime === UserPreference.PreferenceEnterTime.NextDay) {
      enter.setDate(enter.getDate() + 1);
      DateUtil.setTimeFromTimeString(enter, prefWorkdayStart);
    } else if (
      prefEnterTime === UserPreference.PreferenceEnterTime.NextWorkday
    ) {
      enter.setDate(enter.getDate() + 1);
      let add = 0;
      let nextDayFound = false;
      let lookFor = enter.getDay();
      while (!nextDayFound) {
        if (prefWorkdays.includes(lookFor) || add > 7) {
          nextDayFound = true;
        } else {
          add++;
          lookFor++;
          if (lookFor > 6) {
            lookFor = 0;
          }
        }
      }
      enter.setDate(enter.getDate() + add);
      DateUtil.setTimeFromTimeString(enter, prefWorkdayStart);
    }

    let leave = new Date(enter);
    DateUtil.setTimeFromTimeString(leave, prefWorkdayEnd);

    if (dailyBasisBooking) {
      enter = DateUtil.setHoursToMin(enter);
      leave = DateUtil.setHoursToMax(leave);
    }

    return { enter, leave };
  }
}
