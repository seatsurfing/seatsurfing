export default class DateUtil {
  /**
   * @param date Date object to format
   * @returns formatted date string in "YYYY-MM-DD" format
   */
  private static formatToDateString(date: Date): string {
    return date.toISOString().split("T")[0];
  }

  /**
   * This methods formats a given date in the format "YYYY-MM-DDTHH:DD"
   * and ignores the date's timezone.
   *
   * @param date Date object to format
   * @returns formatted date string in "YYYY-MM-DDTHH:DD" (ISO 8601) format
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
    );
  };

  static isInPast(date: Date): boolean {
    return this.convertToUTC(date) < new Date();
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
   * @param date Date to set the seconds and milliseconds to a maximum
   * @returns new Date with the seconds and milliseconds set to maximum
   */
  static setSecondsToMax(date: Date): Date {
    const dateMaxSeconds = new Date(date);
    dateMaxSeconds.setSeconds(59, 999);
    return dateMaxSeconds;
  }
}
