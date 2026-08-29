import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import React from "react";
import { createRoot } from "react-dom/client";
import { act } from "react";
import withPermission from "../withPermission";
import RuntimeConfig from "../RuntimeConfig";
import { Permission, PermissionLevel } from "@/types/Permission";

vi.mock("../Loading", () => ({
  default: () => React.createElement("div", { "data-testid": "loading" }),
}));

function Page() {
  return React.createElement("div", { "data-testid": "page" }, "content");
}

let container: HTMLDivElement;

function renderGuarded(permission?: string, level: number = PermissionLevel.Admin) {
  const push = vi.fn();
  const Guarded = withPermission(Page as any, permission, level) as any;
  const root = createRoot(container);
  act(() => {
    root.render(React.createElement(Guarded, { router: { push } }));
  });
  const rendered = container.querySelector('[data-testid="page"]') !== null;
  return { push, rendered, root };
}

describe("withPermission", () => {
  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
    RuntimeConfig.INFOS.permissions = {};
  });

  afterEach(() => {
    container.remove();
  });

  it("renders the page when the permission is held", () => {
    RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Admin };
    const { rendered, push } = renderGuarded(Permission.Groups, PermissionLevel.Admin);
    expect(rendered).toBe(true);
    expect(push).not.toHaveBeenCalled();
  });

  it("renders when the level held exceeds the level required", () => {
    RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Admin };
    expect(renderGuarded(Permission.Groups, PermissionLevel.Read).rendered).toBe(true);
  });

  it("withholds the page below the required level", () => {
    RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Read };
    const { rendered, push } = renderGuarded(Permission.Groups, PermissionLevel.Admin);
    expect(rendered).toBe(false);
    expect(push).toHaveBeenCalledWith("/admin/dashboard/");
  });

  it("sends a user with no administrative access to the booking UI", () => {
    // Redirecting to the dashboard would bounce them straight back, since the
    // dashboard is guarded too.
    const { push } = renderGuarded(Permission.Groups, PermissionLevel.Read);
    expect(push).toHaveBeenCalledWith("/search/");
  });

  it("accepts any administrative permission when none is named", () => {
    RuntimeConfig.INFOS.permissions = { audit_log: PermissionLevel.Read };
    expect(renderGuarded(undefined).rendered).toBe(true);
  });

  it("withholds an any-permission page from a user with none", () => {
    const { rendered, push } = renderGuarded(undefined);
    expect(rendered).toBe(false);
    expect(push).toHaveBeenCalledWith("/search/");
  });

  it("does not let one permission satisfy another", () => {
    RuntimeConfig.INFOS.permissions = { groups: PermissionLevel.Admin };
    expect(renderGuarded(Permission.Users, PermissionLevel.Read).rendered).toBe(false);
  });
});
