import { Entity } from "../types/Entity";
import AjaxError from "./AjaxError";
import { Mutex } from "async-mutex";
import AjaxCredentials from "./AjaxCredentials";
import AjaxConfigPersister from "./AjaxConfigPersister";
import AjaxConfigBrowserPersister from "./AjaxConfigBrowserPersister";
import JwtDecoder from "./JwtDecoder";

interface AjaxResult {
  json: any;
  status: number;
  objectId: string;
}

export default class Ajax {
  static URL: string = "";
  static PERSISTER: AjaxConfigPersister = new AjaxConfigBrowserPersister();
  private static REFRESH_URL: string = "/auth/refresh";
  private static REFRESH_TOKEN_MUTEX: Mutex = new Mutex();

  static getBackendUrl(): string {
    let url = Ajax.URL.trim();
    if (url.endsWith("/")) {
      url = url.substring(0, url.length - 1);
    }
    return url;
  }

  static async query(
    method: string,
    url: string,
    data?: any,
  ): Promise<AjaxResult> {
    url = Ajax.getBackendUrl() + url;
    return new Promise<AjaxResult>(function (resolve, reject) {
      let performRequest = async () => {
        const credentials: AjaxCredentials =
          await Ajax.PERSISTER.readCredentialsFromLocalStorage();
        const options: RequestInit = Ajax.getFetchOptions(
          method,
          credentials.accessToken,
          data,
        );
        fetch(url, options)
          .then((response) => {
            if (response.status >= 200 && response.status <= 299) {
              response
                .json()
                .then((json) => {
                  resolve(Ajax.getAjaxResult(json, response));
                })
                .catch((err) => {
                  resolve(Ajax.getAjaxResult({}, response));
                });
            } else {
              let appCode = response.headers.get("X-Error-Code");
              reject(
                new AjaxError(response.status, appCode ? parseInt(appCode) : 0),
              );
            }
          })
          .catch((err) => {
            reject(err);
          });
      };
      const refreshToken = Ajax.PERSISTER.readRefreshTokenFromLocalStorage();
      if (refreshToken) {
        Ajax.refreshAccessToken(refreshToken)
          .then(() => {
            performRequest();
          })
          .catch(() => {
            reject(new AjaxError(401, 0));
          });
      } else {
        performRequest();
      }
    });
  }

  static async refreshAccessToken(refreshToken: string): Promise<void> {
    // Acquire mutex so that refreshing the token is not refreshed concurrently
    return Ajax.REFRESH_TOKEN_MUTEX.acquire().then((release) => {
      return new Promise<void>(function (resolve, reject) {
        // Once it's our turn, check if we really need to refresh the token
        const credentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
        if (new Date().getTime() < credentials.accessTokenExpiry.getTime()) {
          // Token is still valid, nothing to do
          release();
          resolve();
          return;
        }
        // Refresh the token
        let data = {
          refreshToken: refreshToken,
        };
        let options: RequestInit = Ajax.getFetchOptions("POST", null, data);
        let url = Ajax.getBackendUrl() + Ajax.REFRESH_URL;
        const oldCredentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
        fetch(url, options)
          .then((response) => {
            if (response.status >= 200 && response.status <= 299) {
              response
                .json()
                .then((json) => {
                  let c: AjaxCredentials = {
                    accessToken: json.accessToken,
                    accessTokenExpiry: JwtDecoder.getExpiryDate(
                      json.accessToken,
                    ),
                    logoutUrl: oldCredentials.logoutUrl,
                    profilePageUrl: oldCredentials.profilePageUrl,
                  };
                  Ajax.PERSISTER.updateCredentialsLocalStorage(c);
                  Ajax.PERSISTER.persistRefreshTokenInLocalStorage(
                    json.refreshToken,
                  );
                  release();
                  resolve();
                })
                .catch((err) => {
                  release();
                  reject(new AjaxError(response.status, 0));
                });
            } else {
              // token invalid
              Ajax.PERSISTER.deleteCredentialsFromStorage();
              resolve();
              release();
            }
          })
          .catch((err) => {
            release();
            reject(err);
          });
      });
    });
  }

  static hasAccessToken(): boolean {
    const credentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
    return !!credentials.accessToken;
  }

  static getAjaxResult(json: any, response: Response): AjaxResult {
    let objectId: string = "";
    if (response.headers.get("X-Object-Id") != null) {
      objectId = String(response.headers.get("X-Object-Id"));
    }
    let res: AjaxResult = {
      json: json,
      status: response.status,
      objectId: objectId,
    };
    return res;
  }

  static getFetchOptions(
    method: string,
    accessToken?: string | null,
    data?: any,
  ): RequestInit {
    let headers = new Headers();
    if (accessToken) {
      headers.append("Authorization", "Bearer " + accessToken);
    }
    if (data && !(data instanceof File)) {
      headers.append("Content-Type", "application/json");
    }
    let options: RequestInit = {
      method: method,
      mode: "cors",
      cache: "no-cache",
      credentials: "same-origin",
      headers: headers,
      signal: AbortSignal.timeout(30000),
    };
    if (data) {
      if (data instanceof File) {
        options.body = data;
      } else {
        options.body = JSON.stringify(data);
      }
    }
    return options;
  }

  static async postData(url: string, data?: any): Promise<AjaxResult> {
    return Ajax.query("POST", url, data);
  }

  static async putData(url: string, data?: any): Promise<AjaxResult> {
    return Ajax.query("PUT", url, data);
  }

  static async head(url: string, params?: any): Promise<AjaxResult> {
    if (params) {
      let s = "";
      for (const k in params) {
        if (s !== "") {
          s += "&";
        }
        s += k + "=" + encodeURIComponent(params[k]);
      }
      url += "?" + s;
    }
    return Ajax.query("HEAD", url, null);
  }

  static async saveEntity(e: Entity, url: string): Promise<AjaxResult> {
    if (!url.endsWith("/")) {
      url += "/";
    }
    if (e.id) {
      return Ajax.putData(url + e.id, e.serialize());
    } else {
      return Ajax.postData(url, e.serialize()).then((result) => {
        e.id = result.objectId;
        return result;
      });
    }
  }

  static async get(url: string): Promise<AjaxResult> {
    return Ajax.query("GET", url);
  }

  static async delete(url: string): Promise<AjaxResult> {
    return Ajax.query("DELETE", url);
  }
}
