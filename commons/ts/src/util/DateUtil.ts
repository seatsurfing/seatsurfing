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
}
