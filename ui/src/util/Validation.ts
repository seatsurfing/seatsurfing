export default class Validation {
  static isAbsoluteUrl(url: string): boolean {
    return /^https?:\/\//i.test(url);
  }
}
