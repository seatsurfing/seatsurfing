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
  static onUnauthorized: (() => void) | null = null;
  static onServerError: (() => void) | null = null;
  static onNotFound: (() => void) | null = null;

  private static REFRESH_URL: string = "/auth/refresh";
  private static REFRESH_TOKEN_MUTEX: Mutex = new Mutex();
  private static HEADER_X_OBJECT_ID: string = "X-Object-Id";
  private static HEADER_X_ERROR_CODE: string = "X-Error-Code";

  static getBackendUrl(): string {
    let url = Ajax.URL.trim();
    if (url.endsWith("/")) {
      url = url.substring(0, url.length - 1);
    }
    return url;
  }

  private static async query(
    method: string,
    url: string,
    data?: any,
    haltOnGlobalError: boolean = true,
  ): Promise<AjaxResult> {
    // refresh access token (if required)
    const refreshToken = Ajax.PERSISTER.readRefreshTokenFromLocalStorage();
    if (refreshToken) {
      try {
        await Ajax.refreshAccessToken(refreshToken);
      } catch {
        if (haltOnGlobalError) {
          Ajax.handleGlobalError(401);
          return new Promise<AjaxResult>(() => {});
        }
        throw new AjaxError(401, 0);
      }
    }

    const credentials: AjaxCredentials =
      Ajax.PERSISTER.readCredentialsFromLocalStorage();
    const options: RequestInit = Ajax.getFetchOptions(
      method,
      credentials.accessToken,
      data,
    );

    url = Ajax.getBackendUrl() + url;

    let response: Response;
    try {
      response = await fetch(url, options);
    } catch {
      // Network error (backend unreachable, timeout, etc.)
      if (haltOnGlobalError) {
        Ajax.handleGlobalError(0);
        return new Promise<AjaxResult>(() => {});
      }
      throw new AjaxError(0, 0);
    }

    if (response.status >= 200 && response.status <= 299) {
      try {
        const json = await response.json();
        return Ajax.getAjaxResult(json, response);
      } catch {
        return Ajax.getAjaxResult({}, response);
      }
    } else {
      let appCode = 0;
      try {
        appCode = parseInt(
          response.headers.get(this.HEADER_X_ERROR_CODE) ?? "0",
        );
      } catch {}

      // global error handlers if appCode is not defined
      if (appCode === 0 && haltOnGlobalError) {
        Ajax.handleGlobalError(response.status);
        return new Promise<AjaxResult>(() => {});
      }

      let body: string | undefined;
      try {
        body = await response.text();
      } catch {}
      throw new AjaxError(response.status, appCode, body);
    }
  }

  private static handleGlobalError(httpStatus: number): void {
    if (httpStatus === 401) {
      // Only trigger the modal if credentials still exist (first 401).
      // Clear them immediately so subsequent in-flight 401s are silent.
      const hadCredentials = Ajax.hasAccessToken();
      Ajax.PERSISTER.deleteCredentialsFromStorage();
      if (hadCredentials) {
        Ajax.onUnauthorized?.();
      }
    } else if (httpStatus === 404) {
      Ajax.onNotFound?.();
    } else {
      // 500, network errors (0), and any other unexpected status
      Ajax.onServerError?.();
    }
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
        const data = {
          refreshToken: refreshToken,
        };
        const options: RequestInit = Ajax.getFetchOptions("POST", null, data);
        const url = Ajax.getBackendUrl() + Ajax.REFRESH_URL;
        const oldCredentials = Ajax.PERSISTER.readCredentialsFromLocalStorage();
        fetch(url, options)
          .then((response) => {
            if (response.status >= 200 && response.status <= 299) {
              response
                .json()
                .then((json) => {
                  const c: AjaxCredentials = {
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
                .catch(() => {
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

  private static getAjaxResult(json: any, response: Response): AjaxResult {
    const objectId =
      response.headers.get(this.HEADER_X_OBJECT_ID) != null
        ? String(response.headers.get(this.HEADER_X_OBJECT_ID))
        : "";
    const res: AjaxResult = {
      json,
      status: response.status,
      objectId,
    };
    return res;
  }

  static getFetchOptions(
    method: string,
    accessToken?: string | null,
    data?: any,
  ): RequestInit {
    const headers = new Headers();
    if (accessToken) {
      headers.append("Authorization", "Bearer " + accessToken);
    }
    if (data && !(data instanceof File)) {
      headers.append("Content-Type", "application/json");
    }
    const options: RequestInit = {
      method,
      mode: "cors",
      cache: "no-cache",
      credentials: "same-origin",
      headers,
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

  static async postData(
    url: string,
    data?: any,
    haltOnGlobalError: boolean = true,
  ): Promise<AjaxResult> {
    return Ajax.query("POST", url, data, haltOnGlobalError);
  }

  static async putData(
    url: string,
    data?: any,
    haltOnGlobalError: boolean = true,
  ): Promise<AjaxResult> {
    return Ajax.query("PUT", url, data, haltOnGlobalError);
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

  static async get(
    url: string,
    haltOnGlobalError: boolean = true,
  ): Promise<AjaxResult> {
    return Ajax.query("GET", url, undefined, haltOnGlobalError);
  }

  static async delete(
    url: string,
    haltOnGlobalError: boolean = true,
  ): Promise<AjaxResult> {
    return Ajax.query("DELETE", url, undefined, haltOnGlobalError);
  }
}
