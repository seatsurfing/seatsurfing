import { describe, it, expect, beforeEach } from "vitest";
import CalendarUtil from "./CalendarUtil";
import RuntimeConfig from "@/components/RuntimeConfig";

describe("CalendarUtil", () => {
  describe("getRange", () => {
    beforeEach(() => {
      RuntimeConfig.INFOS.weekStartDay = 1;
    });

    it("returns the full week for the week view", () => {
      // Wednesday
      const date = new Date(Date.UTC(2030, 8, 4, 12, 0, 0));
      const range = CalendarUtil.getRange(date, "week");

      expect(range.start.toISOString()).toBe("2030-09-02T00:00:00.000Z");
      expect(range.end.toISOString()).toBe("2030-09-08T23:59:59.999Z");
    });

    it("includes adjacent days shown in the month view", () => {
      const date = new Date(Date.UTC(2030, 8, 15, 12, 0, 0));
      const range = CalendarUtil.getRange(date, "month");

      expect(range.start.toISOString()).toBe("2030-08-26T00:00:00.000Z");
      expect(range.end.toISOString()).toBe("2030-10-06T23:59:59.999Z");
    });
  });
});
