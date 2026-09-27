import { describe, it, expect } from "vitest";
import FrameProtection from "./FrameProtection";

describe("FrameProtection", () => {
  describe("isNonFramablePath", () => {
    it("should protect the login form and IdP selection", () => {
      expect(FrameProtection.isNonFramablePath("/login")).toBe(true);
      expect(FrameProtection.isNonFramablePath("/login/")).toBe(true);
      expect(FrameProtection.isNonFramablePath("/login?redir=%2Fsearch")).toBe(
        true,
      );
    });

    it("should protect password reset, set password and public booking pages", () => {
      expect(FrameProtection.isNonFramablePath("/resetpw")).toBe(true);
      expect(FrameProtection.isNonFramablePath("/resetpw/[id]")).toBe(true);
      expect(FrameProtection.isNonFramablePath("/setpw/[id]")).toBe(true);
      expect(FrameProtection.isNonFramablePath("/book")).toBe(true);
      expect(
        FrameProtection.isNonFramablePath("/book/details/[externalId]"),
      ).toBe(true);
      expect(FrameProtection.isNonFramablePath("/book/confirm/[id]")).toBe(
        true,
      );
    });

    it("should not protect pages used by the MS Teams and Confluence integrations", () => {
      expect(FrameProtection.isNonFramablePath("/login/success/[id]")).toBe(
        false,
      );
      expect(FrameProtection.isNonFramablePath("/login/failed")).toBe(false);
      expect(FrameProtection.isNonFramablePath("/")).toBe(false);
      expect(FrameProtection.isNonFramablePath("/search")).toBe(false);
      expect(FrameProtection.isNonFramablePath("/bookings")).toBe(false);
      expect(FrameProtection.isNonFramablePath("/preferences")).toBe(false);
    });

    it("should not match paths that only share a prefix", () => {
      expect(FrameProtection.isNonFramablePath("/loginx")).toBe(false);
      expect(FrameProtection.isNonFramablePath("/bookings")).toBe(false);
      expect(FrameProtection.isNonFramablePath("/resetpwx")).toBe(false);
    });
  });
});
