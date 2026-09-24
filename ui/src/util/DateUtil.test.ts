import { describe, it, expect, afterEach, vi } from "vitest";
import DateUtil from "./DateUtil";
import RuntimeConfig from "@/components/RuntimeConfig";
import UserPreference, {
  PreferenceEnterTimeType,
} from "@/types/UserPreference";

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

  describe("getMondayOfWeek / getSundayOfWeek", () => {
    it("should return Monday 00:00 and Sunday 23:59 for a weekday", () => {
      const wednesday = new Date(2026, 8, 23, 14, 30);
      const monday = DateUtil.getMondayOfWeek(wednesday);
      const sunday = DateUtil.getSundayOfWeek(wednesday);

      expect(monday).toEqual(new Date(2026, 8, 21, 0, 0, 0, 0));
      expect(sunday).toEqual(new Date(2026, 8, 27, 23, 59, 59, 999));
      expect(wednesday).toEqual(new Date(2026, 8, 23, 14, 30));
    });

    it("should treat Sunday as last day of the week", () => {
      const sundayInput = new Date(2026, 8, 27, 10, 0);
      expect(DateUtil.getMondayOfWeek(sundayInput)).toEqual(
        new Date(2026, 8, 21, 0, 0, 0, 0),
      );
    });

    it("should return same day for Monday", () => {
      const mondayInput = new Date(2026, 8, 21, 10, 0);
      expect(DateUtil.getMondayOfWeek(mondayInput)).toEqual(
        new Date(2026, 8, 21, 0, 0, 0, 0),
      );
    });

    it("should apply week offset across month boundaries", () => {
      const date = new Date(2026, 8, 30, 12, 0);
      expect(DateUtil.getMondayOfWeek(date, 1)).toEqual(
        new Date(2026, 9, 5, 0, 0, 0, 0),
      );
      expect(DateUtil.getSundayOfWeek(date, -1)).toEqual(
        new Date(2026, 8, 27, 23, 59, 59, 999),
      );
    });
  });

  describe("getThisWeekMonday / getNextWeekMonday", () => {
    afterEach(() => {
      vi.useRealTimers();
    });

    it("should be based on the current date", () => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date(2026, 8, 23, 9, 0));

      expect(DateUtil.getThisWeekMonday()).toEqual(new Date(2026, 8, 21));
      expect(DateUtil.getThisWeekSunday()).toEqual(
        new Date(2026, 8, 27, 23, 59, 59, 999),
      );
      expect(DateUtil.getNextWeekMonday()).toEqual(new Date(2026, 8, 28));
      expect(DateUtil.getNextWeekSunday()).toEqual(
        new Date(2026, 9, 4, 23, 59, 59, 999),
      );
    });

    it("should format week boundaries as local date strings", () => {
      vi.useFakeTimers();
      vi.setSystemTime(new Date(2026, 8, 21, 0, 30));

      expect(DateUtil.getThisWeekMondayDateString()).toBe("2026-09-21");
      expect(DateUtil.getThisWeekSundayDateString()).toBe("2026-09-27");
      expect(DateUtil.getLastWeekMondayDateString()).toBe("2026-09-14");
      expect(DateUtil.getLastWeekSundayDateString()).toBe("2026-09-20");
    });
  });

  describe("parseTimeString", () => {
    it("should parse a valid HH:MM string", () => {
      expect(DateUtil.parseTimeString("08:30")).toBe("08:30");
    });

    it("should parse an hour-only string as null", () => {
      expect(DateUtil.parseTimeString("8")).toBeNull();
    });

    it("should parse a two-digit hour-only string as null", () => {
      expect(DateUtil.parseTimeString("18")).toBeNull();
    });

    it("should accept hour 23 as the maximum", () => {
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

    it("should return null if hour is single-digit", () => {
      expect(DateUtil.parseTimeString("8:30")).toBeNull();
    });

    it("should return null if minute is single-digit", () => {
      expect(DateUtil.parseTimeString("08:3")).toBeNull();
    });

    it("should return null if minute is not two-digit", () => {
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

    it("should skip to the next bookable weekday if the following day is not bookable", () => {
      // 2026-04-24 is a Friday, 2026-04-25/26 are Sat/Sun, 2026-04-27 is the next Monday
      RuntimeConfig.INFOS = { ...RuntimeConfig.INFOS, dailyBasisBooking: true };
      const leave = new Date("2026-04-24T16:59:59.000Z");
      const result = DateUtil.getNextFreeEnterTime(leave, [1, 2, 3, 4, 5]);
      expect(result.toISOString()).toBe("2026-04-27T00:00:00.000Z");
    });

    it("should not skip any day if the following day is already bookable", () => {
      RuntimeConfig.INFOS = { ...RuntimeConfig.INFOS, dailyBasisBooking: true };
      const leave = new Date("2026-04-23T16:59:59.000Z");
      const result = DateUtil.getNextFreeEnterTime(leave, [1, 2, 3, 4, 5]);
      expect(result.toISOString()).toBe("2026-04-24T00:00:00.000Z");
    });

    it("should ignore bookableDays when it is an empty array", () => {
      RuntimeConfig.INFOS = { ...RuntimeConfig.INFOS, dailyBasisBooking: true };
      const leave = new Date("2026-04-24T16:59:59.000Z");
      const result = DateUtil.getNextFreeEnterTime(leave, []);
      expect(result.toISOString()).toBe("2026-04-25T00:00:00.000Z");
    });
  });

  describe("getNextPreferredEnterAndLeaveTime", () => {
    afterEach(() => {
      vi.useRealTimers();
    });

    it("should always return enter with seconds=0/milliseconds=0 and leave with seconds=59/milliseconds=999", () => {
      // "now" has non-zero seconds/milliseconds to make sure they aren't leaked into the result
      vi.useFakeTimers();
      vi.setSystemTime(new Date(2026, 6, 15, 10, 30, 45, 123));

      const combinations: Array<[PreferenceEnterTimeType, boolean]> = [
        [UserPreference.PreferenceEnterTime.Now, false],
        [UserPreference.PreferenceEnterTime.Now, true],
        [UserPreference.PreferenceEnterTime.NextDay, false],
        [UserPreference.PreferenceEnterTime.NextDay, true],
        [UserPreference.PreferenceEnterTime.NextWorkday, false],
        [UserPreference.PreferenceEnterTime.NextWorkday, true],
      ];

      for (const [prefEnterTime, dailyBasisBooking] of combinations) {
        const { enter, leave } = DateUtil.getNextPreferredEnterAndLeaveTime(
          prefEnterTime,
          "00:00",
          "23:59",
          [0, 1, 2, 3, 4, 5, 6],
          dailyBasisBooking,
        );

        expect(enter.getSeconds()).toBe(0);
        expect(enter.getMilliseconds()).toBe(0);
        expect(leave.getSeconds()).toBe(59);
        expect(leave.getMilliseconds()).toBe(999);
      }
    });
  });
});
