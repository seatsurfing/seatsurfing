import { describe, it, expect } from "vitest";
import Validation from "./Validation";

describe("Validation", () => {
  describe("isRelativeUrl", () => {
    it("should return true for root-relative paths", () => {
      expect(Validation.isRelativeUrl("/some/path")).toBe(true);
      expect(Validation.isRelativeUrl("/")).toBe(true);
    });

    it("should return false for protocol-relative URLs", () => {
      expect(Validation.isRelativeUrl("//example.com")).toBe(false);
    });

    it("should return false for backslash/control-character URLs", () => {
      expect(Validation.isRelativeUrl("/\\example.com")).toBe(false);
      expect(Validation.isRelativeUrl("/\\\\example.com")).toBe(false);
      expect(Validation.isRelativeUrl("/\nexample")).toBe(false);
    });

    it("should return false for absolute URLs", () => {
      expect(Validation.isRelativeUrl("http://example.com")).toBe(false);
      expect(Validation.isRelativeUrl("https://example.com")).toBe(false);
    });

    it("should return false for non-http schemes", () => {
      expect(Validation.isRelativeUrl("ftp://example.com")).toBe(false);
      expect(Validation.isRelativeUrl("mailto:user@example.com")).toBe(false);
    });

    it("should return false for path-relative URLs", () => {
      expect(Validation.isRelativeUrl("relative/path")).toBe(false);
      expect(Validation.isRelativeUrl("../path")).toBe(false);
    });

    it("should return false for empty string", () => {
      expect(Validation.isRelativeUrl("")).toBe(false);
    });
  });

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

    it("should return true for protocol-relative URLs", () => {
      expect(Validation.isAbsoluteUrl("//example.com")).toBe(true);
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

  describe("isValidDomain", () => {
    it("should return true for valid domains", () => {
      expect(Validation.isValidDomain("example.com")).toBe(true);
      expect(Validation.isValidDomain("sub.example.com")).toBe(true);
      expect(Validation.isValidDomain("example.co.uk")).toBe(true);
    });

    it("should return false when the dot is too early in the string", () => {
      expect(Validation.isValidDomain("a.b")).toBe(false);
      expect(Validation.isValidDomain("ab.c")).toBe(false);
      expect(Validation.isValidDomain("")).toBe(false);
    });

    it("should return false for seatsurfing.app subdomains", () => {
      expect(Validation.isValidDomain("foo.seatsurfing.app")).toBe(false);
      expect(Validation.isValidDomain("FOO.SEATSURFING.APP")).toBe(false);
    });

    it("should return false for seatsurfing.io subdomains", () => {
      expect(Validation.isValidDomain("foo.seatsurfing.io")).toBe(false);
      expect(Validation.isValidDomain("FOO.SEATSURFING.IO")).toBe(false);
    });

    it("should return false when the last dot is too close to the end", () => {
      expect(Validation.isValidDomain("example.c")).toBe(false);
      expect(Validation.isValidDomain("example.")).toBe(false);
    });
  });
});
