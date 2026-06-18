import { describe, it, expect, beforeEach, afterEach } from "vitest";
import DateUtil from "./DateUtil";
import RuntimeConfig from "@/components/RuntimeConfig";

describe("DateUtil", () => {
  describe("setHoursToMin", () => {
    it("should return new Date", () => {
      const date = new Date();
      const dateTimestamp = date.getTime();
      const dateHoursToMin = DateUtil.setHoursToMin(date);

      expect(dateTimestamp).toBe(date.getTime());
      expect(dateHoursToMin.getHours()).toBe(0);
    });
  });

  describe("getTodayStart", () => {
    it("should return date with time 00:00:00.000", () => {
      const todayStart = DateUtil.getTodayStart();

      expect(todayStart.getHours()).toBe(0);
      expect(todayStart.getMinutes()).toBe(0);
      expect(todayStart.getSeconds()).toBe(0);
      expect(todayStart.getMilliseconds()).toBe(0);
    });
  });

  describe("isSameDate", () => {
    it("should return true if date1=date2", () => {
      const date = new Date();
      expect(DateUtil.isSameDay(date, date)).toBe(true);
    });
  });

  describe("getNextFreeEnterTime", () => {
    it("should return 2026-04-25T00:00:00.000Z for leave 2026-04-24T16:59:59.000Z with dailyBasisBooking", () => {
      RuntimeConfig.INFOS = { ...RuntimeConfig.INFOS, dailyBasisBooking: true };
      const leave = new Date("2026-04-24T16:59:59.000Z");
      const result = DateUtil.getNextFreeEnterTime(leave);
      expect(result.toISOString()).toBe("2026-04-25T00:00:00.000Z");
    });
  });
});
