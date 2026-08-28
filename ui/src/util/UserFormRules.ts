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
 * instance-wide, an auth provider is the only method that can work: invitation
 * links are rejected by the server in that case.
 */
export function defaultAuthMethodForNewUser(
  disablePasswordLogin: boolean,
): string {
  return disablePasswordLogin
    ? User.AuthMethodProvider
    : User.AuthMethodPassword;
}

/**
 * With password login disabled and no auth provider configured, there is no way
 * for a new user to ever sign in, so the form must not pretend otherwise.
 */
export function canCreateUser(opts: {
  disablePasswordLogin: boolean;
  hasAuthProviders: boolean;
}): boolean {
  return !opts.disablePasswordLogin || opts.hasAuthProviders;
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
