export const PASSWORD_PATTERN =
  "^(?=.*[A-Z])(?=.*[a-z])(?=.*[0-9])(?=.*[^a-zA-Z0-9]).{8,}$";
export const PASSWORD_MIN_LENGTH = 8;
export const PASSWORD_MAX_LENGTH = 64;

export default class Validation {
  static isAbsoluteUrl(url: string): boolean {
    return /^https?:\/\//i.test(url);
  }
}
