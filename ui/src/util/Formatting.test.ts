import { describe, it, expect, beforeEach, afterEach } from "vitest";
import Formatting from "./Formatting";
import RuntimeConfig from "@/components/RuntimeConfig";

describe("Formatting", () => {
  let infosSnapshot: typeof RuntimeConfig.INFOS;

  beforeEach(() => {
    infosSnapshot = { ...RuntimeConfig.INFOS };
  });

  afterEach(() => {
    RuntimeConfig.INFOS = infosSnapshot;
  });
  describe("getFormatterDate", () => {
    const date = new Date(Date.UTC(2026, 6, 4, 11, 41));

    it("formats using the user's dateFormat preference", () => {
      RuntimeConfig.INFOS = { ...RuntimeConfig.INFOS, dateFormat: "d.m.Y" };
      expect(Formatting.getFormatterDate().format(date)).toBe("04.07.2026");
    });

    it("reorders parts for a different dateFormat preference", () => {
      RuntimeConfig.INFOS = { ...RuntimeConfig.INFOS, dateFormat: "m/d/Y" };
      expect(Formatting.getFormatterDate().format(date)).toBe("07/04/2026");
    });
  });

  describe("getFormatterShort", () => {
    const date = new Date(Date.UTC(2026, 6, 4, 11, 41));

    it("combines the dateFormat preference with the time", () => {
      RuntimeConfig.INFOS = {
        ...RuntimeConfig.INFOS,
        dateFormat: "d/m/Y",
        use24HourTime: true,
      };
      expect(Formatting.getFormatterShort().format(date)).toBe(
        "04/07/2026, 11:41",
      );
    });
  });

  describe("formatRange", () => {
    const enter = new Date(Date.UTC(2026, 6, 4, 9, 0));
    const leaveSameDay = new Date(Date.UTC(2026, 6, 4, 17, 30));
    const leaveNextDay = new Date(Date.UTC(2026, 6, 5, 12, 0));

    beforeEach(() => {
      RuntimeConfig.INFOS = {
        ...RuntimeConfig.INFOS,
        dateFormat: "Y-m-d",
        use24HourTime: true,
        dailyBasisBooking: false,
      };
    });

    it("shows the date only once for a same-day range", () => {
      expect(Formatting.formatRange(enter, leaveSameDay)).toBe(
        "2026-07-04, 09:00 – 17:30",
      );
    });

    it("shows both dates for a multi-day range", () => {
      expect(Formatting.formatRange(enter, leaveNextDay)).toBe(
        "2026-07-04, 09:00 – 2026-07-05, 12:00",
      );
    });

    it("prefixes the weekday when requested", () => {
      Formatting.Language = "en";
      expect(Formatting.formatRange(enter, leaveSameDay, true)).toBe(
        "Saturday, 2026-07-04, 09:00 – 17:30",
      );
    });

    it("omits times for daily-basis bookings", () => {
      RuntimeConfig.INFOS = { ...RuntimeConfig.INFOS, dailyBasisBooking: true };
      expect(Formatting.formatRange(enter, leaveSameDay)).toBe("2026-07-04");
      expect(Formatting.formatRange(enter, leaveNextDay)).toBe(
        "2026-07-04 – 2026-07-05",
      );
    });
  });

  describe("stripTimezoneDetails", () => {
    it("replaces positive UTC offset with .000Z", () => {
      expect(Formatting.stripTimezoneDetails("2024-06-18T10:30:00+02:00")).toBe(
        "2024-06-18T10:30:00.000Z",
      );
    });

    it("truncates fractional seconds to milliseconds", () => {
      expect(
        Formatting.stripTimezoneDetails("2026-06-19T07:10:02.611091+02:00"),
      ).toBe("2026-06-19T07:10:02.611Z");
    });

    it("replaces negative UTC offset with .000Z", () => {
      expect(Formatting.stripTimezoneDetails("2024-06-18T10:30:00-05:00")).toBe(
        "2024-06-18T10:30:00.000Z",
      );
    });

    it("leaves strings without timezone offset unchanged", () => {
      expect(Formatting.stripTimezoneDetails("2024-06-18T10:30:00")).toBe(
        "2024-06-18T10:30:00",
      );
    });

    it("leaves strings already ending in Z unchanged", () => {
      expect(Formatting.stripTimezoneDetails("2024-06-18T10:30:00.000Z")).toBe(
        "2024-06-18T10:30:00.000Z",
      );
    });

    it("leaves short strings unchanged", () => {
      expect(Formatting.stripTimezoneDetails("hello")).toBe("hello");
    });

    it("leaves empty string unchanged", () => {
      expect(Formatting.stripTimezoneDetails("")).toBe("");
    });
  });
});
