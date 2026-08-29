import { describe, it, expect } from "vitest";
import User from "@/types/User";
import {
  canCreateUser,
  defaultAuthMethodForNewUser,
  isAuthMethodGroupHidden,
  isAuthProviderGroupHidden,
  isAuthProviderRequired,
  isPasswordGroupHidden,
  isPasswordRequired,
  UserFormRuleInput,
} from "./UserFormRules";

const base: UserFormRuleInput = {
  hasUserId: false,
  isServiceAccount: false,
  authMethod: User.AuthMethodPassword,
  changePassword: false,
  disablePasswordLogin: false,
};

describe("defaultAuthMethodForNewUser", () => {
  it("uses password login when it is enabled", () => {
    expect(defaultAuthMethodForNewUser(false)).toBe(User.AuthMethodPassword);
  });

  it("uses the auth provider when password login is disabled", () => {
    expect(defaultAuthMethodForNewUser(true)).toBe(User.AuthMethodProvider);
  });
});

describe("canCreateUser", () => {
  it("allows creating users while password login is available", () => {
    expect(
      canCreateUser({ disablePasswordLogin: false, hasAuthProviders: false }),
    ).toBe(true);
  });

  it("allows creating users bound to an auth provider", () => {
    expect(
      canCreateUser({ disablePasswordLogin: true, hasAuthProviders: true }),
    ).toBe(true);
  });

  it("refuses when no auth method could ever work", () => {
    // Invitation links are rejected by the server while password login is
    // disabled, so such a user could never sign in.
    expect(
      canCreateUser({ disablePasswordLogin: true, hasAuthProviders: false }),
    ).toBe(false);
  });
});

describe("isAuthMethodGroupHidden", () => {
  it("hides the method choice when password login is disabled", () => {
    expect(
      isAuthMethodGroupHidden({ ...base, disablePasswordLogin: true }),
    ).toBe(true);
  });

  it("hides the method choice for service accounts", () => {
    expect(isAuthMethodGroupHidden({ ...base, isServiceAccount: true })).toBe(
      true,
    );
  });

  it("shows the method choice otherwise", () => {
    expect(isAuthMethodGroupHidden(base)).toBe(false);
  });
});

describe("password field", () => {
  it("is required for a new user with password auth", () => {
    expect(isPasswordRequired(base)).toBe(true);
    expect(isPasswordGroupHidden(base)).toBe(false);
  });

  it("is not required while hidden because password login is disabled", () => {
    // Regression: a new user created on an instance with DISABLE_PASSWORD_LOGIN=1
    // used to keep authMethod "password", leaving an empty required field inside a
    // hidden group. The browser then blocked the submit without any visible error.
    const s: UserFormRuleInput = {
      ...base,
      disablePasswordLogin: true,
      authMethod: User.AuthMethodProvider,
    };
    expect(isPasswordGroupHidden(s)).toBe(true);
    expect(isPasswordRequired(s)).toBe(false);
  });
});

describe("auth provider select", () => {
  it("is visible and required when the provider auth method is active", () => {
    const s: UserFormRuleInput = {
      ...base,
      disablePasswordLogin: true,
      authMethod: User.AuthMethodProvider,
    };
    expect(isAuthProviderGroupHidden(s)).toBe(false);
    expect(isAuthProviderRequired(s)).toBe(true);
  });

  it("is hidden and not required for another auth method", () => {
    const s: UserFormRuleInput = {
      ...base,
      authMethod: User.AuthMethodInvitation,
    };
    expect(isAuthProviderGroupHidden(s)).toBe(true);
    expect(isAuthProviderRequired(s)).toBe(false);
  });
});

describe("invariant: a hidden control is never required", () => {
  const bools = [false, true];
  const methods = [
    User.AuthMethodPassword,
    User.AuthMethodProvider,
    User.AuthMethodInvitation,
  ];
  it("holds for every combination of form state", () => {
    for (const hasUserId of bools)
      for (const isServiceAccount of bools)
        for (const changePassword of bools)
          for (const disablePasswordLogin of bools)
            for (const authMethod of methods) {
              const s: UserFormRuleInput = {
                hasUserId,
                isServiceAccount,
                authMethod,
                changePassword,
                disablePasswordLogin,
              };
              const label = JSON.stringify(s);
              if (isPasswordGroupHidden(s)) {
                expect(isPasswordRequired(s), `password ${label}`).toBe(false);
              }
              if (isAuthProviderGroupHidden(s)) {
                expect(isAuthProviderRequired(s), `provider ${label}`).toBe(
                  false,
                );
              }
            }
  });
});
