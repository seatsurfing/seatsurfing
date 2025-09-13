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

  updateCredentialsSessionStorage(c: AjaxCredentials): void {
    try {
      window.sessionStorage.setItem("accessToken", c.accessToken);
      window.sessionStorage.setItem(
        "accessTokenExpiry",
        c.accessTokenExpiry.getTime().toString(),
      );
      window.sessionStorage.setItem("logoutUrl", c.logoutUrl);
    } catch (e) {}
  }

  readCredentialsFromSessionStorage(): AjaxCredentials {
    let c: AjaxCredentials = new AjaxCredentials();
    try {
      let accessToken = window.sessionStorage.getItem("accessToken");
      let accessTokenExpiry =
        window.sessionStorage.getItem("accessTokenExpiry");
      let logoutUrl = window.sessionStorage.getItem("logoutUrl");
      if (accessToken && accessTokenExpiry) {
        c = {
          accessToken: accessToken,
          accessTokenExpiry: new Date(window.parseInt(accessTokenExpiry)),
          logoutUrl: logoutUrl || "",
        };
      }
    } catch (e) {}
    return c;
  }

  deleteCredentialsFromStorage(): void {
    try {
      window.sessionStorage.removeItem("accessToken");
      window.sessionStorage.removeItem("accessTokenExpiry");
      window.sessionStorage.removeItem("logoutUrl");
      window.localStorage.removeItem("refreshToken");
    } catch (e) {}
  }
}
