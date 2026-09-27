import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import React from "react";
import { createRoot, type Root } from "react-dom/client";
import { act } from "react";
import Ajax from "@/util/Ajax";
import PluginEmbed from "../PluginEmbed";

// Register a trivial custom element once; PluginEmbed only ever sets plain
// properties on it, so a bare HTMLElement subclass is enough.
if (!customElements.get("plugin-test-el")) {
  customElements.define("plugin-test-el", class extends HTMLElement {});
}

function renderIntoDiv(element: React.ReactElement): {
  div: HTMLDivElement;
  root: Root;
} {
  const div = document.createElement("div");
  document.body.appendChild(div);
  const root = createRoot(div);
  act(() => {
    root.render(element);
  });
  return { div, root };
}

// The plugin module is loaded via a <script type="module"> tag appended to
// document.head; jsdom never actually fetches it, so the test resolves it by
// hand, the same way the browser would via the script's onload.
function resolvePendingScript(src: string) {
  const script = document.head.querySelector(
    `script[src="${src}"]`,
  ) as HTMLScriptElement | null;
  if (!script) {
    throw new Error(`No pending script for ${src}`);
  }
  act(() => {
    script.onload?.(new Event("load"));
  });
}

async function flush() {
  await act(async () => {
    await Promise.resolve();
    await Promise.resolve();
  });
}

describe("PluginEmbed", () => {
  const credentials = () => ({
    accessToken: "token-1",
    accessTokenExpiry: new Date(Date.now() + 100000),
    logoutUrl: "",
    profilePageUrl: "",
  });

  beforeEach(() => {
    document.head.innerHTML = "";
    Ajax.PERSISTER = {
      persistRefreshTokenInLocalStorage: vi.fn(),
      readRefreshTokenFromLocalStorage: vi.fn(() => "refresh-1"),
      updateCredentialsLocalStorage: vi.fn(),
      readCredentialsFromLocalStorage: vi.fn(() => credentials()),
      deleteCredentialsFromStorage: vi.fn(),
    } as any;
    Ajax.onUnauthorized = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    Ajax.onUnauthorized = null;
  });

  it("sets the access token when the host broadcasts a refresh", async () => {
    const src = "/plugin-a.js";
    const { div } = renderIntoDiv(
      <PluginEmbed id="test-a" src={src} tagName="plugin-test-el" />,
    );
    resolvePendingScript(src);
    await flush();

    const el = div.querySelector("plugin-test-el") as any;
    expect(el.accessToken).toBe("token-1");

    act(() => {
      window.dispatchEvent(
        new CustomEvent(Ajax.ACCESS_TOKEN_REFRESHED_EVENT, {
          detail: { accessToken: "token-2" },
        }),
      );
    });

    expect(el.accessToken).toBe("token-2");
  });

  it("stops pushing refreshed tokens after unmount", async () => {
    const src = "/plugin-b.js";
    const { div, root } = renderIntoDiv(
      <PluginEmbed id="test-b" src={src} tagName="plugin-test-el" />,
    );
    resolvePendingScript(src);
    await flush();

    const el = div.querySelector("plugin-test-el") as any;
    act(() => {
      root.unmount();
    });

    // Should not throw even though the element is detached.
    act(() => {
      window.dispatchEvent(
        new CustomEvent(Ajax.ACCESS_TOKEN_REFRESHED_EVENT, {
          detail: { accessToken: "token-2" },
        }),
      );
    });
    expect(el.accessToken).toBe("token-1");
  });

  it("refreshes the token when the element reports it expired", async () => {
    const src = "/plugin-c.js";
    const refreshedCredentials = { ...credentials(), accessToken: "token-3" };
    const readCredentials = Ajax.PERSISTER
      .readCredentialsFromLocalStorage as ReturnType<typeof vi.fn>;
    readCredentials
      .mockReturnValueOnce(credentials())
      .mockReturnValue(refreshedCredentials);
    const refreshSpy = vi
      .spyOn(Ajax, "refreshAccessToken")
      .mockResolvedValue(undefined);

    const { div } = renderIntoDiv(
      <PluginEmbed id="test-c" src={src} tagName="plugin-test-el" />,
    );
    resolvePendingScript(src);
    await flush();

    const el = div.querySelector("plugin-test-el") as any;
    expect(el.accessToken).toBe("token-1");

    await act(async () => {
      el.dispatchEvent(new CustomEvent("plugin-token-expired"));
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(refreshSpy).toHaveBeenCalledWith("refresh-1");
    expect(el.accessToken).toBe("token-3");
    expect(Ajax.onUnauthorized).not.toHaveBeenCalled();
  });

  it("falls back to session-expired handling when the refresh fails", async () => {
    const src = "/plugin-d.js";
    vi.spyOn(Ajax, "refreshAccessToken").mockRejectedValue(
      new Error("refresh token invalid"),
    );

    const { div } = renderIntoDiv(
      <PluginEmbed id="test-d" src={src} tagName="plugin-test-el" />,
    );
    resolvePendingScript(src);
    await flush();

    const el = div.querySelector("plugin-test-el") as any;

    await act(async () => {
      el.dispatchEvent(new CustomEvent("plugin-token-expired"));
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(Ajax.onUnauthorized).toHaveBeenCalled();
  });

  it("falls back to session-expired handling when there is no refresh token", async () => {
    const src = "/plugin-e.js";
    (
      Ajax.PERSISTER.readRefreshTokenFromLocalStorage as ReturnType<
        typeof vi.fn
      >
    ).mockReturnValue("");
    const refreshSpy = vi.spyOn(Ajax, "refreshAccessToken");

    const { div } = renderIntoDiv(
      <PluginEmbed id="test-e" src={src} tagName="plugin-test-el" />,
    );
    resolvePendingScript(src);
    await flush();

    const el = div.querySelector("plugin-test-el") as any;

    await act(async () => {
      el.dispatchEvent(new CustomEvent("plugin-token-expired"));
      await Promise.resolve();
    });

    expect(refreshSpy).not.toHaveBeenCalled();
    expect(Ajax.onUnauthorized).toHaveBeenCalled();
  });

  it("forwards plugin-data-changed to onDataChanged", async () => {
    const src = "/plugin-f.js";
    const onDataChanged = vi.fn();
    const { div } = renderIntoDiv(
      <PluginEmbed
        id="test-f"
        src={src}
        tagName="plugin-test-el"
        onDataChanged={onDataChanged}
      />,
    );
    resolvePendingScript(src);
    await flush();

    const el = div.querySelector("plugin-test-el") as any;
    act(() => {
      el.dispatchEvent(new CustomEvent("plugin-data-changed"));
    });

    expect(onDataChanged).toHaveBeenCalledTimes(1);
  });
});
