import { describe, it, expect } from "vitest";

/**
 * Logic tests for the "cannot change own role" feature on the Edit User page.
 *
 * The EditUser component ([id].tsx) determines whether the role field is
 * editable based on two conditions:
 *   1. The current user must NOT be editing themselves (isOwnUser === false).
 *   2. The admin's role must be >= the target user's current role.
 *
 * If either condition fails, the role is shown as read-only text.
 * When the user IS editing themselves, a hint (cannotChangeOwnRole) is shown.
 *
 * These tests verify the pure logic without needing to render the full page.
 */

// Role constants mirroring User.ts
const UserRoleUser = 0;
const UserRoleSpaceAdmin = 10;
const UserRoleOrgAdmin = 20;
const UserRoleServiceAccountRO = 21;
const UserRoleServiceAccountRW = 22;
const UserRoleSuperAdmin = 90;

interface RoleSelectConditions {
  adminUserId: string;
  editedUserId: string;
  adminUserRole: number;
  editedUserRole: number;
}

/**
 * Determines whether the role dropdown is editable (true) or read-only (false).
 * This mirrors the logic in EditUser.render():
 *   `if (!isOwnUser && this.adminUserRole >= this.state.role)`
 */
function isRoleEditable(conditions: RoleSelectConditions): boolean {
  const isOwnUser = conditions.adminUserId === conditions.editedUserId;
  if (isOwnUser) return false;
  return conditions.adminUserRole >= conditions.editedUserRole;
}

/**
 * Determines whether the "cannot change own role" hint is shown.
 * The hint is shown only when the user is editing themselves AND the
 * role is read-only (i.e. the else branch is entered).
 */
function showCannotChangeOwnRoleHint(
  conditions: RoleSelectConditions,
): boolean {
  const isOwnUser = conditions.adminUserId === conditions.editedUserId;
  return isOwnUser;
}

describe("EditUser role select logic", () => {
  describe("isRoleEditable", () => {
    it("returns false when editing own user (OrgAdmin editing self)", () => {
      expect(
        isRoleEditable({
          adminUserId: "user-1",
          editedUserId: "user-1",
          adminUserRole: UserRoleOrgAdmin,
          editedUserRole: UserRoleOrgAdmin,
        }),
      ).toBe(false);
    });

    it("returns false when editing own user (SuperAdmin editing self)", () => {
      expect(
        isRoleEditable({
          adminUserId: "user-1",
          editedUserId: "user-1",
          adminUserRole: UserRoleSuperAdmin,
          editedUserRole: UserRoleSuperAdmin,
        }),
      ).toBe(false);
    });

    it("returns true when OrgAdmin edits another user with lower role", () => {
      expect(
        isRoleEditable({
          adminUserId: "admin-1",
          editedUserId: "user-2",
          adminUserRole: UserRoleOrgAdmin,
          editedUserRole: UserRoleUser,
        }),
      ).toBe(true);
    });

    it("returns true when OrgAdmin edits another user with equal role", () => {
      expect(
        isRoleEditable({
          adminUserId: "admin-1",
          editedUserId: "admin-2",
          adminUserRole: UserRoleOrgAdmin,
          editedUserRole: UserRoleOrgAdmin,
        }),
      ).toBe(true);
    });

    it("returns true when OrgAdmin edits a SpaceAdmin", () => {
      expect(
        isRoleEditable({
          adminUserId: "admin-1",
          editedUserId: "user-2",
          adminUserRole: UserRoleOrgAdmin,
          editedUserRole: UserRoleSpaceAdmin,
        }),
      ).toBe(true);
    });

    it("returns false when SpaceAdmin edits an OrgAdmin (higher role)", () => {
      expect(
        isRoleEditable({
          adminUserId: "space-admin-1",
          editedUserId: "org-admin-1",
          adminUserRole: UserRoleSpaceAdmin,
          editedUserRole: UserRoleOrgAdmin,
        }),
      ).toBe(false);
    });

    it("returns true when SuperAdmin edits another SuperAdmin", () => {
      expect(
        isRoleEditable({
          adminUserId: "super-1",
          editedUserId: "super-2",
          adminUserRole: UserRoleSuperAdmin,
          editedUserRole: UserRoleSuperAdmin,
        }),
      ).toBe(true);
    });

    it("returns false when OrgAdmin edits a ServiceAccountRO (higher numeric role)", () => {
      expect(
        isRoleEditable({
          adminUserId: "admin-1",
          editedUserId: "sa-1",
          adminUserRole: UserRoleOrgAdmin,
          editedUserRole: UserRoleServiceAccountRO,
        }),
      ).toBe(false);
    });

    it("returns false when OrgAdmin edits a ServiceAccountRW (higher numeric role)", () => {
      expect(
        isRoleEditable({
          adminUserId: "admin-1",
          editedUserId: "sa-2",
          adminUserRole: UserRoleOrgAdmin,
          editedUserRole: UserRoleServiceAccountRW,
        }),
      ).toBe(false);
    });
  });

  describe("showCannotChangeOwnRoleHint", () => {
    it("returns true when editing own user", () => {
      expect(
        showCannotChangeOwnRoleHint({
          adminUserId: "user-1",
          editedUserId: "user-1",
          adminUserRole: UserRoleOrgAdmin,
          editedUserRole: UserRoleOrgAdmin,
        }),
      ).toBe(true);
    });

    it("returns false when editing a different user", () => {
      expect(
        showCannotChangeOwnRoleHint({
          adminUserId: "admin-1",
          editedUserId: "user-2",
          adminUserRole: UserRoleOrgAdmin,
          editedUserRole: UserRoleUser,
        }),
      ).toBe(false);
    });

    it("returns false when editing a different user with higher role", () => {
      expect(
        showCannotChangeOwnRoleHint({
          adminUserId: "space-admin-1",
          editedUserId: "org-admin-1",
          adminUserRole: UserRoleSpaceAdmin,
          editedUserRole: UserRoleOrgAdmin,
        }),
      ).toBe(false);
    });
  });
});
