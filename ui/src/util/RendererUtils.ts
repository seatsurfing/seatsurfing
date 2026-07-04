import { TranslationFunc } from "@/components/withTranslation";

export default class RendererUtils {
  static readonly BREAKPOINT_SMALL = 576;

  static fullname(
    firstname: string,
    lastname: string,
    fallback?: string,
  ): string {
    if (!firstname && !lastname) return fallback || "";
    if (firstname && lastname) return `${firstname} ${lastname}`;
    if (firstname) return firstname;
    return lastname;
  }

  static preAndSuffixIfDefined(s: string, prefix: string, suffix: string) {
    if (!s) return "";
    return `${prefix}${s}${suffix}`;
  }

  static suffixIfDefined(s: string, suffix: string) {
    if (!s) return "";
    return `${s}${suffix}`;
  }

  static escapeHtml(s: string): string {
    return s
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
  }

  static decodeHtmlEntities(text: string): string {
    const textarea = document.createElement("textarea");
    textarea.innerHTML = text;
    return textarea.value;
  }

  static state(state: boolean | undefined): string {
    return state ? "☑" : "☐";
  }

  static stateXls(state: boolean | undefined, t: TranslationFunc): string {
    return state ? t("yes") : t("no");
  }

  static numberPlus(number: number, max: number): string {
    return number > max ? `${max}+` : String(number);
  }

  static shortenLink(url: string, maxLength: number): string {
    if (url.length <= maxLength) return url;
    const half = Math.floor((maxLength - 1) / 2) - 1;
    return `${url.slice(0, half)}[…]${url.slice(-half)}`;
  }

  static isSpaceVertical(
    width: number,
    height: number,
    rotation: number,
  ): boolean {
    const normalizedRot = ((rotation % 180) + 180) % 180;
    const dimensionsSwapped = normalizedRot >= 45 && normalizedRot <= 135;
    return dimensionsSwapped ? width > height : width < height;
  }

  static capitalize = (s: string) => s.charAt(0).toUpperCase() + s.slice(1);

  static SECRET_PLACEHOLDER = "••••••••••••••••";

  static readonly SPACE_FONT_SIZES: { [key: string]: number } = {
    small: 10,
    normal: 12,
    big: 14,
    bigger: 18,
  };

  static readonly SPACE_FONT_SIZE_OPTIONS: string[] = Object.keys(
    RendererUtils.SPACE_FONT_SIZES,
  );

  static spaceFontSizePx(spaceFontSize: string | undefined): number {
    return (
      RendererUtils.SPACE_FONT_SIZES[spaceFontSize || "normal"] ||
      RendererUtils.SPACE_FONT_SIZES.normal
    );
  }
}
