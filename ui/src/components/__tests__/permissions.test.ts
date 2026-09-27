import { describe, it, expect, beforeEach, afterEach } from "vitest";
import RuntimeConfig from "../RuntimeConfig";
import { Permission, PermissionLevel } from "@/types/Permission";

/**
 * The permission checks that decide what the administration UI offers.
 *
 * These govern presentation only: the server checks every request
 * independently, so these tests are about showing the right things, not about
 * enforcement.
 */
describe("RuntimeConfig permission checks", () => {
  beforeEach(() => {
    RuntimeConfig.INFOS.permissions = {};
  });

  describe("hasPermission", () => {
    it("returns false when the permission is absent", () => {
      expect(
        RuntimeConfig.hasPermission(Permission.Groups, PermissionLevel.Read),
      ).toBe(false);
    });

    it("returns true at exactly the required level", () => {
      RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Read };
      expect(
        RuntimeConfig.hasPermission(Permission.Groups, PermissionLevel.Read),
      ).toBe(true);
    });

    it("returns true above the required level", () => {
      RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Admin };
      expect(
        RuntimeConfig.hasPermission(Permission.Groups, PermissionLevel.Write),
      ).toBe(true);
    });

    it("returns false below the required level", () => {
      RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Write };
      expect(
        RuntimeConfig.hasPermission(Permission.Groups, PermissionLevel.Admin),
      ).toBe(false);
    });

    it("defaults to requiring the admin level", () => {
      RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Read };
      expect(RuntimeConfig.hasPermission(Permission.Groups)).toBe(false);
      RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Admin };
      expect(RuntimeConfig.hasPermission(Permission.Groups)).toBe(true);
    });

    it("does not let one permission satisfy another", () => {
      RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Admin };
      expect(
        RuntimeConfig.hasPermission(Permission.Users, PermissionLevel.Read),
      ).toBe(false);
    });
  });

  describe("hasAnyPermission", () => {
    it("is false for a user with no administrative access", () => {
      expect(RuntimeConfig.hasAnyPermission()).toBe(false);
    });

    it("is true given any granted permission", () => {
      RuntimeConfig.INFOS.permissions = { audit_log: PermissionLevel.Read };
      expect(RuntimeConfig.hasAnyPermission()).toBe(true);
    });

    it("ignores entries explicitly granted at the none level", () => {
      RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.None };
      expect(RuntimeConfig.hasAnyPermission()).toBe(false);
    });
  });

  describe("canSeePluginMenuItem", () => {
    it("honours a permission declared by the plugin", () => {
      const item = {
        requiredPermission: "plugin.plus.scim",
        requiredLevel: PermissionLevel.Admin,
        visibility: "admin",
      };
      expect(RuntimeConfig.canSeePluginMenuItem(item, "admin")).toBe(false);
      RuntimeConfig.INFOS.permissions = {
        "plugin.plus.scim": PermissionLevel.Admin,
      };
      expect(RuntimeConfig.canSeePluginMenuItem(item, "admin")).toBe(true);
    });

    it("honours any one of several alternative permissions", () => {
      const item = {
        requiredPermissionsAny: [
          "plugin.plus.exchange",
          "plugin.plus.scim",
          "plugin.plus.msteams",
        ],
        requiredLevel: PermissionLevel.Admin,
        visibility: "admin",
      };
      RuntimeConfig.INFOS.permissions = {};
      expect(RuntimeConfig.canSeePluginMenuItem(item, "admin")).toBe(false);

      RuntimeConfig.INFOS.permissions = {
        "plugin.plus.msteams": PermissionLevel.Admin,
      };
      expect(RuntimeConfig.canSeePluginMenuItem(item, "admin")).toBe(true);
    });

    it("falls back to the old visibility for plugins that declare nothing", () => {
      const adminItem = { visibility: "admin" };
      const spaceAdminItem = { visibility: "spaceadmin" };

      RuntimeConfig.INFOS.permissions = { audit_log: PermissionLevel.Read };
      // Some administrative access, but not settings administration.
      expect(RuntimeConfig.canSeePluginMenuItem(adminItem, "admin")).toBe(
        false,
      );
      expect(
        RuntimeConfig.canSeePluginMenuItem(spaceAdminItem, "spaceadmin"),
      ).toBe(true);

      RuntimeConfig.INFOS.permissions = { org_settings: PermissionLevel.Admin };
      expect(RuntimeConfig.canSeePluginMenuItem(adminItem, "admin")).toBe(true);
    });

    it("keeps items out of the section they were not declared for", () => {
      RuntimeConfig.INFOS.permissions = { org_settings: PermissionLevel.Admin };
      expect(
        RuntimeConfig.canSeePluginMenuItem(
          { visibility: "admin" },
          "spaceadmin",
        ),
      ).toBe(false);
    });
  });

  describe("canSeeBookingUIIntegration", () => {
    it("is visible to everyone when the plugin declares no permission", () => {
      expect(RuntimeConfig.canSeeBookingUIIntegration({})).toBe(true);
    });

    it("honours a permission declared by the plugin", () => {
      const item = {
        requiredPermission: "plugin.plus.exchange",
        requiredLevel: PermissionLevel.Read,
      };
      expect(RuntimeConfig.canSeeBookingUIIntegration(item)).toBe(false);
      RuntimeConfig.INFOS.permissions = {
        "plugin.plus.exchange": PermissionLevel.Read,
      };
      expect(RuntimeConfig.canSeeBookingUIIntegration(item)).toBe(true);
    });

    it("honours any one of several alternative permissions", () => {
      const item = {
        requiredPermissionsAny: ["plugin.plus.exchange", "plugin.plus.scim"],
        requiredLevel: PermissionLevel.Admin,
      };
      expect(RuntimeConfig.canSeeBookingUIIntegration(item)).toBe(false);
      RuntimeConfig.INFOS.permissions = {
        "plugin.plus.scim": PermissionLevel.Admin,
      };
      expect(RuntimeConfig.canSeeBookingUIIntegration(item)).toBe(true);
    });
  });

  describe("pickBookingUIIntegrationTitle", () => {
    afterEach(() => {
      window.localStorage.removeItem("next-export-i18n-lang");
    });

    it("falls back to the plain title when no titles map is given", () => {
      expect(
        RuntimeConfig.pickBookingUIIntegrationTitle({ title: "Chat" }),
      ).toBe("Chat");
    });

    it("prefers the exact language", () => {
      window.localStorage.setItem("next-export-i18n-lang", "de");
      expect(
        RuntimeConfig.pickBookingUIIntegrationTitle({
          title: "Chat",
          titles: { de: "Chatten", "en-GB": "Chat (UK)" },
        }),
      ).toBe("Chatten");
    });

    it("falls back to the base language for a regional code", () => {
      window.localStorage.setItem("next-export-i18n-lang", "de-CH");
      expect(
        RuntimeConfig.pickBookingUIIntegrationTitle({
          title: "Chat",
          titles: { de: "Chatten" },
        }),
      ).toBe("Chatten");
    });

    it("falls back to the plain title when the language is not in the map", () => {
      window.localStorage.setItem("next-export-i18n-lang", "fr");
      expect(
        RuntimeConfig.pickBookingUIIntegrationTitle({
          title: "Chat",
          titles: { de: "Chatten" },
        }),
      ).toBe("Chat");
    });
  });
});
