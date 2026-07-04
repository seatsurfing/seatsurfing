import RuntimeConfig from "@/components/RuntimeConfig";

export interface DateFormatter {
  format(date?: Date | number): string;
}

export default class Formatting {
  static Language: string = "en";
  static t: (key: string, view?: object) => any;

  private static formatDatePart(
    date: Date | number | undefined,
    local?: boolean,
  ): string {
    const d =
      date === undefined
        ? new Date()
        : typeof date === "number"
          ? new Date(date)
          : date;
    const year = local ? d.getFullYear() : d.getUTCFullYear();
    const month = (local ? d.getMonth() : d.getUTCMonth()) + 1;
    const day = local ? d.getDate() : d.getUTCDate();
    return RuntimeConfig.INFOS.dateFormat
      .replace("Y", year.toString().padStart(4, "0"))
      .replace("m", month.toString().padStart(2, "0"))
      .replace("d", day.toString().padStart(2, "0"));
  }

  private static getWeekdayFormatter(local?: boolean): Intl.DateTimeFormat {
    return new Intl.DateTimeFormat(Formatting.Language, {
      timeZone: local ? undefined : "UTC",
      weekday: "long",
    });
  }

  private static getTimeFormatter(local?: boolean): Intl.DateTimeFormat {
    return new Intl.DateTimeFormat(Formatting.Language, {
      timeZone: local ? undefined : "UTC",
      hour: "numeric",
      minute: "numeric",
      hour12: !RuntimeConfig.INFOS.use24HourTime,
    });
  }

  static getFormatter(local?: boolean): DateFormatter {
    const weekdayFormatter = Formatting.getWeekdayFormatter(local);
    const timeFormatter = Formatting.getTimeFormatter(local);
    return {
      format: (date?: Date | number) =>
        `${weekdayFormatter.format(date)}, ${Formatting.formatDatePart(date, local)}, ${timeFormatter.format(date)}`,
    };
  }

  static getFormatterNoTime(local?: boolean): DateFormatter {
    const weekdayFormatter = Formatting.getWeekdayFormatter(local);
    return {
      format: (date?: Date | number) =>
        `${weekdayFormatter.format(date)}, ${Formatting.formatDatePart(date, local)}`,
    };
  }

  static getBookingDateFormatter(): DateFormatter {
    return RuntimeConfig.INFOS.dailyBasisBooking
      ? Formatting.getFormatterNoTime()
      : Formatting.getFormatter();
  }

  static getFormatterShort(local?: boolean): DateFormatter {
    const timeFormatter = Formatting.getTimeFormatter(local);
    return {
      format: (date?: Date | number) =>
        `${Formatting.formatDatePart(date, local)}, ${timeFormatter.format(date)}`,
    };
  }

  static getFormatterDate(local?: boolean): DateFormatter {
    return {
      format: (date?: Date | number) => Formatting.formatDatePart(date, local),
    };
  }

  static getDateTimePickerFormatString(): string {
    const date = Date.UTC(2006, 11, 23, 11, 41, 52, 0);
    const formattedDate = Formatting.getFormatterShort().format(date);
    return formattedDate
      .replace("2006", "y")
      .replace("12", "MM")
      .replace("23", "dd")
      .replace("11", "HH")
      .replace("41", "mm");
  }

  static getDateTimePickerFormatDailyString(): string {
    const date = Date.UTC(2006, 11, 23, 11, 41, 52, 0);
    const formattedDate = Formatting.getFormatterDate().format(date);
    return formattedDate
      .replace("2006", "y")
      .replace("12", "MM")
      .replace("23", "dd");
  }

  static getDayValue(date: Date): number {
    const s =
      date.getFullYear().toString().padStart(4, "0") +
      (date.getMonth() + 1).toString().padStart(2, "0") +
      date.getDate().toString().padStart(2, "0");
    return parseInt(s);
  }

  static getDayDiff(date1: Date, date2: Date): number {
    const d1 = new Date(date1.valueOf());
    d1.setHours(0, 0, 0, 0);
    const d2 = new Date(date2.valueOf());
    d2.setHours(0, 0, 0, 0);
    return Math.floor((d1.getTime() - d2.getTime()) / (1000 * 60 * 60 * 24));
  }

  static getISO8601(date: Date): string {
    const s =
      date.getFullYear().toString().padStart(4, "0") +
      "-" +
      (date.getMonth() + 1).toString().padStart(2, "0") +
      "-" +
      date.getDate().toString().padStart(2, "0");
    return s;
  }

  static getDateOffsetText(enter: Date, leave: Date): string {
    const today = Formatting.getDayValue(new Date());
    const start = Formatting.getDayValue(enter);
    const end = Formatting.getDayValue(leave);
    if (start <= today && today <= end) {
      return Formatting.t("today");
    }
    if (start == today + 1) {
      return Formatting.t("tomorrow");
    }
    if (start > today && start <= today + 7) {
      return Formatting.t("inXdays", { x: start - today });
    }
    return Formatting.getFormatterDate().format(enter);
  }

  static stripTimezoneDetails(s: string): string {
    const match = s.match(/^(.+?)([+-]\d{2}:\d{2})$/);
    if (!match) {
      return s;
    }
    const withoutOffset = match[1];
    const fractionMatch = withoutOffset.match(/\.(\d+)$/);
    if (fractionMatch) {
      const ms = fractionMatch[1].padEnd(3, "0").slice(0, 3);
      return withoutOffset.replace(/\.\d+$/, "." + ms) + "Z";
    }
    return withoutOffset + ".000Z";
  }
}
