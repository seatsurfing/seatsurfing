export const PASSWORD_PATTERN = "^(?=.*[A-Z])(?=.*[a-z])(?=.*[0-9])(?=.*[^a-zA-Z0-9]).{8,}$";

export default class Validation {
  static isAbsoluteUrl(url: string): boolean {
    return /^https?:\/\//i.test(url);
  }
}
