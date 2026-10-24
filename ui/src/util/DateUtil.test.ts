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

  describe("convertFromFakeUTCDate", () => {
    it("should map UTC components to local components", () => {
      const fakeUTC = new Date(Date.UTC(2030, 8, 7, 22, 15, 30, 0));
      const local = DateUtil.convertFromFakeUTCDate(fakeUTC);

      expect(local.getFullYear()).toBe(2030);
      expect(local.getMonth()).toBe(8);
      expect(local.getDate()).toBe(7);
      expect(local.getHours()).toBe(22);
      expect(local.getMinutes()).toBe(15);
      expect(local.getSeconds()).toBe(30);
    });

    it("should be the inverse of convertToFakeUTCDate", () => {
      const local = new Date(2030, 0, 1, 23, 59, 59, 0);
      const roundTrip = DateUtil.convertFromFakeUTCDate(
        DateUtil.convertToFakeUTCDate(local),
      );

      expect(roundTrip.getTime()).toBe(local.getTime());
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

  describe("isInPeriod", () => {
    it("should not block without a period", () => {
      expect(DateUtil.isInPeriod(new Date(2030, 8, 4), "", "")).toBe(false);
    });

    it("should block inclusive start and end days", () => {
      const from = "2030-09-04";
      const until = "2030-09-06";
      expect(DateUtil.isInPeriod(new Date(2030, 8, 3), from, until)).toBe(
        false,
      );
      expect(DateUtil.isInPeriod(new Date(2030, 8, 4, 0, 0), from, until)).toBe(
        true,
      );
      expect(
        DateUtil.isInPeriod(new Date(2030, 8, 6, 23, 59), from, until),
      ).toBe(true);
      expect(DateUtil.isInPeriod(new Date(2030, 8, 7), from, until)).toBe(
        false,
      );
    });

    it("should treat a missing bound as open", () => {
      expect(DateUtil.isInPeriod(new Date(2099, 0, 1), "2030-09-04", "")).toBe(
        true,
      );
      expect(DateUtil.isInPeriod(new Date(2030, 8, 3), "2030-09-04", "")).toBe(
        false,
      );
      expect(DateUtil.isInPeriod(new Date(2020, 0, 1), "", "2030-09-06")).toBe(
        true,
      );
      expect(DateUtil.isInPeriod(new Date(2030, 8, 7), "", "2030-09-06")).toBe(
        false,
      );
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

  describe("resolveLinkEnterAndLeave", () => {
    const defaultEnter = new Date(2030, 5, 11, 9, 0, 0, 0);
    const defaultLeave = new Date(2030, 5, 11, 17, 0, 0, 0);
    const resolve = (
      enterParam: string | undefined,
      leaveParam: string | undefined = undefined,
      dailyBasisBooking = false,
      now = new Date(2030, 5, 7, 10, 23, 45, 500),
    ) =>
      DateUtil.resolveLinkEnterAndLeave(
        enterParam,
        leaveParam,
        defaultEnter,
        defaultLeave,
        "17:00",
        dailyBasisBooking,
        now,
      );

    it("keeps the defaults without an enter parameter", () => {
      const { enter, leave } = resolve(undefined);
      expect(enter).toEqual(defaultEnter);
      expect(leave).toEqual(defaultLeave);
    });

    it("keeps the defaults for an invalid enter parameter", () => {
      const { enter, leave } = resolve("tomorrow");
      expect(enter).toEqual(defaultEnter);
      expect(leave).toEqual(defaultLeave);
    });

    it("uses a valid leave parameter without an enter parameter", () => {
      const { enter, leave } = resolve(undefined, "2030-06-11T12:00");
      expect(enter).toEqual(defaultEnter);
      expect(leave).toEqual(new Date(2030, 5, 11, 12, 0));
    });

    it("sets enter to the current time for 'now' and leave to workday end", () => {
      const { enter, leave } = resolve("now");
      expect(enter).toEqual(new Date(2030, 5, 7, 10, 23, 0, 0));
      expect(leave).toEqual(new Date(2030, 5, 7, 17, 0, 0, 0));
    });

    it("uses the default duration if 'now' is past the workday end", () => {
      const { enter, leave } = resolve(
        "now",
        undefined,
        false,
        new Date(2030, 5, 7, 17, 30),
      );
      expect(enter).toEqual(new Date(2030, 5, 7, 17, 30));
      expect(leave).toEqual(new Date(2030, 5, 7, 23, 59, 59, 999));
    });

    it("uses a datetime enter parameter and adjusts leave", () => {
      const { enter, leave } = resolve("2030-06-20T13:30");
      expect(enter).toEqual(new Date(2030, 5, 20, 13, 30));
      expect(leave).toEqual(new Date(2030, 5, 20, 17, 0));
    });

    it("keeps a valid leave parameter after the enter parameter", () => {
      const { enter, leave } = resolve("2030-06-20T13:30", "2030-06-20T15:00");
      expect(enter).toEqual(new Date(2030, 5, 20, 13, 30));
      expect(leave).toEqual(new Date(2030, 5, 20, 15, 0));
    });

    it("ignores a leave parameter that is not after enter", () => {
      const { leave } = resolve("2030-06-20T13:30", "2030-06-20T08:00");
      expect(leave).toEqual(new Date(2030, 5, 20, 17, 0));
    });

    it("books the whole day with daily basis booking", () => {
      const { enter, leave } = resolve("now", undefined, true);
      expect(enter).toEqual(new Date(2030, 5, 7, 0, 0, 0, 0));
      expect(leave).toEqual(new Date(2030, 5, 7, 23, 59, 59, 999));
    });
  });
});
