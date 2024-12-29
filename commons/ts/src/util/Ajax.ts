import { Entity } from "../types/Entity";
import AjaxError from "./AjaxError";
import { Mutex } from 'async-mutex';
import AjaxCredentials from "./AjaxCredentials";
import AjaxConfigPersister from "./AjaxConfigPersister";
import AjaxConfigBrowserPersister from "./AjaxConfigBrowserPersister";

interface AjaxResult {
  json: any
  status: number
  objectId: string
}

export default class Ajax {
  static URL: string = "";
  static CREDENTIALS: AjaxCredentials = new AjaxCredentials();
  static ACCESS_TOKEN_EXPIRY_OFFSET: number = 60 * 5 * 1000; // 5 minutes
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

  static async query(method: string, url: string, data?: any): Promise<AjaxResult> {
    console.log("AJAX request for: " + url);
    url = Ajax.getBackendUrl() + url;
    return new Promise<AjaxResult>(function (resolve, reject) {
      let performRequest = () => {
        let options: RequestInit = Ajax.getFetchOptions(method, Ajax.CREDENTIALS.accessToken, data);
        fetch(url, options).then((response) => {
          if (response.status >= 200 && response.status <= 299) {
            response.json().then(json => {
              resolve(Ajax.getAjaxResult(json, response));
            }).catch(err => {
              resolve(Ajax.getAjaxResult({}, response));
            });
          } else {
            let appCode = response.headers.get("X-Error-Code");
            reject(new AjaxError(response.status, appCode ? parseInt(appCode) : 0));
          }
        }).catch(err => {
          reject(err);
        });
      };
      if (Ajax.CREDENTIALS.refreshToken) {
        Ajax.refreshAccessToken(Ajax.CREDENTIALS.refreshToken).then(cred => {
          performRequest();
        }).catch(err => {
          reject(new AjaxError(401, 0));
        });
      } else {
        performRequest();
      }
    });
  }

  static async refreshAccessToken(refreshToken: string): Promise<AjaxCredentials> {
    // Acquire mutex so that refreshing the token is not refreshed concurrently
    return Ajax.REFRESH_TOKEN_MUTEX.acquire().then(release => {
      return new Promise<AjaxCredentials>(function (resolve, reject) {
        // Once it's our turn, check if we really need to refresh the token
        if (new Date().getTime() < Ajax.CREDENTIALS.accessTokenExpiry.getTime()) {
          // Token is still valid, nothing to do
          release();
          resolve(Ajax.CREDENTIALS);
        } else {
          // Refresh the token
          console.log("Performing access token refresh with refresh token: " + refreshToken);
          let data = {
            "refreshToken": refreshToken
          };
          let options: RequestInit = Ajax.getFetchOptions("POST", null, data);
          let url = Ajax.getBackendUrl() + Ajax.REFRESH_URL;
          fetch(url, options).then((response) => {
            if (response.status >= 200 && response.status <= 299) {
              response.json().then(json => {
                let c: AjaxCredentials = {
                  accessToken: json.accessToken,
                  refreshToken: json.refreshToken,
                  accessTokenExpiry: new Date(new Date().getTime() + Ajax.ACCESS_TOKEN_EXPIRY_OFFSET),
                  logoutUrl: json.logoutUrl
                };
                Ajax.CREDENTIALS = c;
                Ajax.PERSISTER.updateCredentialsSessionStorage(c).then(() => {
                  Ajax.PERSISTER.readRefreshTokenFromLocalStorage().then(refreshTokenConfig => {
                    if (refreshTokenConfig.refreshToken) {
                      Ajax.PERSISTER.persistRefreshTokenInLocalStorage(c).then(() => {
                        release();
                        resolve(c);
                      });
                    } else {
                      release();
                      resolve(c);
                    }
                  });
                });
              }).catch(err => {
                release();
                reject(new AjaxError(response.status, 0));
              });
            } else {
              // token invalid
              Ajax.CREDENTIALS = new AjaxCredentials();
              Ajax.PERSISTER.deleteCredentialsFromSessionStorage().then(() => {
                resolve(Ajax.CREDENTIALS);
                release();
              });
            }
          }).catch(err => {
            release();
            reject(err);
          });
        }
      });
    });
  }

  static getAjaxResult(json: any, response: Response): AjaxResult {
    let objectId: string = "";
    if (response.headers.get("X-Object-Id") != null) {
      objectId = String(response.headers.get("X-Object-Id"));
    }
    let res: AjaxResult = {
      json: json,
      status: response.status,
      objectId: objectId
    };
    return res;
  }

  static getFetchOptions(method: string, accessToken?: string | null, data?: any): RequestInit {
    let headers = new Headers();
    if (accessToken) {
      headers.append("Authorization", "Bearer " + accessToken)
    }
    if (data && !(data instanceof File)) {
      headers.append("Content-Type", "application/json");
    }
    let options: RequestInit = {
      method: method,
      mode: "cors",
      cache: "no-cache",
      credentials: "same-origin",
      headers: headers
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
      return Ajax.postData(url, e.serialize()).then(result => {
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
