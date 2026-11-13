import AjaxCredentials from "./AjaxCredentials";

export default interface AjaxConfigPersister {
  persistRefreshTokenInLocalStorage(refreshToken: string): void;
  readRefreshTokenFromLocalStorage(): string;
  updateCredentialsSessionStorage(c: AjaxCredentials): void;
  readCredentialsFromSessionStorage(): AjaxCredentials;
  deleteCredentialsFromStorage(): void;
}
