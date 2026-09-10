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

    it("should return true for single-character labels", () => {
      expect(Validation.isValidDomain("a.example.com")).toBe(true);
      expect(Validation.isValidDomain("x.io")).toBe(true);
    });

    it("should return false for a missing or too short TLD", () => {
      expect(Validation.isValidDomain("a.b")).toBe(false);
      expect(Validation.isValidDomain("ab.c")).toBe(false);
      expect(Validation.isValidDomain("")).toBe(false);
      expect(Validation.isValidDomain("example")).toBe(false);
      expect(Validation.isValidDomain("example.c")).toBe(false);
      expect(Validation.isValidDomain("example.")).toBe(false);
    });

    it("should return false for seatsurfing.app subdomains", () => {
      expect(Validation.isValidDomain("foo.seatsurfing.app")).toBe(false);
      expect(Validation.isValidDomain("FOO.SEATSURFING.APP")).toBe(false);
    });

    it("should return false for seatsurfing.io subdomains", () => {
      expect(Validation.isValidDomain("foo.seatsurfing.io")).toBe(false);
      expect(Validation.isValidDomain("FOO.SEATSURFING.IO")).toBe(false);
    });

    it("should return false for IP addresses", () => {
      expect(Validation.isValidDomain("169.254.169.254")).toBe(false);
      expect(Validation.isValidDomain("127.0.0.1")).toBe(false);
      expect(Validation.isValidDomain("10.0.0.5")).toBe(false);
      expect(Validation.isValidDomain("::1")).toBe(false);
      expect(Validation.isValidDomain("2001:db8::1")).toBe(false);
    });

    it("should return false for uppercase characters", () => {
      expect(Validation.isValidDomain("Example.com")).toBe(false);
      expect(Validation.isValidDomain("EXAMPLE.COM")).toBe(false);
    });

    it("should return false for a scheme, path, port or query", () => {
      expect(Validation.isValidDomain("http://example.com")).toBe(false);
      expect(Validation.isValidDomain("example.com/path")).toBe(false);
      expect(Validation.isValidDomain("example.com:8080")).toBe(false);
      expect(Validation.isValidDomain("example.com?a=b")).toBe(false);
    });

    it("should return false for invalid characters or whitespace", () => {
      expect(Validation.isValidDomain("foo_bar.example.com")).toBe(false);
      expect(Validation.isValidDomain("foo bar.example.com")).toBe(false);
      expect(Validation.isValidDomain("exämple.com")).toBe(false);
      expect(Validation.isValidDomain(" example.com")).toBe(false);
      expect(Validation.isValidDomain("example.com ")).toBe(false);
    });

    it("should return false for malformed labels", () => {
      expect(Validation.isValidDomain(".example.com")).toBe(false);
      expect(Validation.isValidDomain("example..com")).toBe(false);
      expect(Validation.isValidDomain("-example.com")).toBe(false);
      expect(Validation.isValidDomain(">example.com")).toBe(false);
      expect(Validation.isValidDomain("example-.com")).toBe(false);
    });

    it("should return false for domains longer than 253 characters", () => {
      const longLabel = "a".repeat(60);
      const domain = [
        longLabel,
        longLabel,
        longLabel,
        longLabel,
        longLabel,
        "com",
      ].join(".");
      expect(domain.length).toBeGreaterThan(253);
      expect(Validation.isValidDomain(domain)).toBe(false);
    });
  });

  describe("ROLE_NAME_PATTERN", () => {
    const regex = new RegExp(Validation.ROLE_NAME_PATTERN, "u");

    it("should match valid role names", () => {
      expect(regex.test("Group Manager")).toBe(true);
      expect(regex.test("admin")).toBe(true);
      expect(regex.test("Team_Lead")).toBe(true);
      expect(regex.test("Read-Only")).toBe(true);
      expect(regex.test("Ångström")).toBe(true);
    });

    it("should reject leading or trailing whitespace", () => {
      expect(regex.test(" Manager")).toBe(false);
      expect(regex.test("Manager ")).toBe(false);
      expect(regex.test(" Manager ")).toBe(false);
    });

    it("should reject blank or disallowed-character names", () => {
      expect(regex.test("")).toBe(false);
      expect(regex.test("   ")).toBe(false);
      expect(regex.test("Manager!")).toBe(false);
      expect(regex.test("Team (A)")).toBe(false);
    });
  });

  describe("HUMAN_NAME_PATTERN", () => {
    const regex = new RegExp(Validation.HUMAN_NAME_PATTERN, "u");

    it("should match valid human names", () => {
      expect(regex.test("John")).toBe(true);
      expect(regex.test("Jane Doe")).toBe(true);
      expect(regex.test("O'Brien")).toBe(true);
      expect(regex.test("Anne-Marie")).toBe(true);
      expect(regex.test("St. John")).toBe(true);
      expect(regex.test("Müller")).toBe(true);
      expect(regex.test("山田太郎")).toBe(true);
      expect(regex.test("A")).toBe(true);
    });

    it("should reject leading or trailing whitespace", () => {
      expect(regex.test(" Peter")).toBe(false);
      expect(regex.test("Peter ")).toBe(false);
      expect(regex.test("   Peter   ")).toBe(false);
    });

    it("should reject blank or disallowed-character names", () => {
      expect(regex.test("")).toBe(false);
      expect(regex.test("   ")).toBe(false);
      expect(regex.test("Name<Tag>")).toBe(false);
      expect(regex.test("Name@Domain")).toBe(false);
    });
  });
});
