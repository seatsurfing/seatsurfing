export default class DateUtil {
  /**
   * @param s string to test
   * @returns true if string is in data format YYYY-MM-DD
   */
  static isValidDate(s: string): boolean {
    const regex = /^\d{4}-\d{2}-\d{2}$/;
    if (!regex.test(s)) return false;

    const date = new Date(s);
    const timestamp = date.getTime();

    if (typeof timestamp !== "number" || Number.isNaN(timestamp)) {
      return false;
    }

    return s === date.toISOString().split("T")[0];
  }

  static getTodayDateString(): string {
    return this.getDateString(0);
  }

  /**
   *
   * @param offset Offset in days
   * @returns return the date in format YYYY-MM-DD
   */
  static getDateString(offset: number): string {
    const date = new Date();
    date.setDate(date.getDate() + offset);
    return date.toISOString().split("T")[0];
  }

  static getLastWeekMondayDateString(): string {
    const d = new Date();
    d.setDate(d.getDate() - (d.getDay() === 0 ? 6 : d.getDay() - 1) - 7);
    return d.toISOString().split("T")[0];
  }

  static getLastWeekSundayDateString(): string {
    const d = new Date();
    d.setDate(d.getDate() - (d.getDay() === 0 ? 6 : d.getDay() - 1) - 1);
    return d.toISOString().split("T")[0];
  }

  static getThisWeekMondayDateString(): string {
    const d = new Date();
    d.setDate(d.getDate() - (d.getDay() === 0 ? 6 : d.getDay() - 1));
    return d.toISOString().split("T")[0];
  }

  static getThisWeekSundayDateString(): string {
    const d = new Date();
    d.setDate(d.getDate() - (d.getDay() === 0 ? 6 : d.getDay() - 1) + 6);
    return d.toISOString().split("T")[0];
  }
}
