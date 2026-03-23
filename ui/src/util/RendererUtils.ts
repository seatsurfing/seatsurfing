export default class RendererUtils {
  static fullname(firstname: string, lastname: string): string {
    if (!firstname && !lastname) return "";
    if (firstname && lastname) return `${firstname} ${lastname}`;
    if (firstname) return firstname;
    return lastname;
  }

  static preAndSuffixIfDefined(s: string, prefix: string, suffix: string) {
    if (!s) return "";
    return `${prefix}${s}${suffix}`;
  }

  static decodeHtmlEntities(text: string): string {
    const textarea = document.createElement("textarea");
    textarea.innerHTML = text;
    return textarea.value;
  }

  static state(state: boolean | undefined): string {
    return state ? "☑" : "☐";
  }

  static shortenLink(url: string, maxLength: number): string {
    if (url.length <= maxLength) return url;
    const half = Math.floor((maxLength - 1) / 2) - 1;
    return `${url.slice(0, half)}[…]${url.slice(-half)}`;
  }

  static capitalize = (s: string) => s.charAt(0).toUpperCase() + s.slice(1);
}
