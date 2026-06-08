import { describe, it, expect } from "vitest";
import Validation from "./Validation";

describe("Validation", () => {
  describe("isAbsoluteUrl", () => {
    it("should return true for http URLs", () => {
      expect(Validation.isAbsoluteUrl("http://example.com")).toBe(true);
    });

    it("should return true for https URLs", () => {
      expect(Validation.isAbsoluteUrl("https://example.com/path?q=1")).toBe(
        true,
      );
    });

    it("should return true regardless of case (HTTP, HTTPS)", () => {
      expect(Validation.isAbsoluteUrl("HTTP://example.com")).toBe(true);
      expect(Validation.isAbsoluteUrl("HTTPS://example.com")).toBe(true);
    });

    it("should return false for relative URLs", () => {
      expect(Validation.isAbsoluteUrl("/some/path")).toBe(false);
    });

    it("should return false for protocol-relative URLs", () => {
      expect(Validation.isAbsoluteUrl("//example.com")).toBe(false);
    });

    it("should return false for non-http schemes", () => {
      expect(Validation.isAbsoluteUrl("ftp://example.com")).toBe(false);
      expect(Validation.isAbsoluteUrl("mailto:user@example.com")).toBe(false);
    });

    it("should return false for empty string", () => {
      expect(Validation.isAbsoluteUrl("")).toBe(false);
    });

    it("should return false for javascript: string", () => {
      expect(Validation.isAbsoluteUrl("javascript:alert('test')")).toBe(false);
    });
  });
});
