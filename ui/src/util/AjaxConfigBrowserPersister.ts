import AjaxConfigPersister from "./AjaxConfigPersister";
import AjaxCredentials from "./AjaxCredentials";

export default class AjaxConfigBrowserPersister implements AjaxConfigPersister {
  persistRefreshTokenInLocalStorage(refreshToken: string): void {
    try {
      window.localStorage.setItem("refreshToken", refreshToken);
    } catch (e) {}
  }

  readRefreshTokenFromLocalStorage(): string {
    let c: string = "";
    try {
      let refreshToken = window.localStorage.getItem("refreshToken");
      if (refreshToken) {
        c = refreshToken;
      }
    } catch (e) {}
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
    } catch (e) {}
  }

  readCredentialsFromLocalStorage(): AjaxCredentials {
    let c: AjaxCredentials = new AjaxCredentials();
    try {
      let accessToken = window.localStorage.getItem("accessToken");
      let accessTokenExpiry = window.localStorage.getItem("accessTokenExpiry");
      let logoutUrl = window.localStorage.getItem("logoutUrl");
      let profilePageUrl = window.localStorage.getItem("profilePageUrl");
      if (accessToken && accessTokenExpiry) {
        c = {
          accessToken: accessToken,
          accessTokenExpiry: new Date(window.parseInt(accessTokenExpiry)),
          logoutUrl: logoutUrl || "",
          profilePageUrl: profilePageUrl || "",
        };
      }
    } catch (e) {}
    return c;
  }

  deleteCredentialsFromStorage(): void {
    try {
      window.localStorage.removeItem("accessToken");
      window.localStorage.removeItem("accessTokenExpiry");
      window.localStorage.removeItem("logoutUrl");
      window.localStorage.removeItem("profilePageUrl");
      window.localStorage.removeItem("refreshToken");
    } catch (e) {}
  }
}
