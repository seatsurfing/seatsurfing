import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import React from "react";
import { createRoot, Root } from "react-dom/client";
import { act } from "react";

let mockQuery: Record<string, any> = {};
const mockPush = vi.fn();

vi.mock("next/router", () => ({
  useRouter: () => ({ isReady: true, query: mockQuery, push: mockPush }),
}));
vi.mock("next/link", () => ({
  default: ({ children }: any) => React.createElement("a", null, children),
}));
vi.mock("next-export-i18n", () => ({
  useTranslation: () => ({ t: (key: string) => key }),
  useSelectedLanguage: () => ({ lang: "en" }),
}));
vi.mock("@/components/SeatsurfingAppLogo", () => ({ default: () => null }));

import Ajax from "@/util/Ajax";
import ConfirmSignup from "@/pages/admin/confirm/[id]";
import LoginSuccess from "@/pages/login/success/[id]";

const EVIL = "..\\..\\..\\user\\totp\\disable";
const UUID = "4271113d-5eaa-4ef1-8a4e-610faf27d5d0";

describe("dynamic route ids never reach a request path unvalidated", () => {
  let container: HTMLDivElement;
  let root: Root;

  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
  });

  afterEach(() => {
    act(() => root.unmount());
    container.remove();
    vi.restoreAllMocks();
  });

  async function mount(Page: any) {
    await act(async () => {
      root.render(React.createElement(Page));
    });
  }

  it("admin/confirm sends no request for a non-UUID id", async () => {
    mockQuery = { id: EVIL };
    const post = vi.spyOn(Ajax, "postData");
    await mount(ConfirmSignup);
    expect(post).not.toHaveBeenCalled();
  });

  it("admin/confirm posts to the confirm endpoint for a UUID", async () => {
    mockQuery = { id: UUID };
    const post = vi
      .spyOn(Ajax, "postData")
      .mockResolvedValue({ status: 200, json: { domain: "x.example" } } as any);
    await mount(ConfirmSignup);
    expect(post).toHaveBeenCalledTimes(1);
    expect(post.mock.calls[0][0]).toBe(
      "/cloud-features/signup/confirm/" + UUID,
    );
  });

  it("login/success redirects to the failure page without a request for a non-UUID id", async () => {
    mockQuery = { id: EVIL };
    mockPush.mockClear();
    const get = vi.spyOn(Ajax, "get");
    await mount(LoginSuccess);
    expect(get).not.toHaveBeenCalled();
    // must not leave the user on an endless spinner
    expect(mockPush).toHaveBeenCalledWith("/login/failed/");
  });
});
