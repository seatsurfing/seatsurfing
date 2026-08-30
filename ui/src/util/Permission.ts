import User from "@/types/User";

export default class Permission {
  static isServiceAccount(role: number): boolean {
    return (
      role === User.UserRoleServiceAccountRO ||
      role === User.UserRoleServiceAccountRW
    );
  }

  static canManageRole(adminRole: number, role: number): boolean {
    // Service account roles (21/22) sit numerically above OrgAdmin (20) but are
    // less privileged, so they can't be compared via the plain role ranking.
    if (Permission.isServiceAccount(role)) {
      return adminRole >= User.UserRoleOrgAdmin;
    }
    return adminRole >= role;
  }
}
