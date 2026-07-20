import { describe, it, expect, beforeEach } from "vitest";
import SearchUtil from "./SearchUtil";
import DateUtil from "./DateUtil";
import RuntimeConfig from "@/components/RuntimeConfig";

describe("SearchUtil", () => {
  describe("getSearchHint", () => {
    const t = (key: string) => key;
    const locationId = "location-1";
    const enter = new Date(2026, 6, 20, 9, 0, 0, 0); // future, within advance/duration limits
    const leave = new Date(2026, 6, 20, 12, 0, 0, 0); // 3h duration

    beforeEach(() => {
      RuntimeConfig.resetInfos();
      RuntimeConfig.INFOS.maxBookingsPerUser = 5;
      RuntimeConfig.INFOS.maxDaysInAdvance = 365;
      RuntimeConfig.INFOS.maxBookingDurationHours = 8;
      RuntimeConfig.INFOS.minBookingDurationHours = 1;
    });

    it("should return no hint for a valid selection", () => {
      const result = SearchUtil.getSearchHint(enter, leave, t, 0, locationId);
      expect(result).toBe("");
    });

    it("should still enforce the booking limit for a plain space admin without noAdminRestrictions", () => {
      RuntimeConfig.INFOS.spaceAdmin = true;
      RuntimeConfig.INFOS.noAdminRestrictions = false;
      RuntimeConfig.INFOS.maxBookingsPerUser = 1;

      const result = SearchUtil.getSearchHint(enter, leave, t, 1, locationId);
      expect(result).toBe("errorBookingLimit");
    });

    it("should still enforce the booking limit when noAdminRestrictions is on but the user is not a space admin", () => {
      // bugfix regression: the previous check compared noAdminRestrictions against
      // the User.UserRoleSpaceAdmin constant (always truthy) instead of the actual
      // user's spaceAdmin flag, so restrictions were bypassed for every user.
      RuntimeConfig.INFOS.spaceAdmin = false;
      RuntimeConfig.INFOS.noAdminRestrictions = true;
      RuntimeConfig.INFOS.maxBookingsPerUser = 1;

      const result = SearchUtil.getSearchHint(enter, leave, t, 1, locationId);
      expect(result).toBe("errorBookingLimit");
    });

    it("should bypass the booking limit for a space admin when noAdminRestrictions is on", () => {
      RuntimeConfig.INFOS.spaceAdmin = true;
      RuntimeConfig.INFOS.noAdminRestrictions = true;
      RuntimeConfig.INFOS.maxBookingsPerUser = 1;

      const result = SearchUtil.getSearchHint(enter, leave, t, 1, locationId);
      expect(result).toBe("");
    });

    it("should bypass the max booking duration restriction for a space admin when noAdminRestrictions is on", () => {
      RuntimeConfig.INFOS.spaceAdmin = true;
      RuntimeConfig.INFOS.noAdminRestrictions = true;
      RuntimeConfig.INFOS.maxBookingDurationHours = 1; // selection is 3h, would normally fail

      const result = SearchUtil.getSearchHint(enter, leave, t, 0, locationId);
      expect(result).toBe("");
    });

    it("should still return errorPickArea for a space admin with noAdminRestrictions when no area is selected", () => {
      RuntimeConfig.INFOS.spaceAdmin = true;
      RuntimeConfig.INFOS.noAdminRestrictions = true;

      const result = SearchUtil.getSearchHint(enter, leave, t, 0, "");
      expect(result).toBe("errorPickArea");
    });
  });

  describe("calculateNewEnterAndLeave", () => {
    const currentEnter = new Date(2026, 6, 15, 9, 0, 0, 0);
    const currentLeave = new Date(2026, 6, 15, 17, 0, 0, 0);

    it("should return null if enter and leave are both null", () => {
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        false,
        false,
        "00:00",
        null,
        null,
      );
      expect(result).toBeNull();
    });

    it("should return null if enter and leave are unchanged", () => {
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        false,
        false,
        "00:00",
        new Date(currentEnter),
        new Date(currentLeave),
      );
      expect(result).toBeNull();
    });

    it("should return null if only enter is given and unchanged", () => {
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        false,
        false,
        "00:00",
        new Date(currentEnter),
        null,
      );
      expect(result).toBeNull();
    });

    it("should return null if only leave is given and unchanged", () => {
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        false,
        false,
        "00:00",
        null,
        new Date(currentLeave),
      );
      expect(result).toBeNull();
    });

    it("should adopt both enter and leave as given when both change", () => {
      const enter = new Date(2026, 6, 16, 9, 0, 0, 0);
      const leave = new Date(2026, 6, 16, 17, 0, 0, 0);
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        false,
        false,
        "00:00",
        enter,
        leave,
      );
      expect(result?.newEnter).toEqual(enter);
      expect(result?.newLeave).toEqual(leave);
    });

    it("should shift leave by the enter/leave diff when only enter changes and leave stays on the same day", () => {
      // diff is 8h, moving enter by one day keeps leave on the same (new) day
      const enter = new Date(2026, 6, 16, 9, 0, 0, 0);
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        false,
        false,
        "00:00",
        enter,
        null,
      );
      expect(result?.newLeave).toEqual(new Date(2026, 6, 16, 17, 0, 0, 0));
    });

    it("bugfix: should clamp leave to end of enter's day if the diff-based leave would roll over to the next day and multi-day selection is disabled", () => {
      // enter stays on 2026-07-15 but moves close to midnight so that
      // adding the current enter/leave diff (1h) rolls the naive leave into 2026-07-16
      const shortEnter = new Date(2026, 6, 15, 9, 0, 0, 0);
      const shortLeave = new Date(2026, 6, 15, 10, 0, 0, 0); // diff = 1h
      const enter = new Date(2026, 6, 15, 23, 30, 0, 0);

      const result = SearchUtil.calculateNewEnterAndLeave(
        shortEnter,
        shortLeave,
        false, // selectionMultiDay
        false,
        false,
        "00:00",
        enter,
        null,
      );

      expect(result?.newEnter).toEqual(enter);
      expect(result?.newLeave).toEqual(DateUtil.setHoursToMax(enter));
      expect(DateUtil.isSameDay(result!.newLeave!, enter)).toBe(true);
    });

    it("should NOT clamp leave when it rolls over to the next day and multi-day selection is enabled", () => {
      const shortEnter = new Date(2026, 6, 15, 9, 0, 0, 0);
      const shortLeave = new Date(2026, 6, 15, 10, 0, 0, 0); // diff = 1h
      const enter = new Date(2026, 6, 15, 23, 30, 0, 0);

      const result = SearchUtil.calculateNewEnterAndLeave(
        shortEnter,
        shortLeave,
        true, // selectionMultiDay
        false,
        false,
        "00:00",
        enter,
        null,
      );

      expect(result?.newLeave).toEqual(new Date(2026, 6, 16, 0, 30, 0, 0));
    });

    it("should adopt leave as given when only leave changes", () => {
      const leave = new Date(2026, 6, 15, 18, 0, 0, 0);
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        false,
        false,
        "00:00",
        null,
        leave,
      );
      expect(result?.newEnter).toBeUndefined();
      expect(result?.newLeave).toEqual(leave);
    });

    it("should clamp enter/leave to day min/max when dailyBasisBooking is active", () => {
      const enter = new Date(2026, 6, 16, 9, 0, 0, 0);
      const leave = new Date(2026, 6, 16, 17, 0, 0, 0);
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        true, // dailyBasisBooking
        false,
        "00:00",
        enter,
        leave,
      );
      expect(result?.newEnter).toEqual(DateUtil.setHoursToMin(enter));
      expect(result?.newLeave).toEqual(DateUtil.setHoursToMax(leave));
    });

    it("should set enter to preferred workday start when the date changes with unchanged time and the new date is in the future", () => {
      const enter = new Date();
      enter.setDate(enter.getDate() + 5);
      enter.setHours(9, 0, 0, 0);
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        false,
        true, // autoUpdateEnterTimeToPrefWorkdayStart
        "08:15",
        enter,
        null,
      );
      expect(result?.newEnter).toEqual(
        DateUtil.setTimeFromTimeString(enter, "08:15"),
      );
      expect(result?.autoUpdateEnterTimeToPrefWorkdayStart).toBe(true);
    });

    it("should turn off autoUpdateEnterTimeToPrefWorkdayStart once the user changes the time themselves", () => {
      const enter = new Date(2026, 6, 15, 11, 0, 0, 0); // same day, different time
      const result = SearchUtil.calculateNewEnterAndLeave(
        currentEnter,
        currentLeave,
        false,
        false,
        true, // autoUpdateEnterTimeToPrefWorkdayStart
        "08:15",
        enter,
        null,
      );
      expect(result?.autoUpdateEnterTimeToPrefWorkdayStart).toBe(false);
    });
  });
});
