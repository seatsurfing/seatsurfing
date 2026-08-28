import { describe, it, expect, vi, beforeEach } from "vitest";
import React from "react";
import { createRoot } from "react-dom/client";
import { act } from "react";

/**
 * Regression test for creating a user on an instance started with
 * DISABLE_PASSWORD_LOGIN=1 (password login off, Keycloak/OIDC only).
 *
 * The page used to keep the "password" auth method for new users, which left an
 * empty *required* password input inside a *hidden* form group. Browsers abort
 * such a submit and cannot show the validation message, so "Save" did nothing at
 * all: no request, no error, no feedback.
 */

vi.mock("next-export-i18n", () => ({
  useTranslation: () => ({ t: (key: string) => key }),
  useSelectedLanguage: () => ({ lang: "en" }),
}));

const { routerPush } = vi.hoisted(() => ({ routerPush: vi.fn() }));
vi.mock("next/router", () => ({
  useRouter: () => ({
    isReady: true,
    query: { id: "add" },
    push: routerPush,
  }),
  default: { push: routerPush },
}));

vi.mock("@/components/FullLayout", () => ({
  default: ({ children, buttons }: any) =>
    React.createElement("div", null, buttons, children),
}));

vi.mock("@/components/RuntimeConfig", () => ({
  default: {
    INFOS: {
      userId: "admin-1",
      disablePasswordLogin: true,
      pluginWelcomeScreens: [],
      pluginMenuItems: [],
    },
  },
}));

vi.mock("@/types/User", async () => {
  const actual: any = await vi.importActual("@/types/User");
  const User = actual.default;
  User.getCount = vi.fn().mockResolvedValue(3);
  User.getSelf = vi
    .fn()
    .mockResolvedValue({ id: "admin-1", role: User.UserRoleOrgAdmin });
  User.get = vi.fn();
  User.getApiTokenStatus = vi.fn().mockResolvedValue(false);
  return { ...actual, default: User };
});

vi.mock("@/types/Settings", async () => {
  const actual: any = await vi.importActual("@/types/Settings");
  const OrgSettings = actual.default;
  OrgSettings.getOne = vi.fn().mockResolvedValue("1"); // no user limit
  return { ...actual, default: OrgSettings };
});

vi.mock("@/types/AuthProvider", async () => {
  const actual: any = await vi.importActual("@/types/AuthProvider");
  const AuthProvider = actual.default;
  AuthProvider.list = vi
    .fn()
    .mockResolvedValue([{ id: "kc-1", name: "Keycloak" }]);
  return { ...actual, default: AuthProvider };
});

import EditUser from "../[id]";

const flush = async () => {
  await act(async () => {
    await new Promise((r) => setTimeout(r, 0));
  });
};

const renderPage = async () => {
  const container = document.createElement("div");
  document.body.appendChild(container);
  const root = createRoot(container);
  await act(async () => {
    root.render(React.createElement(EditUser as any));
  });
  await flush();
  return container;
};

const isVisible = (el: Element | null) => {
  if (!el) return false;
  let node: Element | null = el;
  while (node) {
    if (node.hasAttribute("hidden")) return false;
    node = node.parentElement;
  }
  return true;
};

describe("add user with password login disabled", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    routerPush.mockClear();
  });

  it("does not leave a required password field inside a hidden group", async () => {
    const container = await renderPage();
    const password = container.querySelector<HTMLInputElement>("#password");
    expect(password).not.toBeNull();
    expect(isVisible(password)).toBe(false);
    expect(password!.required).toBe(false);
  });

  it("shows the auth provider selection instead", async () => {
    const container = await renderPage();
    const provider =
      container.querySelector<HTMLSelectElement>("#authProvider");
    expect(isVisible(provider)).toBe(true);
    expect(provider!.required).toBe(true);
    // single provider is preselected, so the admin has nothing left to fill in
    expect(provider!.value).toBe("kc-1");
  });

  it("can be submitted after filling in the visible fields", async () => {
    const container = await renderPage();
    const form = container.querySelector<HTMLFormElement>("#form");
    const setValue = (selector: string, value: string) => {
      const input = container.querySelector<HTMLInputElement>(selector)!;
      const setter = Object.getOwnPropertyDescriptor(
        window.HTMLInputElement.prototype,
        "value",
      )!.set!;
      setter.call(input, value);
      input.dispatchEvent(new Event("input", { bubbles: true }));
    };
    await act(async () => {
      setValue("#email", "j.cermann@example.org");
      setValue("#firstname", "Jessica");
      setValue("#lastname", "Cermann");
    });
    await flush();
    expect(form!.checkValidity()).toBe(true);
  });
});
