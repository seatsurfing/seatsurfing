export default class Validation {
  static readonly PASSWORD_PATTERN =
    "^(?=.*[A-Z])(?=.*[a-z])(?=.*[0-9])(?=.*[^a-zA-Z0-9]).{8,}$";
  static readonly PASSWORD_MIN_LENGTH = 8;
  static readonly PASSWORD_MAX_LENGTH = 64;
  static readonly PASSWORD_MIN_LENGTH_SA = 32;

  static isAbsoluteUrl(url: string): boolean {
    return /^https?:\/\//i.test(url);
  }

  static generatePassword(length: number = 32): string {
    const lower = "abcdefghijklmnopqrstuvwxyz";
    const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    const digits = "0123456789";
    const special = "!@#$%^&*()-_=+[]{}|;:,.<>?";
    const all = lower + upper + digits + special;
    const required = [
      lower.charAt(Math.floor(Math.random() * lower.length)),
      upper.charAt(Math.floor(Math.random() * upper.length)),
      digits.charAt(Math.floor(Math.random() * digits.length)),
      special.charAt(Math.floor(Math.random() * special.length)),
    ];
    const rest = Array.from({ length: length - required.length }, () =>
      all.charAt(Math.floor(Math.random() * all.length)),
    );
    const chars = [...required, ...rest];
    for (let i = chars.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [chars[i], chars[j]] = [chars[j], chars[i]];
    }
    return chars.join("");
  }
}
