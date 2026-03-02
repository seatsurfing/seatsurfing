import RuntimeConfig from "@/components/RuntimeConfig";

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
}
