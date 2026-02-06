import { describe, it, expect } from "vitest";
import Formatting from "./Formatting";

describe("Formatting", () => {
  describe("decodeHtmlEntities", () => {
    it("should replace &#x2F; by /", () => {
      expect(Formatting.decodeHtmlEntities("&#x2F;")).toBe("/");
    });
  });
});
