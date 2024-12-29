import AjaxConfigPersister from "./AjaxConfigPersister";
import AjaxCredentials from "./AjaxCredentials";

export default class AjaxConfigBrowserPersister implements AjaxConfigPersister {
    async persistRefreshTokenInLocalStorage(c: AjaxCredentials): Promise<void> {
        return new Promise<void>(function (resolve, reject) {
            try {
                window.localStorage.setItem("refreshToken", c.refreshToken);
            } catch (e) {
            }
            resolve();
        });
    }

    async readRefreshTokenFromLocalStorage(): Promise<AjaxCredentials> {
        return new Promise<AjaxCredentials>(function (resolve, reject) {
            let c: AjaxCredentials = new AjaxCredentials();
            try {
                let refreshToken = window.localStorage.getItem("refreshToken");
                if (refreshToken) {
                    c.refreshToken = refreshToken;
                }
            } catch (e) {
            }
            resolve(c);
        });
    }

    async updateCredentialsSessionStorage(c: AjaxCredentials): Promise<void> {
        return new Promise<void>(function (resolve, reject) {
            try {
                window.sessionStorage.setItem("accessToken", c.accessToken);
                window.sessionStorage.setItem("refreshToken", c.refreshToken);
                window.sessionStorage.setItem("accessTokenExpiry", c.accessTokenExpiry.getTime().toString());
                window.sessionStorage.setItem("logoutUrl", c.logoutUrl);
            } catch (e) {
            }
            resolve();
        });
    }

    async readCredentialsFromSessionStorage(): Promise<AjaxCredentials> {
        return new Promise<AjaxCredentials>(function (resolve, reject) {
            let c: AjaxCredentials = new AjaxCredentials();
            try {
                let accessToken = window.sessionStorage.getItem("accessToken");
                let refreshToken = window.sessionStorage.getItem("refreshToken");
                let accessTokenExpiry = window.sessionStorage.getItem("accessTokenExpiry");
                let logoutUrl = window.sessionStorage.getItem("logoutUrl");
                if (accessToken && refreshToken && accessTokenExpiry) {
                    c = {
                        accessToken: accessToken,
                        refreshToken: refreshToken,
                        accessTokenExpiry: new Date(window.parseInt(accessTokenExpiry)),
                        logoutUrl: logoutUrl || ''
                    };
                }
            } catch (e) {
            }
            resolve(c);
        });
    }

    async deleteCredentialsFromSessionStorage(): Promise<void> {
        return new Promise<void>(function (resolve, reject) {
            try {
                window.sessionStorage.removeItem("accessToken");
                window.sessionStorage.removeItem("refreshToken");
                window.sessionStorage.removeItem("accessTokenExpiry");
                window.localStorage.removeItem("refreshToken");
                window.localStorage.removeItem("logoutUrl");
            } catch (e) {
            }
            resolve();
        });
    }
}