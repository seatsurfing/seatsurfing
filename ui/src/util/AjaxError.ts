export default class AjaxError extends Error {
  private _httpStatusCode: number;
  private _appErrorCode: number;
  public responseBody: string;

  constructor(
    httpStatusCode: number,
    appErrorCode: number,
    m?: string,
    responseBody?: string,
  ) {
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
    this.responseBody = responseBody ?? "";
  }

  public get httpStatusCode(): number {
    return this._httpStatusCode;
  }

  public get appErrorCode(): number {
    return this._appErrorCode;
  }
}
