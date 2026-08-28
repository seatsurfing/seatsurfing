import { describe, it, expect } from "vitest";
import User from "@/types/User";
import {
  defaultAuthMethodForNewUser,
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
    expect(
      defaultAuthMethodForNewUser({
        disablePasswordLogin: false,
        hasAuthProviders: true,
      }),
    ).toBe(User.AuthMethodPassword);
  });

  it("uses the auth provider when password login is disabled", () => {
    expect(
      defaultAuthMethodForNewUser({
        disablePasswordLogin: true,
        hasAuthProviders: true,
      }),
    ).toBe(User.AuthMethodProvider);
  });

  it("falls back to invitation when password login is disabled and no provider exists", () => {
    expect(
      defaultAuthMethodForNewUser({
        disablePasswordLogin: true,
        hasAuthProviders: false,
      }),
    ).toBe(User.AuthMethodInvitation);
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
