import { describe, it, expect, beforeEach } from "vitest";
import CalendarUtil from "./CalendarUtil";
import RuntimeConfig from "@/components/RuntimeConfig";

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
});
