import { describe, it, expect, beforeEach, afterEach } from "vitest";
import Formatting from "./Formatting";

describe("Formatting", () => {
  describe("stripTimezoneDetails", () => {
    it("replaces positive UTC offset with .000Z", () => {
      expect(Formatting.stripTimezoneDetails("2024-06-18T10:30:00+02:00")).toBe(
        "2024-06-18T10:30:00.000Z",
      );
    });

    it("supports date strings with microseconds", () => {
      expect(
        Formatting.stripTimezoneDetails("2026-06-19T07:10:02.611091+02:00"),
      ).toBe("2026-06-19T07:10:02.611091Z");
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
