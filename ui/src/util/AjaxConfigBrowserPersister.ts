import AjaxConfigPersister from "./AjaxConfigPersister";
import AjaxCredentials from "./AjaxCredentials";

export default class AjaxConfigBrowserPersister implements AjaxConfigPersister {
  persistRefreshTokenInLocalStorage(refreshToken: string): void {
    try {
      window.localStorage.setItem("refreshToken", refreshToken);
      window.sessionStorage.removeItem("refreshToken");
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

  updateCredentialsSessionStorage(c: AjaxCredentials): void {
    try {
      window.sessionStorage.setItem("accessToken", c.accessToken);
      window.sessionStorage.setItem(
        "accessTokenExpiry",
        c.accessTokenExpiry.getTime().toString(),
      );
      window.sessionStorage.setItem("logoutUrl", c.logoutUrl);
      window.sessionStorage.setItem("profilePageUrl", c.profilePageUrl);
    } catch (e) {
      console.error("Failed to update credentials in sessionStorage:", e);
    }
  }

  readCredentialsFromSessionStorage(): AjaxCredentials {
    let c: AjaxCredentials = new AjaxCredentials();
    try {
      const accessToken = window.sessionStorage.getItem("accessToken");
      const accessTokenExpiry =
        window.sessionStorage.getItem("accessTokenExpiry");
      const logoutUrl = window.sessionStorage.getItem("logoutUrl");
      const profilePageUrl = window.sessionStorage.getItem("profilePageUrl");
      if (accessToken && accessTokenExpiry) {
        c = {
          accessToken: accessToken,
          accessTokenExpiry: new Date(window.parseInt(accessTokenExpiry)),
          logoutUrl: logoutUrl || "",
          profilePageUrl: profilePageUrl || "",
        };
      }
    } catch (e) {
      console.error("Failed to read credentials from sessionStorage:", e);
    }
    return c;
  }

  deleteCredentialsFromStorage(): void {
    try {
      window.sessionStorage.removeItem("accessToken");
      window.sessionStorage.removeItem("accessTokenExpiry");
      window.sessionStorage.removeItem("logoutUrl");
      window.sessionStorage.removeItem("profilePageUrl");
      window.localStorage.removeItem("refreshToken");
    } catch (e) {
      console.error("Failed to delete credentials from Storage:", e);
    }
  }
}
