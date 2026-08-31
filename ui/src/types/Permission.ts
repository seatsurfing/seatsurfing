/**
 * Mirrors server/api/permissions.go. The two lists must be kept in step: the
 * Go side is the source of truth, and the catalogue endpoint reports the
 * levels each permission actually offers.
 */
export const PermissionLevel = {
  None: 0,
  Read: 10,
  Write: 20,
  Admin: 30,
} as const;

export type PermissionLevelValue =
  (typeof PermissionLevel)[keyof typeof PermissionLevel];

export const Permission = {
  Areas: "areas",
  SpaceAttributes: "space_attributes",
  Bookings: "bookings",
  Approvals: "approvals",
  Analytics: "analytics",
  PresenceReport: "presence_report",
  Users: "users",
  Groups: "groups",
  Roles: "roles",
  ServiceAccounts: "service_accounts",
  AuthProviders: "auth_providers",
  OrgSettings: "org_settings",
  AuditLog: "audit_log",
} as const;

export type PermissionKey =
  (typeof Permission)[keyof typeof Permission] | string;

/** A user's resolved access, keyed by permission name. */
export type PermissionMap = { [key: string]: number };

/**
 * Translation key for a permission's display name. Plugin permissions carry a
 * "plugin." prefix and have no built-in translation, so their key is used
 * as-is by the caller when no translation exists.
 */
export function permissionLabelKey(key: PermissionKey): string {
  return "permission." + key;
}

/** Translation key for an access level's display name. */
export function permissionLevelLabelKey(level: number): string {
  switch (level) {
    case PermissionLevel.Read:
      return "permissionLevelRead";
    case PermissionLevel.Write:
      return "permissionLevelWrite";
    case PermissionLevel.Admin:
      return "permissionLevelAdmin";
    default:
      return "permissionLevelNone";
  }
}
