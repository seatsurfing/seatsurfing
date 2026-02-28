import AjaxConfigPersister from "./AjaxConfigPersister";
import AjaxCredentials from "./AjaxCredentials";

export default class AjaxConfigBrowserPersister implements AjaxConfigPersister {
  persistRefreshTokenInLocalStorage(refreshToken: string): void {
    try {
      window.localStorage.setItem("refreshToken", refreshToken);
    } catch (e) {
      console.error("Failed to persist refresh token in localStorage:", e);
    }
  }

  readRefreshTokenFromLocalStorage(): string {
    let c: string = "";
    try {
      const refreshToken = window.localStorage.getItem("refreshToken");
      if (refreshToken) {
        c = refreshToken;
      }
    } catch (e) {
      console.error("Failed to read refresh token from localStorage:", e);
    }
    return c;
  }

  updateCredentialsLocalStorage(c: AjaxCredentials): void {
    try {
      window.localStorage.setItem("accessToken", c.accessToken);
      window.localStorage.setItem(
        "accessTokenExpiry",
        c.accessTokenExpiry.getTime().toString(),
      );
      window.localStorage.setItem("logoutUrl", c.logoutUrl);
      window.localStorage.setItem("profilePageUrl", c.profilePageUrl);
    } catch (e) {
      console.error("Failed to update credentials in localStorage:", e);
    }
  }

  readCredentialsFromLocalStorage(): AjaxCredentials {
    let c: AjaxCredentials = new AjaxCredentials();
    try {
      const accessToken = window.localStorage.getItem("accessToken");
      const accessTokenExpiry =
        window.localStorage.getItem("accessTokenExpiry");
      const logoutUrl = window.localStorage.getItem("logoutUrl");
      const profilePageUrl = window.localStorage.getItem("profilePageUrl");
      if (accessToken && accessTokenExpiry) {
        c = {
          accessToken: accessToken,
          accessTokenExpiry: new Date(window.parseInt(accessTokenExpiry)),
          logoutUrl: logoutUrl || "",
          profilePageUrl: profilePageUrl || "",
        };
      }
    } catch (e) {
      console.error("Failed to read credentials from localStorage:", e);
    }
    return c;
  }

  deleteCredentialsFromStorage(): void {
    try {
      window.localStorage.removeItem("accessToken");
      window.localStorage.removeItem("accessTokenExpiry");
      window.localStorage.removeItem("logoutUrl");
      window.localStorage.removeItem("profilePageUrl");
      window.localStorage.removeItem("refreshToken");
    } catch (e) {
      console.error("Failed to delete credentials from localStorage:", e);
    }
  }
}
