import AjaxCredentials from "./AjaxCredentials";

export default interface AjaxConfigPersister {
  persistRefreshTokenInLocalStorage(refreshToken: string): void;
  readRefreshTokenFromLocalStorage(): string;
  updateCredentialsLocalStorage(c: AjaxCredentials): void;
  readCredentialsFromLocalStorage(): AjaxCredentials;
  deleteCredentialsFromStorage(): void;
}
