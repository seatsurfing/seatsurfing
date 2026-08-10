import { describe, it, expect } from "vitest";
import RendererUtils from "./RendererUtils";

describe("RendererUtils", () => {
  describe("fullname", () => {
    it("should return firstname and lastname if both are set", () => {
      expect(RendererUtils.fullname("John", "Doe")).toBe("John Doe");
    });
    it("should return only firstname if lastname is missing", () => {
      expect(RendererUtils.fullname("John", "")).toBe("John");
    });
    it("should return only lastname if firstname is missing", () => {
      expect(RendererUtils.fullname("", "Doe")).toBe("Doe");
    });
    it("should return fallback if firstname and lastname are missing", () => {
      expect(RendererUtils.fullname("", "", "fallback")).toBe("fallback");
    });
    it("should return empty string if firstname, lastname and fallback are missing", () => {
      expect(RendererUtils.fullname("", "")).toBe("");
    });
  });

  describe("escapeHtml", () => {
    it("should escape <, >, &, \", '", () => {
      expect(RendererUtils.escapeHtml('<script>alert("x\'y")</script>&')).toBe(
        "&lt;script&gt;alert(&quot;x&#039;y&quot;)&lt;/script&gt;&amp;",
      );
    });
    it("should return empty string unchanged", () => {
      expect(RendererUtils.escapeHtml("")).toBe("");
    });
  });

  describe("decodeHtmlEntities", () => {
    it("should replace &#x2F; by /", () => {
      expect(RendererUtils.decodeHtmlEntities("&#x2F;")).toBe("/");
    });
  });
});
