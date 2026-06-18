import { describe, it, expect } from "vitest";
import RendererUtils from "./RendererUtils";

describe("RendererUtils", () => {
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
