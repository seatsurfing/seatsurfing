import { describe, it, expect, beforeEach } from "vitest";
import CalendarUtil from "./CalendarUtil";
import RuntimeConfig from "@/components/RuntimeConfig";
import Formatting from "@/util/Formatting";

describe("CalendarUtil", () => {
  describe("getWeekRange", () => {
    beforeEach(() => {
      RuntimeConfig.INFOS.weekStartDay = 1;
    });

    it("returns the full week for the week view", () => {
      // Wednesday
      const date = new Date(Date.UTC(2030, 8, 4, 12, 0, 0));
      const range = CalendarUtil.getWeekRange(date);

      expect(range.start.toISOString()).toBe("2030-09-02T00:00:00.000Z");
      expect(range.end.toISOString()).toBe("2030-09-08T23:59:59.999Z");
    });
  });

  describe("getFormats", () => {
    const local = CalendarUtil.getLocalizer();
    const start = new Date(Date.UTC(2030, 8, 4, 9, 0, 0));
    const end = new Date(Date.UTC(2030, 8, 4, 17, 30, 0));

    it("uses 24-hour time when configured", () => {
      RuntimeConfig.INFOS.use24HourTime = true;
      const formats = CalendarUtil.getFormats();
      expect(local.format(end, formats.timeGutterFormat as string)).toBe(
        "17:30",
      );
      expect(
        (formats.eventTimeRangeFormat as any)({ start, end }, undefined, local),
      ).toBe("09:00 – 17:30");
    });

    it("formats weekdays in the configured language", () => {
      Formatting.Language = "de";
      const formats = CalendarUtil.getFormats();
      expect((formats.weekdayFormat as any)(start)).toBe("Mi");
      expect((formats.dayFormat as any)(start)).toBe("04 Mi");
      Formatting.Language = "en-GB";
      expect((CalendarUtil.getFormats().weekdayFormat as any)(start)).toBe(
        "Wed",
      );
    });

    it("formats the month header in the configured language", () => {
      const date = new Date(Date.UTC(2030, 2, 1));
      Formatting.Language = "de";
      expect((CalendarUtil.getFormats().monthHeaderFormat as any)(date)).toBe(
        "März 2030",
      );
      Formatting.Language = "en-GB";
      expect((CalendarUtil.getFormats().monthHeaderFormat as any)(date)).toBe(
        "March 2030",
      );
    });

    it("uses 12-hour time when configured", () => {
      RuntimeConfig.INFOS.use24HourTime = false;
      const formats = CalendarUtil.getFormats();
      expect(local.format(end, formats.timeGutterFormat as string)).toBe(
        "5:30 PM",
      );
      expect(
        (formats.eventTimeRangeFormat as any)({ start, end }, undefined, local),
      ).toBe("9:00 AM – 5:30 PM");
    });
  });

  describe("getPeriodLabel", () => {
    beforeEach(() => {
      RuntimeConfig.INFOS.weekStartDay = 1;
    });

    const date = new Date(Date.UTC(2030, 8, 4, 12, 0, 0));
    const label = (view: "week" | "month" | "day", d: Date) =>
      CalendarUtil.getPeriodLabel(view, d).replace(/\s+/g, " ");

    it("formats the week range compactly", () => {
      Formatting.Language = "en-GB";
      const l = label("week", date);
      expect(l).toMatch(/^2\s?–\s?8 Sept? 2030$/);
      Formatting.Language = "de";
      expect(label("week", date)).toMatch(/^2\.\s?–\s?8\. Sept?\. 2030$/);
    });

    it("includes both months for a week spanning two months", () => {
      Formatting.Language = "en-GB";
      const l = label("week", new Date(Date.UTC(2030, 8, 30, 12, 0, 0)));
      expect(l).toMatch(/^30 Sept? – 6 Oct 2030$/);
    });

    it("formats month and day views", () => {
      Formatting.Language = "en-GB";
      expect(label("month", date)).toBe("September 2030");
      expect(label("day", date)).toMatch(/^Wednesday,? 4 September 2030$/);
    });
  });

  describe("formatTime", () => {
    const date = new Date(Date.UTC(2030, 8, 4, 14, 5, 0));

    it("respects the configured hour format", () => {
      RuntimeConfig.INFOS.use24HourTime = true;
      expect(CalendarUtil.formatTime(date)).toBe("14:05");
      RuntimeConfig.INFOS.use24HourTime = false;
      expect(CalendarUtil.formatTime(date)).toBe("2:05 PM");
    });
  });

  describe("prop getters", () => {
    it("marks pending events with a class", () => {
      expect(
        CalendarUtil.eventPropGetter({ approved: false } as any).className,
      ).toBe("calendar-event calendar-event-pending");
      expect(
        CalendarUtil.eventPropGetter({ approved: true } as any).className,
      ).toBe("calendar-event");
    });

    it("marks days outside the allowed weekdays", () => {
      const getter = CalendarUtil.getDayPropGetter([1, 2, 3, 4, 5]);
      const saturday = new Date(Date.UTC(2030, 8, 7));
      const monday = new Date(Date.UTC(2030, 8, 2));
      expect(getter(saturday)).toEqual({ className: "calendar-day-off" });
      expect(getter(monday)).toEqual({});
      expect(CalendarUtil.getDayPropGetter([])(saturday)).toEqual({});
    });
  });
});
