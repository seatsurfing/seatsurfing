import { describe, it, expect } from "vitest";
import RendererUtils from "./RendererUtils";

describe("RendererUtils", () => {
  describe("decodeHtmlEntities", () => {
    it("should replace &#x2F; by /", () => {
      expect(RendererUtils.decodeHtmlEntities("&#x2F;")).toBe("/");
    });
  });
});
