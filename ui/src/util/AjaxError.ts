export default class AjaxError extends Error {
  private _httpStatusCode: number;
  private _appErrorCode: number;
  public responseBody: string;
  public handledGlobally: boolean;

  constructor(
    httpStatusCode: number,
    appErrorCode: number,
    responseBody?: string,
    handledGlobally?: boolean,
  ) {
    super(`HTTP Status ${httpStatusCode} with app error code ${appErrorCode}`);
    Object.setPrototypeOf(this, AjaxError.prototype);
    this._httpStatusCode = httpStatusCode;
    this._appErrorCode = appErrorCode;
    this.responseBody = responseBody ?? "";
    this.handledGlobally = handledGlobally ?? false;
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
