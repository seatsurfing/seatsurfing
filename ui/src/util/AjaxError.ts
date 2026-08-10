export default class AjaxError extends Error {
  private _httpStatusCode: number;
  private _appErrorCode: number;
  public responseBody: string;

  constructor(
    httpStatusCode: number,
    appErrorCode: number,
    responseBody?: string,
  ) {
    super(`HTTP Status ${httpStatusCode} with app error code ${appErrorCode}`);
    Object.setPrototypeOf(this, AjaxError.prototype);
    this._httpStatusCode = httpStatusCode;
    this._appErrorCode = appErrorCode;
    this.responseBody = responseBody ?? "";
  }

  public get httpStatusCode(): number {
    return this._httpStatusCode;
  }

  public get appErrorCode(): number {
    return this._appErrorCode;
  }

  static getAppErrorCode(e: Error): number {
    if (e instanceof AjaxError) {
      return e.appErrorCode;
    }
    return 0;
  }
}
