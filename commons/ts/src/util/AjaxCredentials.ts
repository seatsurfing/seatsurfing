export default class AjaxCredentials {
    accessToken: string = "";
    refreshToken: string = "";
    accessTokenExpiry: Date = new Date(0);
    logoutUrl: string = "";
}
