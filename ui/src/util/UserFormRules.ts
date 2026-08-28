import User from "@/types/User";

/**
 * Visibility and validation rules for the "Edit User" admin form.
 *
 * These live outside the component so the relationship between a field's
 * visibility and its "required" flag can be tested directly. A required form
 * control inside a hidden group makes the browser abort the submit without
 * being able to show its validation message, which leaves the user with a
 * form that silently does nothing.
 */
export interface UserFormRuleInput {
  /** false while creating a new user */
  hasUserId: boolean;
  isServiceAccount: boolean;
  /** one of User.AuthMethod* */
  authMethod: string;
  changePassword: boolean;
  disablePasswordLogin: boolean;
}

/**
 * Auth method a newly created user starts with. With password login disabled
 * instance-wide, a new user can only authenticate through an auth provider.
 */
export function defaultAuthMethodForNewUser(opts: {
  disablePasswordLogin: boolean;
  hasAuthProviders: boolean;
}): string {
  if (!opts.disablePasswordLogin) {
    return User.AuthMethodPassword;
  }
  return opts.hasAuthProviders
    ? User.AuthMethodProvider
    : User.AuthMethodInvitation;
}

export function isAuthMethodGroupHidden(s: UserFormRuleInput): boolean {
  return s.isServiceAccount || s.disablePasswordLogin;
}

export function isAuthProviderGroupHidden(s: UserFormRuleInput): boolean {
  return s.isServiceAccount || s.authMethod !== User.AuthMethodProvider;
}

export function isAuthProviderRequired(s: UserFormRuleInput): boolean {
  return !isAuthProviderGroupHidden(s);
}

export function isPasswordGroupHidden(s: UserFormRuleInput): boolean {
  if (s.isServiceAccount) {
    return false;
  }
  return s.disablePasswordLogin || s.authMethod !== User.AuthMethodPassword;
}

export function isPasswordRequired(s: UserFormRuleInput): boolean {
  if (isPasswordGroupHidden(s)) {
    return false;
  }
  if (s.isServiceAccount) {
    return true;
  }
  if (!s.hasUserId) {
    return s.authMethod === User.AuthMethodPassword;
  }
  return s.changePassword && s.authMethod === User.AuthMethodPassword;
}
