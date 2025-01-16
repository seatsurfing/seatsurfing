import { TFunction } from "i18next";

export default class Formatting {
    static Language: string = "en";
    static t: TFunction;

    static tbool(s: string) {
        return Formatting.t(s) === "1";
    }

    static getFormatter(local?: boolean): Intl.DateTimeFormat {
        let formatter = new Intl.DateTimeFormat(Formatting.Language, {
            timeZone: local ? undefined : 'UTC',
            weekday: 'long',
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: 'numeric',
            minute: 'numeric',
            hour12: this.tbool("hour12")
        });
        return formatter;
    }

    static getFormatterNoTime(local?: boolean): Intl.DateTimeFormat {
        let formatter = new Intl.DateTimeFormat(Formatting.Language, {
            timeZone: local ? undefined : 'UTC',
            weekday: 'long',
            year: 'numeric',
            month: '2-digit',
            day: '2-digit'
        });
        return formatter;
    }

    static getFormatterShort(local?: boolean): Intl.DateTimeFormat {
        let formatter = new Intl.DateTimeFormat(Formatting.Language, {
            timeZone: local ? undefined : 'UTC',
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: 'numeric',
            minute: 'numeric',
            hour12: false
        });
        return formatter;
    }

    static getFormatterDate(local?: boolean): Intl.DateTimeFormat {
        let formatter = new Intl.DateTimeFormat(Formatting.Language, {
            timeZone: local ? undefined : 'UTC',
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
        });
        return formatter;
    }

    static getDateTimePickerFormatString(): string {
        let date = Date.UTC(2006, 11, 23, 11, 41, 52, 0);
        let formattedDate = Formatting.getFormatterShort().format(date);
        return formattedDate.replace('2006', 'y').replace('12', 'MM').replace('23', 'dd').replace('11', 'HH').replace('41', 'mm');
    }

    static getDateTimePickerFormatDailyString(): string {
        let date = Date.UTC(2006, 11, 23, 11, 41, 52, 0);
        let formattedDate = Formatting.getFormatterDate().format(date);
        return formattedDate.replace('2006', 'y').replace('12', 'MM').replace('23', 'dd');
    }

    static getDayValue(date: Date): number {
        let s = date.getFullYear().toString().padStart(4, "0") + (date.getMonth() + 1).toString().padStart(2, "0") + date.getDate().toString().padStart(2, "0");
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
        let s = date.getFullYear().toString().padStart(4, "0") + "-" + (date.getMonth() + 1).toString().padStart(2, "0") + "-" + date.getDate().toString().padStart(2, "0");
        return s;
    }

    static getDateOffsetText(enter: Date, leave: Date): string {
        let today = Formatting.getDayValue(new Date());
        let start = Formatting.getDayValue(enter);
        let end = Formatting.getDayValue(leave);
        if (start <= today && today <= end) {
            return Formatting.t("today");
        }
        if (start == today + 1) {
            return Formatting.t("tomorrow");
        }
        if (start > today && start <= today + 7) {
            return Formatting.t("inXdays", { "x": (start - today) });
        }
        return Formatting.getFormatterDate().format(enter);
    }

    static convertToFakeUTCDate(d: Date): Date {
        return new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours(), d.getMinutes(), d.getSeconds(), 0));
    }

    static stripTimezoneDetails(s: string): string {
        if ((s.length > 6) && ((s[s.length - 6] === "+") || (s[s.length - 6] === "-"))) {
            return s.substring(0, s.length - 6) + ".000Z";
        }
        return s;
    }
}