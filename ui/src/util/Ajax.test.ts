import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import Ajax from "./Ajax";

// A JWT with just enough shape for JwtDecoder.getExpiryDate to read the exp
// claim; signature/header content don't matter for these tests.
function makeJwt(expSecondsFromNow: number): string {
  const header = window.btoa(JSON.stringify({ alg: "none" }));
  const payload = window.btoa(
    JSON.stringify({
      exp: Math.floor(Date.now() / 1000) + expSecondsFromNow,
    }),
  );
  return `${header}.${payload}.sig`;
}

describe("Ajax.refreshAccessToken", () => {
  beforeEach(() => {
    window.localStorage.clear();
    window.sessionStorage.clear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    window.localStorage.clear();
    window.sessionStorage.clear();
  });

  it("skips the network call when the cached token is still valid", async () => {
    window.sessionStorage.setItem("accessToken", makeJwt(3600));
    window.sessionStorage.setItem(
      "accessTokenExpiry",
      (Date.now() + 3600 * 1000).toString(),
    );
    const fetchSpy = vi.spyOn(global, "fetch");

    await Ajax.refreshAccessToken("some-refresh-token");

    expect(fetchSpy).not.toHaveBeenCalled();
  });

  it("forces the network call even when the cached token still looks valid", async () => {
    window.sessionStorage.setItem("accessToken", makeJwt(3600));
    window.sessionStorage.setItem(
      "accessTokenExpiry",
      (Date.now() + 3600 * 1000).toString(),
    );
    const newToken = makeJwt(900);
    const fetchSpy = vi.spyOn(global, "fetch").mockResolvedValue({
      status: 200,
      json: async () => ({
        accessToken: newToken,
        refreshToken: "new-refresh-token",
      }),
    } as Response);

    // force=true, e.g. because the backend just rejected the cached token
    // (PluginEmbed's "plugin-token-expired" handling) - the stale local
    // expiry must not short-circuit the refresh in that case.
    await Ajax.refreshAccessToken("some-refresh-token", true);

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    expect(window.sessionStorage.getItem("accessToken")).toBe(newToken);
    expect(window.localStorage.getItem("refreshToken")).toBe(
      "new-refresh-token",
    );
    expect(window.sessionStorage.getItem("refreshToken")).toBeNull();
  });

  it("does not refresh when no access token exists", async () => {
    const fetchSpy = vi.spyOn(global, "fetch");

    await Ajax.refreshAccessToken("some-refresh-token");

    expect(fetchSpy).not.toHaveBeenCalled();
  });
});
