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

  describe("parseTimeString", () => {
    it("should parse a valid HH:MM string", () => {
      expect(DateUtil.parseTimeString("08:30")).toBe("08:30");
    });

    it("should parse a valid single-digit HH:MM string", () => {
      expect(DateUtil.parseTimeString("8:30")).toBe("08:30");
    });

    it("should parse an hour-only string as HH:00", () => {
      expect(DateUtil.parseTimeString("8")).toBe("08:00");
    });

    it("should parse a two-digit hour-only string as HH:00", () => {
      expect(DateUtil.parseTimeString("18")).toBe("18:00");
    });

    it("should accept hour 23 as the maximum", () => {
      expect(DateUtil.parseTimeString("23")).toBe("23:00");
      expect(DateUtil.parseTimeString("23:59")).toBe("23:59");
    });

    it("should accept minute 59 as the maximum", () => {
      expect(DateUtil.parseTimeString("12:59")).toBe("12:59");
    });

    it("should return null if hour is 24 or greater", () => {
      expect(DateUtil.parseTimeString("24")).toBeNull();
      expect(DateUtil.parseTimeString("24:00")).toBeNull();
      expect(DateUtil.parseTimeString("25")).toBeNull();
      expect(DateUtil.parseTimeString("25:00")).toBeNull();
    });

    it("should return null if minute is greater than 59", () => {
      expect(DateUtil.parseTimeString("12:60")).toBeNull();
    });

    it("should accept 00:00 as the minimum", () => {
      expect(DateUtil.parseTimeString("00:00")).toBe("00:00");
    });

    it("should return null for non-numeric input", () => {
      expect(DateUtil.parseTimeString("abc")).toBeNull();
    });

    it("should return null for an empty string", () => {
      expect(DateUtil.parseTimeString("")).toBeNull();
    });

    it("should return null for malformed separators", () => {
      expect(DateUtil.parseTimeString("12:3:0")).toBeNull();
      expect(DateUtil.parseTimeString("12-30")).toBeNull();
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
