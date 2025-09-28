export default class AjaxError extends Error {
  private _httpStatusCode: number;
  private _appErrorCode: number;

  constructor(httpStatusCode: number, appErrorCode: number, m?: string) {
    if (!m) {
      m =
        "HTTP Status " +
        httpStatusCode +
        " with app error code " +
        appErrorCode;
    }
    super(m);
    Object.setPrototypeOf(this, AjaxError.prototype);
    this._httpStatusCode = httpStatusCode;
    this._appErrorCode = appErrorCode;
  }

  public get httpStatusCode(): number {
    return this._httpStatusCode;
  }

  public get appErrorCode(): number {
    return this._appErrorCode;
  }
}
